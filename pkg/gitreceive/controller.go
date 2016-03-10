package gitreceive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/gitreceive/git"
	"github.com/deis/pkg/log"
)

var (
	potentialExploit = regexp.MustCompile(`\(\)\s+\{[^\}]+\};\s+(.*)`)
)

type unexpectedControllerStatusCode struct {
	endpoint string
	expected int
	actual   int
}

func newUnexpectedControllerStatusCode(endpoint string, expectedCode, actualCode int) unexpectedControllerStatusCode {
	return unexpectedControllerStatusCode{endpoint: endpoint, expected: expectedCode, actual: actualCode}
}

func (u unexpectedControllerStatusCode) Error() string {
	return fmt.Sprintf("Deis controller endpoint %s: expected status code %d, got %d", u.endpoint, u.expected, u.actual)
}

func controllerURLStr(conf *Config, additionalPath ...string) string {
	return fmt.Sprintf("http://%s:%s/%s", conf.ControllerHost, conf.ControllerPort, strings.Join(additionalPath, "/"))
}

func setReqHeaders(builderKey string, req *http.Request) {
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("X-Deis-Builder-Auth", builderKey)
}

func getAppConfig(conf *Config, builderKey, userName, appName string) (*pkg.Config, error) {
	url := controllerURLStr(conf, "v2", "hooks", "config")
	data, err := json.Marshal(&pkg.ConfigHook{
		ReceiveUser: userName,
		ReceiveRepo: appName,
	})
	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(data)
	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		return nil, err
	}

	setReqHeaders(builderKey, req)

	log.Debug("Controller request POST /v2/hooks/config\n%s", string(data))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, newUnexpectedControllerStatusCode(url, 200, res.StatusCode)
	}
	ret := &pkg.Config{}
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func publishRelease(conf *Config, builderKey string, buildHook *pkg.BuildHook) (*pkg.BuildHookResponse, error) {

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(buildHook); err != nil {
		return nil, err
	}

	postBody := strings.Replace(string(b.Bytes()), "'", "", -1)
	if potentialExploit.MatchString(postBody) {
		return nil, fmt.Errorf("an environment variable in the app is trying to exploit Shellshock")
	}

	url := controllerURLStr(conf, "v2", "hooks", "build")
	log.Debug("Controller request POST /v2/hooks/build\n%s", postBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(postBody))
	if err != nil {
		return nil, err
	}
	setReqHeaders(builderKey, req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, newUnexpectedControllerStatusCode(url, 200, res.StatusCode)
	}

	ret := new(pkg.BuildHookResponse)
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func receive(conf *Config, builderKey, gitSha string) error {
	urlStr := controllerURLStr(conf, "v2", "hooks", "push")
	bodyMap := map[string]string{
		"receive_user":         conf.Username,
		"receive_repo":         conf.App(),
		"sha":                  gitSha,
		"fingerprint":          conf.Fingerprint,
		"ssh_connection":       conf.SSHConnection,
		"ssh_original_command": conf.SSHOriginalCommand,
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(bodyMap); err != nil {
		return err
	}

	log.Debug("Controller request /v2/hooks/push (body elided)")
	req, err := http.NewRequest("POST", urlStr, &body)
	if err != nil {
		return err
	}
	setReqHeaders(builderKey, req)

	// TODO: use ctxhttp here (https://godoc.org/golang.org/x/net/context/ctxhttp)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return newUnexpectedControllerStatusCode(urlStr, 201, resp.StatusCode)
	}
	return nil
}

func createBuildHook(
	slugBuilderInfo *SlugBuilderInfo,
	gitSha *git.SHA,
	username,
	appName string,
	procType pkg.ProcessType,
	usingDockerfile bool,
) *pkg.BuildHook {
	ret := &pkg.BuildHook{
		Sha:         gitSha.Short(),
		ReceiveUser: username,
		ReceiveRepo: appName,
		Image:       appName,
		Procfile:    procType,
	}
	if !usingDockerfile {
		ret.Dockerfile = ""
		// need this to tell the controller what URL to give the slug runner
		ret.Image = slugBuilderInfo.AbsoluteSlugURL()
	} else {
		ret.Dockerfile = "true"
	}
	return ret
}
