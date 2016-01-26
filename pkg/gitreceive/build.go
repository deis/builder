package gitreceive

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/gitreceive/log"
	"github.com/pborman/uuid"
	"gopkg.in/yaml.v2"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

const (
	// this constant represents the length of a shortened git sha - 8 characters long
	shortShaIdx = 8
)

type errGitShaTooShort struct {
	sha string
}

func (e errGitShaTooShort) Error() string {
	return fmt.Sprintf("git sha %s was too short", e.sha)
}

// repoCmd returns exec.Command(first, others...) with its current working directory repoDir
func repoCmd(repoDir, first string, others ...string) *exec.Cmd {
	cmd := exec.Command(first, others...)
	cmd.Dir = repoDir
	return cmd
}

// mcCmd returns a command to execute the 'mc' binary, so that it reads config from configDir.
// the command outputs its stderr to os.Stderr
func mcCmd(configDir string, args ...string) *exec.Cmd {
	cmd := exec.Command("mc", "-C", configDir, "--quiet")
	cmd.Args = append(cmd.Args, args...)
	cmd.Stderr = os.Stderr
	return cmd
}

// run prints the command it will execute to the debug log, then runs it and returns the result of run
func run(cmd *exec.Cmd) error {
	cmdStr := strings.Join(cmd.Args, " ")
	if cmd.Dir != "" {
		log.Debug("running [%s] in directory %s", cmdStr, cmd.Dir)
	} else {
		log.Debug("running [%s]", cmdStr)
	}
	return cmd.Run()
}

