package gitreceive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/git"
	"github.com/deis/pkg/log"
)

var (
	potentialExploit = regexp.MustCompile(`\(\)\s+\{[^\}]+\};\s+(.*)`)
)

type unexpectedControllerError struct {
	errorMsg string
}

func newUnexpectedControllerError(errorMsg string) unexpectedControllerError {
	return unexpectedControllerError{errorMsg: errorMsg}
}

func (u unexpectedControllerError) Error() string {
	return fmt.Sprintf("Unexpected error occurred: %s", u.errorMsg)
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

	if res.StatusCode < 200 || res.StatusCode > 299 {
		errMsg := new(pkg.ControllerErrorResponse)
		if err := json.NewDecoder(res.Body).Decode(errMsg); err != nil {
			//If an error occurs decoding the json print the whole response body
			respBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return nil, err
			}
			return nil, newUnexpectedControllerError(string(respBody))
		}

		return nil, newUnexpectedControllerError(errMsg.ErrorMsg)
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

	if res.StatusCode < 200 || res.StatusCode > 299 {
		errMsg := new(pkg.ControllerErrorResponse)
		if err := json.NewDecoder(res.Body).Decode(errMsg); err != nil {
			//If an error occurs decoding the json print the whole response body
			respBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return nil, err
			}
			return nil, newUnexpectedControllerError(string(respBody))
		}

		return nil, newUnexpectedControllerError(errMsg.ErrorMsg)
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

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errMsg := new(pkg.ControllerErrorResponse)
		if err := json.NewDecoder(resp.Body).Decode(errMsg); err != nil {
			//If an error occurs decoding the json print the whole response body
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return newUnexpectedControllerError(string(respBody))
		}

		return newUnexpectedControllerError(errMsg.ErrorMsg)
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
		ret.Image = slugBuilderInfo.AbsoluteSlugObjectKey()
	} else {
		ret.Dockerfile = "true"
	}
	return ret
}
