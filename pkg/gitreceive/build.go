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

	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/gitreceive/git"
	"github.com/deis/builder/pkg/storage"
	"github.com/deis/builder/pkg/sys"
	"github.com/deis/pkg/log"
	"gopkg.in/yaml.v2"

	storagedriver "github.com/deis/builder/pkg/storage/driver"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

// repoCmd returns exec.Command(first, others...) with its current working directory repoDir
func repoCmd(repoDir, first string, others ...string) *exec.Cmd {
	cmd := exec.Command(first, others...)
	cmd.Dir = repoDir
	return cmd
}

// run prints the command it will execute to the debug log, then runs it and returns the result
// of run
func run(cmd *exec.Cmd) error {
	cmdStr := strings.Join(cmd.Args, " ")
	if cmd.Dir != "" {
		log.Debug("running [%s] in directory %s", cmdStr, cmd.Dir)
	} else {
		log.Debug("running [%s]", cmdStr)
	}
	return cmd.Run()
}

func build(conf *Config, storageDriver storagedriver.StorageDriver, kubeClient *client.Client, fs sys.FS, env sys.Env, builderKey, rawGitSha string) error {
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

	slugBuilderInfo := storage.NewSlugBuilderInfo(slugName)

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

	appTgzdata, err := ioutil.ReadFile(absAppTgz)
	if err != nil {
		return fmt.Errorf("error while reading file %s: (%s)", appTgz, err)
	}

	//log.Debug("Uploading tar to %s/%s/%s", s3Client.Endpoint.FullURL(), conf.Bucket, slugBuilderInfo.TarKey())

	if err := storageDriver.PutContent(slugBuilderInfo.TarKey(), appTgzdata); err != nil {
		return fmt.Errorf("uploading %s to %s (%v)", absAppTgz, slugBuilderInfo.TarKey(), err)
	}

	creds := true

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
			slugBuilderInfo.TarKey(),
			slugName,
			conf.StorageRegion,
			conf.DockerBuilderImage,
		)
	} else {
		buildPodName = slugBuilderPodName(appName, gitSha.Short())
		pod = slugbuilderPod(
			conf.Debug,
			creds,
			buildPodName,
			conf.PodNamespace,
			appConf.Values,
			slugBuilderInfo.TarKey(),
			slugBuilderInfo.PushKey(),
			buildPackURL,
			conf.SlugBuilderImage,
			conf.StorageType,
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

	log.Debug(
		"Waiting for the %s/%s pod to end. Checking every %s for %s",
		newPod.Namespace,
		newPod.Name,
		conf.BuilderPodTickDuration(),
		conf.BuilderPodWaitDuration(),
	)
	// check the state and exit code of the build pod.
	// if the code is not 0 return error
	if err := waitForPodEnd(kubeClient, newPod.Namespace, newPod.Name, conf.BuilderPodTickDuration(), conf.BuilderPodWaitDuration()); err != nil {
		return fmt.Errorf("error getting builder pod status (%s)", err)
	}
	log.Debug("Done")
	log.Debug("Checking for builder pod exit code")
	buildPod, err := kubeClient.Pods(newPod.Namespace).Get(newPod.Name)
	if err != nil {
		return fmt.Errorf("error getting builder pod status (%s)", err)
	}

	for _, containerStatus := range buildPod.Status.ContainerStatuses {
		state := containerStatus.State.Terminated
		if state.ExitCode != 0 {
			return fmt.Errorf("Build pod exited with code %d, stopping build.", state.ExitCode)
		}
	}
	log.Debug("Done")

	log.Debug(
		"Polling the S3 server every %s for %s for the resultant slug at %s",
		conf.ObjectStorageTickDuration(),
		conf.ObjectStorageWaitDuration(),
		slugBuilderInfo.AbsoluteSlugObjectKey(),
	)
	// poll the s3 server to ensure the slug exists
	if !usingDockerfile {
		if err := storage.WaitForObject(
			storageDriver,
			slugBuilderInfo.AbsoluteSlugObjectKey(),
			conf.ObjectStorageTickDuration(),
			conf.ObjectStorageWaitDuration(),
		); err != nil {
			return fmt.Errorf("Timed out waiting for object in storage, aborting build (%s)", err)
		}
	}

	procType := pkg.ProcessType{}
	if bType == buildTypeProcfile {
		if procType, err = getProcFile(storageDriver, tmpDir, slugBuilderInfo.AbsoluteProcfileKey()); err != nil {
			return err
		}
	}

	log.Info("Build complete.")
	log.Info("Launching app.")
	log.Info("Launching...")

	buildHook := createBuildHook(slugBuilderInfo, gitSha, conf.Username, appName, procType, usingDockerfile)
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

func getProcFile(storageDriver storagedriver.StorageDriver, dirName, procfileKey string) (pkg.ProcessType, error) {
	procType := pkg.ProcessType{}
	if _, err := os.Stat(fmt.Sprintf("%s/Procfile", dirName)); err == nil {
		rawProcFile, err := ioutil.ReadFile(fmt.Sprintf("%s/Procfile", dirName))
		if err != nil {
			return nil, fmt.Errorf("error in reading %s/Procfile (%s)", dirName, err)
		}
		if err := yaml.Unmarshal(rawProcFile, &procType); err != nil {
			return nil, fmt.Errorf("procfile %s/ProcFile is malformed (%s)", dirName, err)
		}
		return procType, nil
	}
	log.Debug("Procfile not present. Getting it from the buildpack")
	rawProcFile, err := storageDriver.GetContent(procfileKey)
	if err != nil {
		return nil, fmt.Errorf("error in reading %s (%s)", procfileKey, err)
	}
	if err := yaml.Unmarshal(rawProcFile, &procType); err != nil {
		return nil, fmt.Errorf("procfile %s is malformed (%s)", procfileKey, err)
	}
	return procType, nil
}
