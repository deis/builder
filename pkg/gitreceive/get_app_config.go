package gitreceive

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/deis/builder/pkg"
)

func getAppConfig(conf *Config, builderKey, userName, appName string) (*pkg.Config, error) {
	data, err := json.Marshal(&pkg.ConfigHook{
		ReceiveUser: userName,
		ReceiveRepo: appName,
	})

	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(data)
	url := controllerURLStr(conf, "v2", "hooks", "config")
	req, err := http.NewRequest("POST", url, b)

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

	if res.StatusCode == 404 {
		return nil, errControllerNotFound
	} else if res.StatusCode != 200 {
		return nil, unexpectedControllerStatusCode{expected: 200, actual: res.StatusCode}
	}

	ret := &pkg.Config{}
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}
