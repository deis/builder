package gitreceive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/gitreceive/git"
	"github.com/deis/builder/pkg/gitreceive/storage"
	"github.com/deis/pkg/log"
	"gopkg.in/yaml.v2"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/wait"
)

// repoCmd returns exec.Command(first, others...) with its current working directory repoDir
func repoCmd(repoDir, first string, others ...string) *exec.Cmd {
	cmd := exec.Command(first, others...)
	cmd.Dir = repoDir
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

func build(conf *Config, s3Client *s3.S3, kubeClient *client.Client, builderKey, rawGitSha string) error {
	repo := conf.Repository
	gitSha, err := git.NewSha(rawGitSha)
	if err != nil {
		return err
	}

	appName := conf.App()

	repoDir := filepath.Join(conf.GitHome, repo)
	buildDir := filepath.Join(repoDir, "build")

	slugName := fmt.Sprintf("%s:git-%s", appName, gitSha.Short())
	if err := os.MkdirAll(buildDir, os.ModeDir); err != nil {
		return fmt.Errorf("making the build directory %s (%s)", buildDir, err)
	}

	tmpDir, err := ioutil.TempDir(buildDir, "tmp")
	if err != nil {
		return fmt.Errorf("unable to create tmpdir %s (%s)", buildDir, err)
	}

	slugBuilderInfo := storage.NewSlugBuilderInfo(s3Client.Endpoint, conf.Bucket, appName, slugName, gitSha)

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
	gitArchiveCmd := repoCmd(repoDir, "git", "archive", "--format=tar.gz", fmt.Sprintf("--output=%s", appTgz), gitSha.Short())
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

	procType := pkg.ProcessType{}
	if bType == buildTypeProcfile {
		rawProcFile, err := ioutil.ReadFile(fmt.Sprintf("%s/Procfile", tmpDir))
		if err != nil {
			return fmt.Errorf("reading %s/Procfile", tmpDir)
		}
		if err := yaml.Unmarshal(rawProcFile, &procType); err != nil {
			return fmt.Errorf("procfile %s/ProcFile is malformed (%s)", tmpDir, err)
		}
	}

	if err := storage.CreateBucket(s3Client, conf.Bucket); err != nil {
		log.Warn("create bucket error: %+v", err)
	}

	appTgzReader, err := os.Open(absAppTgz)
	if err != nil {
		return fmt.Errorf("opening %s for read (%s)", appTgz, err)
	}

	log.Debug("Uploading tar to %s/%s/%s", s3Client.Endpoint, conf.Bucket, slugBuilderInfo.TarKey())
	if err := storage.UploadObject(s3Client, conf.Bucket, slugBuilderInfo.TarKey(), appTgzReader); err != nil {
		return fmt.Errorf("uploading %s to %s/%s (%v)", absAppTgz, conf.Bucket, slugBuilderInfo.TarKey(), err)
	}

	creds := storage.CredsOK()

	var pod *api.Pod
	var buildPodName string
	if usingDockerfile {
		buildPodName = dockerBuilderPodName(appName, gitSha.Short())
		pod = dockerBuilderPod(
			conf.Debug,
			creds,
			buildPodName,
			conf.PodNamespace,
			appConf.Values,
			slugBuilderInfo.TarURL(),
			slugName,
			conf.StorageRegion,
		)
	} else {
		buildPodName = slugBuilderPodName(appName, gitSha.Short())
		pod = slugbuilderPod(
			conf.Debug,
			creds,
			buildPodName,
			conf.PodNamespace,
			appConf.Values,
			slugBuilderInfo.TarURL(),
			slugBuilderInfo.PushURL(),
			buildPackURL,
		)
	}

	log.Info("Starting build... but first, coffee!")
	log.Debug("Starting pod %s", buildPodName)
	json, err := prettyPrintJSON(pod)
	if err == nil {
		log.Debug("Pod spec: %v", json)
	} else {
		log.Debug("Error creating json representaion of pod spec: %v", err)
	}

	podsInterface := kubeClient.Pods(conf.PodNamespace)

	newPod, err := podsInterface.Create(pod)
	if err != nil {
		return fmt.Errorf("creating builder pod (%s)", err)
	}

	if err := waitForPod(kubeClient, newPod.Namespace, newPod.Name, conf.BuilderPodTickDuration(), conf.BuilderPodWaitDuration()); err != nil {
		return fmt.Errorf("watching events for builder pod startup (%s)", err)
	}

	req := kubeClient.Get().Namespace(newPod.Namespace).Name(newPod.Name).Resource("pods").SubResource("log").VersionedParams(
		&api.PodLogOptions{
			Follow: true,
		}, api.Scheme)

	rc, err := req.Stream()
	if err != nil {
		return fmt.Errorf("attempting to stream logs (%s)", err)
	}
	defer rc.Close()

	size, err := io.Copy(os.Stdout, rc)
	if err != nil {
		return fmt.Errorf("fetching builder logs (%s)", err)
	}
	log.Debug("size of streamed logs %v", size)

	// check the state and exit code of the build pod.
	// if the code is not 0 return error
	if err := waitForPodEnd(kubeClient, newPod.Namespace, newPod.Name, conf.BuilderPodTickEndDuration(), conf.BuilderPodWaitDuration()); err != nil {
		return fmt.Errorf("error getting builder pod status (%s)", err)
	}
	buildPod, err := kubeClient.Pods(newPod.Namespace).Get(newPod.Name)
	if err != nil {
		return fmt.Errorf("error getting builder pod status (%s)", err)
	}

	for _, containerStatus := range buildPod.Status.ContainerStatuses {
		state := containerStatus.State.Terminated
		if state.ExitCode != 0 {
			return fmt.Errorf("Stopping build.")
		}
	}

	// poll the s3 server to ensure the slug exists
	err = wait.PollImmediate(conf.ObjectStorageTickDuration(), conf.ObjectStorageWaitDuration(), func() (bool, error) {
		exists, err := storage.ObjectExists(s3Client, conf.Bucket, slugBuilderInfo.PushKey())
		if err != nil {
			return false, fmt.Errorf("Checking if object %s/%s exists (%s)", conf.Bucket, slugBuilderInfo.PushKey(), err)
		}
		return exists, nil
	})

	if err != nil {
		return fmt.Errorf("Timed out waiting for object in storage. Aborting build...")
	}
	log.Info("Build complete.")
	log.Info("Launching app.")
	log.Info("Launching...")

	buildHook := &pkg.BuildHook{
		Sha:         gitSha.Short(),
		ReceiveUser: conf.Username,
		ReceiveRepo: appName,
		Image:       appName,
		Procfile:    procType,
	}
	if !usingDockerfile {
		buildHook.Dockerfile = ""
		// need this to tell the controller what URL to give the slug runner
		buildHook.Image = slugBuilderInfo.PushURL() + "/slug.tgz"
	} else {
		buildHook.Dockerfile = "true"
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

func prettyPrintJSON(data interface{}) (string, error) {
	output := &bytes.Buffer{}
	if err := json.NewEncoder(output).Encode(data); err != nil {
		return "", err
	}
	formatted := &bytes.Buffer{}
	if err := json.Indent(formatted, output.Bytes(), "", "  "); err != nil {
		return "", err
	}
	return string(formatted.Bytes()), nil
}
