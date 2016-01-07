package gitreceive

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/deis/builder/pkg"
)

func publishRelease(conf *Config, builderKey string, buildHook *pkg.BuildHook) (*pkg.BuildHookResponse, error) {

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(buildHook); err != nil {
		return nil, err
	}
	url := controllerURLStr(conf, "v2", "hooks", "build")
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("X-Deis-Builder-Auth", builderKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, errControllerNotFound
	} else if res.StatusCode == http.StatusServiceUnavailable {
		return nil, errControllerServiceUnavailable
	} else if res.StatusCode != 200 {
		return nil, newUnexpectedControllerStatusCode(url, 200, res.StatusCode)
	}

	ret := new(pkg.BuildHookResponse)
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return nil, err
	}

	return ret, nil
}
