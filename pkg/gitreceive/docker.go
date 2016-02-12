package gitreceive

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/deis/builder/pkg/gitreceive/git"
	"github.com/fsouza/go-dockerclient"
)

const (
	endpoint   = "unix:///var/run/docker.sock"
	slugRunner = "dev-registry.soficom.cl/deis/slugrunner:v1.0.0"

	appTemplate = `FROM {{ .baseImage }}

ADD {{ .tgzURL }}
GIT_SHA {{ .gitSHA }}

`
)

var appTemplateTpl = template.Must(template.New("appTemplate").Parse(appTemplate))

type buildContext struct {
	AppName string
	Sha     *git.SHA
	Tgz     string
}

func buildImage(context *buildContext) (string, error) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return "", err
	}

	var hookByteBuf bytes.Buffer
	err = appTemplateTpl.Execute(&hookByteBuf, map[string]string{
		"baseImage": slugRunner,
		"tgzURL":    context.Tgz,
		"gitSHA":    context.Sha.Full(),
	})
	if err != nil {
		return "", err
	}

	tempDir, err := ioutil.TempDir("", "build-app")
	if err != nil {
		return "", fmt.Errorf("unexpected error creating temporal directory %v", err)
	}

	err = ioutil.WriteFile(tempDir+"/Dockerfile", hookByteBuf.Bytes(), 0644)
	if err != nil {
		return "", err
	}

	dockerName := getImageName(context.AppName, context.Sha.Short())

	opts := docker.BuildImageOptions{
		Name:           dockerName,
		Dockerfile:     tempDir + "/Dockerfile",
		RmTmpContainer: true,
		OutputStream:   os.Stdout,
		Pull:           true,
	}

	err = client.BuildImage(opts)
	if err != nil {
		return "", err
	}

	err = client.PushImage(docker.PushImageOptions{
		Name:         dockerName,
		OutputStream: os.Stdout,
	}, docker.AuthConfiguration{})
	if err != nil {
		return "", err
	}

	return dockerName, nil
}

func getImageName(appName, shortSha string) string {
	host := os.Getenv("DEIS_REGISTRY_SERVICE_HOST")
	port := os.Getenv("DEIS_REGISTRY_SERVICE_PORT")

	return fmt.Sprintf("%v:%v/%v:git-%v", host, port, appName, shortSha)
}
