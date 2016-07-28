package controller

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/deis/builder/pkg/conf"
	"golang.org/x/crypto/ssh"
)

const (
	hostEnvName = "DEIS_CONTROLLER_SERVICE_HOST"
	portEnvName = "DEIS_CONTROLLER_SERVICE_PORT"
)

// UserInfo represents the required information from a user to make a push and interact with the
// controller
type UserInfo struct {
	Username    string
	Key         string
	Fingerprint string
	Apps        []string
}

// URLStr returns the url of the controller with the given path prepended.
func URLStr(additionalPath ...string) (string, error) {
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

// fingerprint generates a colon-separated fingerprint string from a public key.
func fingerprint(key ssh.PublicKey) string {
	hash := md5.Sum(key.Marshal())
	buf := make([]byte, hex.EncodedLen(len(hash)))
	hex.Encode(buf, hash[:])
	// We need this in colon notation:
	fp := make([]byte, len(buf)+15)

	i, j := 0, 0
	for ; i < len(buf); i++ {
		if i > 0 && i%2 == 0 {
			fp[j] = ':'
			j++
		}
		fp[j] = buf[i]
		j++
	}
	return string(fp)
}

// UserInfoFromKey makes a request to the controller to get the user info from they given key.
func UserInfoFromKey(key ssh.PublicKey) (*UserInfo, error) {
	fp := fingerprint(key)
	url, err := URLStr("v2", "hooks", "key", fp)
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
	ret.Fingerprint = fp
	return ret, nil
}
