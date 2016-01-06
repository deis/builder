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
	"github.com/coreos/go-etcd/etcd"
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

func build(conf *Config, etcdClient *etcd.Client) error {
	// TODO: replace etcd usage here with something else. See https://github.com/deis/builder/issues/81
	builderKey, err := getBuilderKey(etcdClient)
	if err != nil {
		return fmt.Errorf("couldn't get builder key %s (%s)", builderKey, err)
	}

	// HTTP_PREFIX="http"
	// REMOTE_STORAGE="0"
	// # if minio is in the cluster, use it. otherwise use fetcher
	// # TODO: figure out something for using S3 also
	// if [[ -n "$DEIS_MINIO_SERVICE_HOST" && -n "$DEIS_MINIO_SERVICE_PORT" ]]; then
	//   S3EP=${DEIS_MINIO_SERVICE_HOST}:${DEIS_MINIO_SERVICE_PORT}
	//   REMOTE_STORAGE="1"
	// elif [[ -n "$DEIS_OUTSIDE_STORAGE_HOST" && -n "$DEIS_OUTSIDE_STORAGE_PORT" ]]; then
	//   HTTP_PREFIX="https"
	//   S3EP=${DEIS_OUTSIDE_STORAGE_HOST}:${DEIS_OUTSIDE_STORAGE_PORT}
	//   REMOTE_STORAGE="1"
	// elif [ -z "$S3EP" ]; then
	//   S3EP=${HOST}:3000
	// fi
	//
	// TAR_URL=$HTTP_PREFIX://$S3EP/git/home/${SLUG_NAME}/tar
	// PUSH_URL=$HTTP_PREFIX://$S3EP/git/home/${SLUG_NAME}/push
	storage, err := getStorageConfig()
	if err != nil {
		return err
	}
	creds, err := getStorageCreds()
	if err == errMissingKey || err == errMissingSecret {
		return err
	}

	// #!/usr/bin/env bash
	// #
	// # builder hook called on every git receive-pack
	// # NOTE: this script must be run as root (for docker access)
	// #
	// set -eo pipefail
	//
	// ARGS=3
	// HOST=`ifconfig eth0 | grep 'inet addr:' | cut -d: -f2 | awk '{ print $1}'`
	// indent() {
	//     echo "       $@"
	// }
	//
	// puts-step() {
	//     echo "-----> $@"
	// }
	//
	// puts-step-sameline() {
	//     echo -n "-----> $@"
	// }
	//
	// puts-warn() {
	//     echo " !     $@"
	// }
	//
	// usage() {
	//     echo "Usage: $0 <user> <repo> <sha>"
	// }
	//
	// parse-string(){
	//     # helper to avoid the single quote escape
	//     # occurred in command substitution
	//     local args=() idx=0 IFS=' ' c
	//     for c; do printf -v args[idx++] '%s ' "$c"; done
	//     printf "%s\n" "${args[*]}"
	// }
	//
	// if [ $# -ne $ARGS ]; then
	//     usage
	//     exit 1
	// fi
	//

	// USER=$1
	// REPO=$2
	// GIT_SHA=$3
	// SHORT_SHA=${GIT_SHA:0:8}
	// APP_NAME="${REPO%.*}"
	repo := conf.Repository
	gitSha := conf.SHA
	if len(gitSha) <= shortShaIdx {
		return errGitShaTooShort{sha: gitSha}
	}
	shortSha := conf.SHA[0:8]
	appName := conf.App
	//
	// cd $(dirname $0) # ensure we are in the root dir
	//
	// ROOT_DIR=$(pwd)
	// REPO_DIR="${ROOT_DIR}/${REPO}"
	// BUILD_DIR="${REPO_DIR}/build"
	// CACHE_DIR="${REPO_DIR}/cache"
	rootDir, err := os.Getwd()
	if err != nil {
		return nil
	}
	repoDir := filepath.Join(rootDir, repo)
	buildDir := filepath.Join(repoDir, "build")
	// cacheDir := filepath.Join(repoDir, "cache")
	//
	// # define image names
	// SLUG_NAME="$APP_NAME:git-$SHORT_SHA"
	// META_NAME=`echo ${SLUG_NAME}| tr ":" "-"`
	// TMP_IMAGE="$DEIS_REGISTRY_SERVICE_HOST:$DEIS_REGISTRY_SERVICE_PORT/$IMAGE_NAME"
	// # create app directories
	// mkdir -p $BUILD_DIR $CACHE_DIR
	// # create temporary directory inside the build dir for this push
	// TMP_DIR=$(mktemp -d -p $BUILD_DIR)
	slugName := fmt.Sprintf("%s:git-%s", appName, shortSha)
	// metaName := strings.Replace(slugName, ":", "-", -1)
	tmpImage := fmt.Sprintf("%s:%s/%s", conf.RegistryHost, conf.RegistryPort, conf.ImageName)
	if err := os.MkdirAll(buildDir, os.ModeDir); err != nil {
		return fmt.Errorf("making the build directory %s (%s)", buildDir, err)
	}
	tmpDir := os.TempDir()

	tarURL := fmt.Sprintf("%s://%s:%s/git/home/%s/tar", storage.schema(), storage.host(), storage.port(), slugName)
	pushURL := fmt.Sprintf("%s://%s:%s/git/hom/%s/push", storage.schema(), storage.host(), storage.port(), slugName)

	//
	// cd $REPO_DIR
	// # use Procfile if provided, otherwise try default process types from ./release
	// git archive --format=tar.gz ${GIT_SHA} > ${APP_NAME}.tar.gz
	cmd := exec.Command("git", "archive", "--format=tar.gz", fmt.Sprintf("%s > %s.tar.gz", gitSha, appName))
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running %s", strings.Join(cmd.Args, " "))
	}
	// tar -xzf ${APP_NAME}.tar.gz -C $TMP_DIR/
	tarCmd := exec.Command("tar", "-xzf", fmt.Sprintf("%s.tar.gz", appName), "-C", fmt.Sprintf("%s/", tmpDir))
	tarCmd.Dir = repoDir
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("running %s", strings.Join(cmd.Args, " "))
	}

	// USING_DOCKERFILE=true
	// if [ -f $TMP_DIR/Procfile ]; then
	//     PROCFILE=$(cat $TMP_DIR/Procfile | yaml2json-procfile)
	//     USING_DOCKERFILE=false
	// else
	//     PROCFILE="{}"
	// fi

	usingDockerfile := true
	rawProcFile, err := ioutil.ReadFile(fmt.Sprintf("%s/Procfile", tmpDir))
	if err != nil {
		usingDockerfile = false
	}
	var procType pkg.ProcessType
	if err := yaml.Unmarshal(rawProcFile, &procType); err != nil {
		return fmt.Errorf("procfile %s/ProcFile is malformed (%s)", tmpDir, err)
	}

	// if [[ ! -f /var/run/secrets/object/store/access-key-id ]]; then
	//   if $USING_DOCKERFILE ; then
	//     l1=`grep -n "object-store" /etc/deis-dockerbuilder.yaml | head -n1 |cut -d ":" -f1`
	//     l2=$(($l1+3))
	//     sed "$l1,$l2 d" /etc/deis-dockerbuilder.yaml > /etc/${SLUG_NAME}.yaml.tmp
	//     l1=`grep -n "object-store" /etc/deis-dockerbuilder.yaml.tmp | head -n1 |cut -d ":" -f1`
	//     l2=$(($l1+3))
	//     sed "$l1,$l2 d" /etc/${SLUG_NAME}.yaml.tmp > /etc/${SLUG_NAME}.yaml
	//     sed -i -- "s#repo_name#$TMP_IMAGE#g" /etc/${SLUG_NAME}.yaml
	//   else
	//     head -n 21 /etc/deis-slugbuilder.yaml > /etc/${SLUG_NAME}.yaml
	//   fi
	// else
	//   if $USING_DOCKERFILE ; then
	//     cp /etc/deis-dockerbuilder.yaml /etc/${SLUG_NAME}.yaml
	//     sed -i -- "s#repo_name#$TMP_IMAGE#g" /etc/${SLUG_NAME}.yaml
	//   else
	//     cp /etc/deis-slugbuilder.yaml /etc/${SLUG_NAME}.yaml
	//   fi
	// fi

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

	// sed -i -- "s#repo_name#$META_NAME#g" /etc/${SLUG_NAME}.yaml
	// sed -i -- "s#puturl#$PUSH_URL#g" /etc/${SLUG_NAME}.yaml
	// sed -i -- "s#tar-url#$TAR_URL#g" /etc/${SLUG_NAME}.yaml
	finalManifestFileName := fmt.Sprintf("/etc/%s", slugName)
	var buildPodName string
	var finalManifest string
	if usingDockerfile {
		buildPodName = fmt.Sprintf("%s-%s", tmpImage, uuid.New())
		finalManifest = strings.Replace(string(fileBytes), "repo_name", buildPodName, -1)
		finalManifest = strings.Replace(finalManifest, "puturl", pushURL, -1)
		finalManifest = strings.Replace(finalManifest, "tar-url", tarURL, -1)
	} else {
		buildPodName = fmt.Sprintf("%s-%s", slugName, uuid.New())
		finalManifest = strings.Replace(string(fileBytes), "repo_name", buildPodName, -1)
		finalManifest = strings.Replace(finalManifest, "puturl", pushURL, -1)
		finalManifest = strings.Replace(finalManifest, "tar-url", tarURL, -1)
	}

	if err := ioutil.WriteFile(finalManifestFileName, []byte(finalManifest), os.ModePerm); err != nil {
		return fmt.Errorf("writing final manifest %s (%s)", finalManifestFileName, err)
	}
	//
	// git archive --format=tar.gz ${GIT_SHA} > ${APP_NAME}.tar.gz

	gitArchiveCmd := exec.Command("git", "archive", "--format=tar.gz", fmt.Sprintf("%s > %s.tar.gz", gitSha, appName))
	gitArchiveCmd.Dir = repoDir
	gitArchiveCmd.Stdout = os.Stdout
	gitArchiveCmd.Stderr = os.Stderr
	if err := gitArchiveCmd.Run(); err != nil {
		return fmt.Errorf("running %s", strings.Join(cmd.Args, " "))
	}

	//
	// ACCESS_KEY=`cat /var/run/secrets/object/store/access-key-id`
	// ACCESS_SECRET=`cat /var/run/secrets/object/store/access-secret-key`
	// # copy the self signed cert into the CA directory for alpine.
	// # note: we're not running minio with SSL at all right now, so no need for this.
	// # future SSL rollouts for in-cluster storage may not need it either if we set up an intermediate CA
	// # CERT_FILE="/var/run/secrets/object/ssl/access-cert"
	// # cp $CERT_FILE /etc/ssl/certs/deis-minio-self-signed-cert.crt
	// mkdir -p /var/minio-conf
	// CONFIG_DIR=/var/minio-conf
	// MC_PREFIX="mc -C $CONFIG_DIR --quiet"
	configDir := "/var/minio-conf"
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating minio config file (%s)", err)
	}
	baseMinioCmd := exec.Command("mc", "-C", configDir, "--quiet")
	baseMinioCmd.Stderr = os.Stderr

	// $MC_PREFIX config host add "$HTTP_PREFIX://$S3EP" $ACCESS_KEY $ACCESS_SECRET &>/dev/null
	configCmd := baseMinioCmd
	configCmd.Args = append(
		configCmd.Args,
		"config",
		"host",
		"add",
		fmt.Sprintf("%s://%s:%s", storage.schema(), storage.host(), storage.port()),
		creds.key,
		creds.secret,
	)
	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("configuring the minio client (%s)", err)
	}

	// $MC_PREFIX mb "$HTTP_PREFIX://${S3EP}/git" &>/dev/null
	makeBucketCmd := baseMinioCmd
	makeBucketCmd.Args = append(
		makeBucketCmd.Args,
		"mb",
		fmt.Sprintf("%s://%s:%s/git", storage.schema(), storage.host(), storage.port()),
	)
	// Don't look for errors here. Buckets may already exist
	makeBucketCmd.Run()

	// $MC_PREFIX cp ${APP_NAME}.tar.gz $TAR_URL &>/dev/null
	cpCmd := baseMinioCmd
	cpCmd.Args = append(
		cpCmd.Args,
		"cp",
		fmt.Sprintf("%s.tar.gz", appName),
		tarURL,
	)
	cpCmd.Dir = repoDir
	if err := cpCmd.Run(); err != nil {
		return fmt.Errorf("copying %s.tar.gz to %s (%s)", appName, tarURL, err)
	}

	//
	// puts-step "Starting build"
	// kubectl --namespace=${POD_NAMESPACE} create -f /etc/${SLUG_NAME}.yaml >/dev/null

	log.Info("Starting build")
	kubectlCmd := exec.Command(
		"kubectl",
		fmt.Sprintf("--namespace=%s", conf.PodNamespace),
		"create",
		"-f",
		fmt.Sprintf("/etc/%s.yaml", slugName),
	)
	kubectlCmd.Stderr = os.Stderr
	if err := kubectlCmd.Run(); err != nil {
		return fmt.Errorf("creating builder pod (%s)", err)
	}

	//
	// # wait for pod to be running and then pull its logs
	// until [ "`kubectl --namespace=${POD_NAMESPACE} get pods -o yaml ${META_NAME} | grep "phase: " | awk {'print $2'}`" == "Running" ]; do
	//     sleep 0.1
	// done
	// kubectl --namespace=${POD_NAMESPACE} logs -f ${META_NAME} 2>/dev/null &

	// poll kubectl every 100ms to determine when the build pod is running
	// TODO: use the k8s client and watch the event stream instead (https://github.com/deis/builder/issues/65)
	getCmd := exec.Command(
		"kubectl",
		fmt.Sprintf("--namespace=%s", conf.PodNamespace),
		fmt.Sprintf("get"),
		fmt.Sprintf("pods"),
		"-o",
		"yaml",
		buildPodName,
	)
	for {
		var out bytes.Buffer
		getCmd.Stdout = &out
		if err := getCmd.Run(); err != nil {
			return fmt.Errorf(
				"running %s while determining if builder pod %s is running (%s)",
				strings.Join(getCmd.Args, " "),
				buildPodName,
				err,
			)
		}
		if strings.Contains(string(out.Bytes()), "phase: Running") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// get logs from the builder pod
	logsCmd := exec.Command(
		"kubectl",
		fmt.Sprintf("--namespace=%s", conf.PodNamespace),
		"logs",
		"-f",
		buildPodName,
	)
	logsCmd.Stdout = os.Stdout
	if err := logsCmd.Run(); err != nil {
		return fmt.Errorf("running %s to get builder logs (%s)", strings.Join(logsCmd.Args, " "), err)
	}

	//
	// #check for image creation or slug existence in S3EP
	//
	// if [[ "$REMOTE_STORAGE" == "1" ]]; then
	//   LS_CMD="$MC_PREFIX ls $PUSH_URL"
	//   until $LS_CMD &> /dev/null; do
	//     echo -ne "."
	//     sleep 2
	//   done
	// else
	//   while [ ! -f /apps/${SLUG_NAME}/slug.tgz ]
	//   do
	//     echo -ne "."
	//     sleep 2
	//   done
	// fi

	// poll the s3 server to ensure the slug exists
	lsCmd := baseMinioCmd
	lsCmd.Args = append(lsCmd.Args, "ls", pushURL)
	for {
		// for now, assume the error indicates that the slug wasn't there, nothing else
		// TODO: implement https://github.com/deis/builder/issues/80, which will clean this up siginficantly
		if err := lsCmd.Run(); err == nil {
			break
		}
	}

	//
	// # build completed
	//
	// puts-step "Build complete."
	// puts-step "Launching app."
	//

	log.Info("Build complete.")
	log.Info("Launching app.")

	// URL="http://$DEIS_WORKFLOW_SERVICE_HOST:$DEIS_WORKFLOW_SERVICE_PORT/v2/hooks/config"
	// RESPONSE=$(get-app-config -url="$URL" -key="{{ getv "/deis/controller/builderKey" }}" -user=$USER -app=$APP_NAME)
	// CODE=$?
	// if [ $CODE -ne 0 ]; then
	//     puts-warn $RESPONSE
	//     exit 1
	// fi
	//

	cfg, err := getAppConfig(
		conf,
		builderKey,
		conf.Username,
		conf.App,
	)
	if err != nil {
		return fmt.Errorf("getting app config for %s (%s)", conf.App, err)
	}

	// # use Procfile if provided, otherwise try default process types from ./release
	//
	// puts-step "Launching... "
	// URL="http://$DEIS_WORKFLOW_SERVICE_HOST:$DEIS_WORKFLOW_SERVICE_PORT/v2/hooks/build"
	// DATA="$(generate-buildhook "$SHORT_SHA" "$USER" "$APP_NAME" "$APP_NAME" "$PROCFILE" "$USING_DOCKERFILE")"
	// PUBLISH_RELEASE=$(echo "$DATA" | publish-release-controller -url=$URL -key={{ getv "/deis/controller/builderKey" }})
	// CODE=$?
	// if [ $CODE -ne 0 ]; then
	//     puts-warn "ERROR: Failed to launch container"
	//     puts-warn $PUBLISH_RELEASE
	//     exit 1
	// fi

	log.Info("Launching...")

	buildHook := &pkg.BuildHook{
		Sha:         conf.SHA,
		ReceiveUser: conf.Username,
		ReceiveRepo: conf.Repository,
		Image:       conf.ImageName,
		Procfile:    procType,
		Dockerfile:  strings.Title(fmt.Sprintf("%t", usingDockerfile)),
	}
	buildHookResp, err := publishRelease(
		conf,
		builderKey,
		buildHook,
	)
	if err != nil {
		return fmt.Errorf("publishing release (%s)", err)
	}
	//
	// RELEASE=$(echo $PUBLISH_RELEASE | extract-version)
	// DOMAIN=$(echo $PUBLISH_RELEASE | extract-domain)
	// indent "done, $APP_NAME:v$RELEASE deployed to Deis"
	// echo
	// indent "http://$DOMAIN"
	// echo
	// indent "To learn more, use \`deis help\` or visit http://deis.io"
	// echo

	release, ok := buildHookResp.Release["version"]
	if !ok {
		return fmt.Errorf("No release returned from Deis controller")
	}
	if buildHookResp.Domains == nil || len(buildHookResp.Domains) == 0 {
		return fmt.Errorf("No domains returned from Deis controller")
	}
	domain := buildHookResp.Domains[0]

	log.Info("Done, %s:v%s deployed to Deis", appName, release)
	log.Info(fmt.Sprintf("http://%s", domain))
	log.Info("To learn more, use 'deis help' or visit http://deis.io")

	//
	// # cleanup
	// cd $REPO_DIR
	// git gc &>/dev/null

	gcCmd := exec.Command("git", "gc")
	gcCmd.Dir = repoDir
	if err := gcCmd.Run(); err != nil {
		return fmt.Errorf("cleaning up the repository with %s (%s)", strings.Join(gcCmd.Args, " "), err)
		// TODO: is it ok not to exit even if the repo was not cleaned up
	}

	return nil
}
