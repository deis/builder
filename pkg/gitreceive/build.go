package gitreceive

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/log"
	"gopkg.in/yaml.v2"
)

const (
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

func kGetCmd(podNS, podName string) *exec.Cmd {
	return exec.Command("kubectl", fmt.Sprintf("--namespace=%s", podNS), "get", "pods", "-o", "yaml", podName)
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

func build(conf *Config, builderKey, gitSha string) error {
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

	// Ensure that the app config can be gotten from workflow. We don't do anything with this information
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
	gitArchiveCmd := repoCmd(repoDir, "git", "archive", "--format=tar.gz", fmt.Sprintf("--output=%s.tar.gz", appName), gitSha)
	gitArchiveCmd.Stdout = os.Stdout
	gitArchiveCmd.Stderr = os.Stderr
	if err := run(gitArchiveCmd); err != nil {
		return fmt.Errorf("running %s (%s)", strings.Join(gitArchiveCmd.Args, " "), err)
	}

	// untar the archive into the temp dir
	tarCmd := repoCmd(repoDir, "tar", "-xzf", fmt.Sprintf("%s.tar.gz", appName), "-C", fmt.Sprintf("%s/", tmpDir))
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr
	if err := run(tarCmd); err != nil {
		return fmt.Errorf("running %s (%s)", strings.Join(tarCmd.Args, " "), err)
	}

	usingDockerfile := true
	rawProcFile, err := ioutil.ReadFile(fmt.Sprintf("%s/Procfile", tmpDir))
	if err == nil {
		usingDockerfile = false
	}
	var procType pkg.ProcessType
	if err := yaml.Unmarshal(rawProcFile, &procType); err != nil {
		return fmt.Errorf("procfile %s/ProcFile is malformed (%s)", tmpDir, err)
	}

	var srcManifest string
	if err == os.ErrNotExist {
		// both key and secret are missing, proceed with no credentials
		if usingDockerfile {
			srcManifest = "/etc/deis-dockerbuilder-no-creds.yaml"
		} else {
			srcManifest = "/etc/deis-slugbuilder-no-creds.yaml"
		}
	} else if err == nil {
		// both key and secret are in place, so proceed with credentials
		if usingDockerfile {
			srcManifest = "/etc/deis-dockerbuilder.yaml"
		} else {
			srcManifest = "/etc/deis-slugbuilder.yaml"
		}
	} else if err != nil {
		// unexpected error, fail
		return fmt.Errorf("unexpected error (%s)", err)
	}

	fileBytes, err := ioutil.ReadFile(srcManifest)
	if err != nil {
		return fmt.Errorf("reading kubernetes manifest %s (%s)", srcManifest, err)
	}

	finalManifestFileLocation := fmt.Sprintf("/etc/%s", slugName)
	var buildPodName string
	var finalManifest string
	uid := uuid.New()[:8]
	if usingDockerfile {
		buildPodName = fmt.Sprintf("dockerbuild-%s-%s-%s", appName, shortSha, uid)
		finalManifest = strings.Replace(string(fileBytes), "repo_name", buildPodName, -1)
		finalManifest = strings.Replace(finalManifest, "puturl", pushURL, -1)
		finalManifest = strings.Replace(finalManifest, "tar-url", tarURL, -1)
	} else {
		buildPodName = fmt.Sprintf("slugbuild-%s-%s-%s", appName, shortSha, uid)
		finalManifest = strings.Replace(string(fileBytes), "repo_name", buildPodName, -1)
		finalManifest = strings.Replace(finalManifest, "puturl", pushURL, -1)
		finalManifest = strings.Replace(finalManifest, "tar-url", tarURL, -1)
		finalManifest = strings.Replace(finalManifest, "buildurl", buildPackURL, -1)
	}

	log.Debug("writing builder manifest to %s", finalManifestFileLocation)
	if err := ioutil.WriteFile(finalManifestFileLocation, []byte(finalManifest), os.ModePerm); err != nil {
		return fmt.Errorf("writing final manifest %s (%s)", finalManifestFileLocation, err)
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

	cpCmd := mcCmd(configDir, "cp", fmt.Sprintf("%s.tar.gz", appName), tarURL)
	cpCmd.Dir = repoDir
	if err := run(cpCmd); err != nil {
		return fmt.Errorf("copying %s.tar.gz to %s (%s)", appName, tarURL, err)
	}

	log.Info("Starting build")
	log.Debug("Starting pod %s", buildPodName)
	kCreateCmd := exec.Command(
		"kubectl",
		fmt.Sprintf("--namespace=%s", conf.PodNamespace),
		"create",
		"-f",
		finalManifestFileLocation,
	)
	if log.IsDebugging {
		kCreateCmd.Stdout = os.Stdout
	}
	kCreateCmd.Stderr = os.Stderr
	if err := run(kCreateCmd); err != nil {
		return fmt.Errorf("creating builder pod (%s)", err)
	}

	// poll kubectl every 100ms to determine when the build pod is running
	// TODO: use the k8s client and watch the event stream instead (https://github.com/deis/builder/issues/65)
	for {
		cmd := kGetCmd(conf.PodNamespace, buildPodName)
		var out bytes.Buffer
		cmd.Stdout = &out
		// ignore errors
		run(cmd)
		outStr := string(out.Bytes())
		if strings.Contains(outStr, "phase: Running") {
			break
		} else if strings.Contains(outStr, "phase: Failed") {
			return fmt.Errorf("build pod %s entered phase: Failed", buildPodName)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// get logs from the builder pod
	kLogsCmd := exec.Command(
		"kubectl",
		fmt.Sprintf("--namespace=%s", conf.PodNamespace),
		"logs",
		"-f",
		buildPodName,
	)
	kLogsCmd.Stdout = os.Stdout
	if err := run(kLogsCmd); err != nil {
		return fmt.Errorf("running %s to get builder logs (%s)", strings.Join(kLogsCmd.Args, " "), err)
	}

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

	log.Info("Done, %s:v%d deployed to Deis", appName, release)
	log.Info("Use 'deis open' to view this application in your browser")
	log.Info("To learn more, use 'deis help' or visit http://deis.io")

	gcCmd := repoCmd(repoDir, "git", "gc")
	if err := run(gcCmd); err != nil {
		return fmt.Errorf("cleaning up the repository with %s (%s)", strings.Join(gcCmd.Args, " "), err)
	}

	return nil
}
