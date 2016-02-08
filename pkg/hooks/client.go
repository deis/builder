package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/deis/builder/pkg"
)

const (
	contentType string = "application/json"
	userAgent   string = "deis-builder"
)

var (
	potentialExploit = regexp.MustCompile(`\(\)\s+\{[^\}]+\};\s+(.*)`)
)

type UnexpectedControllerStatusCode struct {
	endpoint string
	expected int
	actual   int
}

func NewUnexpectedControllerStatusCode(endpoint string, expectedCode, actualCode int) UnexpectedControllerStatusCode {
	return UnexpectedControllerStatusCode{endpoint: endpoint, expected: expectedCode, actual: actualCode}
}

func (u UnexpectedControllerStatusCode) Error() string {
	return fmt.Sprintf("Deis controller endpoint %s: expected status code %d, got %d", u.endpoint, u.expected, u.actual)
}

func controllerURLStr(host string, port int, additionalPath ...string) string {
	return fmt.Sprintf("http://%s:%d/%s", host, port, strings.Join(additionalPath, "/"))
}

func setReqHeaders(builderKey string, req *http.Request) {
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("X-Deis-Builder-Auth", builderKey)
}

type Client struct {
	// TODO: use ctxhttp here (https://godoc.org/golang.org/x/net/context/ctxhttp)
	cl   *http.Client
	host string
	port int
	bKey string
}

func NewClient(cl *http.Client, host string, port int, builderKey string) *Client {
	return &Client{cl: cl, host: host, port: port, bKey: builderKey}
}

func (c *Client) GetAppConfig(appName, userName string) (*pkg.Config, error) {
	url := controllerURLStr(c.host, c.port, "v2", "hooks", "config")
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

	setReqHeaders(c.bKey, req)

	res, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, NewUnexpectedControllerStatusCode(url, 200, res.StatusCode)
	}
	ret := &pkg.Config{}
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) PublishRelease(buildHook *pkg.BuildHook) (*pkg.BuildHookResponse, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(buildHook); err != nil {
		return nil, err
	}

	postBody := strings.Replace(string(b.Bytes()), "'", "", -1)
	if potentialExploit.MatchString(postBody) {
		return nil, fmt.Errorf("an environment variable in the app is trying to exploit Shellshock")
	}

	urlStr := controllerURLStr(c.host, c.port, "v2", "hooks", "build")
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(postBody))
	if err != nil {
		return nil, err
	}
	setReqHeaders(c.bKey, req)

	res, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, NewUnexpectedControllerStatusCode(urlStr, 200, res.StatusCode)
	}

	ret := new(pkg.BuildHookResponse)
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return nil, err
	}

	return ret, nil
}

type ReceiveInfo struct {
	ReceiveUser        string `json:"receive_user"`
	ReceiveRepo        string `json:"receive_repo"`
	SHA                string `json:"sha"`
	Fingerprint        string `json:"fingerprint"`
	SSHConnection      string `json:"ssh_connection"`
	SSHOriginalCommand string `json:"ssh_original_command"`
}

func (c *Client) Receive(info *ReceiveInfo) error {
	urlStr := controllerURLStr(c.host, c.port, "v2", "hooks", "push")

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(info); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", urlStr, &body)
	if err != nil {
		return err
	}
	setReqHeaders(c.bKey, req)

	resp, err := c.cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return NewUnexpectedControllerStatusCode(urlStr, 201, resp.StatusCode)
	}
	return nil
}