func build(conf *Config, kubeClient *client.Client, builderKey, gitSha string) error {
	storage, err := getStorageConfig()
	if err != nil {
		return err
	}
	creds, err := getStorageCreds()
	if err == errMissingKey || err == errMissingSecret {
		return err
	}

	repo := conf.Repository
	if len(gitSha) <= shortShaIdx {
		return errGitShaTooShort{sha: gitSha}
	}
	shortSha := gitSha[0:8]
	appName := conf.App()

	repoDir := filepath.Join(conf.GitHome, repo)
	buildDir := filepath.Join(repoDir, "build")

	slugName := fmt.Sprintf("%s:git-%s", appName, shortSha)
	imageName := strings.Replace(slugName, ":", "-", -1)
	if err := os.MkdirAll(buildDir, os.ModeDir); err != nil {
		return fmt.Errorf("making the build directory %s (%s)", buildDir, err)
	}
	tmpDir := os.TempDir()

	tarURL := fmt.Sprintf("%s://%s:%s/git/home/%s/tar", storage.schema(), storage.host(), storage.port(), slugName)

	// this is where workflow tells slugrunner to download the slug from, so we have to tell slugbuilder to upload it to here
	pushURL := fmt.Sprintf("%s://%s:%s/git/home/%s/push", storage.schema(), storage.host(), storage.port(), fmt.Sprintf("%s:git-%s", appName, gitSha))

	// Get the application config from the controller, so we can check for a custom buildpack URL
	appConf, err := getAppConfig(conf, builderKey, conf.Username, appName)
	if err != nil {
		return fmt.Errorf("getting app config for %s (%s)", appName, err)
	}
	log.Debug("got the following config back for app %s: %+v", appName, *appConf)
	var buildPackURL string
	if buildPackURLInterface, ok := appConf.Values["BUILDPACK_URL"]; ok {
		if bpStr, ok := buildPackURLInterface.(string); ok {
			log.Debug("found custom buildpack URL %s", bpStr)
			buildPackURL = bpStr
		}
	}

	// build a tarball from the new objects
	appTgz := fmt.Sprintf("%s.tar.gz", appName)
	gitArchiveCmd := repoCmd(repoDir, "git", "archive", "--format=tar.gz", fmt.Sprintf("--output=%s", appTgz), gitSha)
	gitArchiveCmd.Stdout = os.Stdout
	gitArchiveCmd.Stderr = os.Stderr
	if err := run(gitArchiveCmd); err != nil {
		return fmt.Errorf("running %s (%s)", strings.Join(gitArchiveCmd.Args, " "), err)
	}

	// untar the archive into the temp dir
	tarCmd := repoCmd(repoDir, "tar", "-xzf", appTgz, "-C", fmt.Sprintf("%s/", tmpDir))
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr
	if err := run(tarCmd); err != nil {
		return fmt.Errorf("running %s (%s)", strings.Join(tarCmd.Args, " "), err)
	}

	bType := getBuildTypeForDir(tmpDir)
	usingDockerfile := bType == buildTypeDockerfile

	var procType pkg.ProcessType
	if bType == buildTypeProcfile {
		rawProcFile, err := ioutil.ReadFile(fmt.Sprintf("%s/Procfile", tmpDir))
		if err != nil {
			return fmt.Errorf("reading %s/Procfile", tmpDir)
		}
		if err := yaml.Unmarshal(rawProcFile, &procType); err != nil {
			return fmt.Errorf("procfile %s/ProcFile is malformed (%s)", tmpDir, err)
		}
	}

	configDir := "/var/minio-conf"
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating minio config file (%s)", err)
	}

	configCmd := mcCmd(configDir, "config", "host", "add", fmt.Sprintf("%s://%s:%s", storage.schema(), storage.host(), storage.port()), creds.key, creds.secret)
	if err := run(configCmd); err != nil {
		return fmt.Errorf("configuring the minio client (%s)", err)
	}

	makeBucketCmd := mcCmd(configDir, "mb", fmt.Sprintf("%s://%s:%s/git", storage.schema(), storage.host(), storage.port()))
	// Don't look for errors here. Buckets may already exist
	// https://github.com/deis/builder/issues/80 will eliminate this distaste
	run(makeBucketCmd)

	cpCmd := mcCmd(configDir, "cp", appTgz, tarURL)
	cpCmd.Dir = repoDir
	if err := run(cpCmd); err != nil {
		return fmt.Errorf("copying %s to %s (%s)", appTgz, tarURL, err)
	}

	var pod *api.Pod
	var buildPodName string
	if usingDockerfile {
		buildPodName = dockerBuilderPodName(appName, shortSha)
		pod = dockerBuilderPod(
			conf.Debug,
			creds != nil,
			buildPodName,
			conf.PodNamespace,
			"deis",
			"2.0.0-beta",
			tarURL,
			imageName,
		)
	} else {
		buildPodName = slugBuilderPodName(appName, shortSha)
		pod = slugbuilderPod(
			conf.Debug,
			creds != nil,
			buildPodName,
			conf.PodNamespace,
			"deis",
			"2.0.0-beta",
			tarURL,
			pushURL,
			buildPackURL,
		)
	}
	log.Info("Starting build... but first, coffee!")
	log.Debug("Starting pod %s", buildPodName)

	podsInterface := kubeClient.Pods(conf.PodNamespace)

	newPod, err := podsInterface.Create(pod)
	if err != nil {
		return fmt.Errorf("creating builder pod (%s)", err)
	}

	watcher, err := podsInterface.Watch(labels.Nothing(), fields.OneTermEqualSelector("name", pod.Name), v1Version)
	if err != nil {
		return fmt.Errorf("watching events for builder pod startup")
	}

	ch := watcher.ResultChan()
	for evt := range ch {
		if evt.Type == watch.Added {
			watcher.Stop()
			break
		} else if evt.Type == watch.Error {
			watcher.Stop()
			return fmt.Errorf("builder pod failed to launch with ERROR")
		}
	}

	req := podsInterface.GetLogs(newPod.Name, &api.PodLogOptions{
		Follow:   true,
		Previous: true,
	})

	rc, err := req.Stream()
	if err != nil {
		return fmt.Errorf("attempting to stream logs (%s)", err)
	}

	// TODO: use a bufio Scanner to stream the logs

	// poll the s3 server to ensure the slug exists
	for {
		// for now, assume the error indicates that the slug wasn't there, nothing else
		// TODO: implement https://github.com/deis/builder/issues/80, which will clean this up siginficantly
		lsCmd := mcCmd(configDir, "ls", pushURL)
		if err := run(lsCmd); err == nil {
			break
		}
	}

	log.Info("Build complete.")
	log.Info("Launching app.")
	log.Info("Launching...")

	buildHook := &pkg.BuildHook{
		Sha:         gitSha,
		ReceiveUser: conf.Username,
		ReceiveRepo: appName,
		Image:       appName,
		Procfile:    procType,
	}
	if !usingDockerfile {
		buildHook.Dockerfile = ""
	} else {
		buildHook.Dockerfile = "true"
		buildHook.Image = imageName
	}
	buildHookResp, err := publishRelease(conf, builderKey, buildHook)
	if err != nil {
		return fmt.Errorf("publishing release (%s)", err)
	}
	release, ok := buildHookResp.Release["version"]
	if !ok {
		return fmt.Errorf("No release returned from Deis controller")
	}

	log.Info("Done, %s:v%d deployed to Deis\n", appName, release)
	log.Info("Use 'deis open' to view this application in your browser\n")
	log.Info("To learn more, use 'deis help' or visit http://deis.io\n")

	gcCmd := repoCmd(repoDir, "git", "gc")
	if err := run(gcCmd); err != nil {
		return fmt.Errorf("cleaning up the repository with %s (%s)", strings.Join(gcCmd.Args, " "), err)
	}

	return nil
}
