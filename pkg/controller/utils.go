package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/deis/builder/pkg/conf"
)

const (
	hostEnvName = "DEIS_WORKFLOW_SERVICE_HOST"
	portEnvName = "DEIS_WORKFLOW_SERVICE_PORT"
)

// UserInfo represent the required information from a user to make a push and interact with deis/workflow
type UserInfo struct {
	Username    string
	Key         string
	FingerPrint string
	Apps        []string
}

func controllerURLStr(additionalPath ...string) (string, error) {
	host := os.Getenv(hostEnvName)
	port := os.Getenv(portEnvName)

	if host == "" {
		return "", fmt.Errorf("missing required '%v' environment variable", hostEnvName)
	}

	if port == "" {
		return "", fmt.Errorf("missing required '%v' environment variable", portEnvName)
	}

	return fmt.Sprintf("http://%s:%s/%s", host, port, strings.Join(additionalPath, "/")), nil
}

func UserInfoFromKey(key string) (*UserInfo, error) {
	keyB64 := base64.RawURLEncoding.EncodeToString([]byte(key))
	url, err := controllerURLStr("v2", "hooks", "key", keyB64)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	builderKey, err := conf.GetBuilderKey()
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "deis-builder")
	req.Header.Add("X-Deis-Builder-Auth", builderKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("user has no permissions to push in application")
	}

	ret := &UserInfo{}
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}
