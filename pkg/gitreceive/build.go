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

	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/gitreceive/log"
	"github.com/deis/builder/pkg/gitreceive/storage"
	"github.com/mitchellh/goamz/s3"
	"github.com/pborman/uuid"
	"gopkg.in/yaml.v2"
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

func build(conf *Config, s3Client *s3.S3, builderKey, gitSha string) error {
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

	tarObjKey := fmt.Sprintf("home/%s/tar", slugName)
	tarURL := fmt.Sprintf("%s/git/%s", s3Client.S3Endpoint, tarObjKey)

	// this is where workflow tells slugrunner to download the slug from, so we have to tell slugbuilder to upload it to here
	pushObjKey := fmt.Sprintf("home/%s/push", fmt.Sprintf("%s:git-%s", appName, gitSha))
	pushURL := fmt.Sprintf("%s/%s", s3Client.S3Endpoint, pushObjKey)

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
	absAppTgz := fmt.Sprintf("%s/%s", repoDir, appTgz)

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

	var srcManifest string
	if creds == nil {
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

	bucketName := "git"
	if err := storage.CreateBucket(s3Client, bucketName); err != nil {
		log.Warn("create bucket error: %+v", err)
	}

	appTgzReader, err := os.Open(absAppTgz)
	if err != nil {
		return fmt.Errorf("opening %s for read (%s)", appTgz, err)
	}
	if err := storage.UploadObject(s3Client, bucketName, tarObjKey, appTgzReader); err != nil {
		return fmt.Errorf("uploading %s to %s/%s (%s)", absAppTgz, bucketName, tarObjKey, err)
	}

	log.Info("Starting build... but first, coffee!")
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
	// TODO: time out looking
	for {
		exists, err := storage.ObjectExists(s3Client, bucketName, pushObjKey)
		if err != nil {
			return fmt.Errorf("Checking if object %s/%s exists (%s)", bucketName, pushObjKey, err)
		}
		if exists {
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
