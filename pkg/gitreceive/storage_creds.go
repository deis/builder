package gitreceive

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	accessKeyIDFile     = "/var/run/secrets/object/store/access-key-id"
	accessSecretKeyFile = "/var/run/secrets/object/store/access-secret-key"
)

var (
	errMissingKey    = fmt.Errorf("missing %s", accessKeyIDFile)
	errMissingSecret = fmt.Errorf("missing %s", accessSecretKeyFile)
)

type storageCreds struct {
	key    string
	secret string
}

// getStorageCreds gets storage credentials from accessKeyIDFile and accessSecretKeyFile.
// returns os.ErrNotExist if both files are missing and otherwise, if a file was missing,
// returns errMissingKey or errMissingSecret according to the file
func getStorageCreds() (*storageCreds, error) {
	accessKeyIDBytes, accessKeyErr := ioutil.ReadFile(accessKeyIDFile)
	accessSecretKeyBytes, accessSecretKeyErr := ioutil.ReadFile(accessSecretKeyFile)
	if accessKeyErr == os.ErrNotExist && accessSecretKeyErr == os.ErrNotExist {
		return nil, os.ErrNotExist
	}
	if accessKeyErr != nil && accessSecretKeyErr == nil {
		return nil, errMissingKey
	}
	if accessKeyErr == nil && accessSecretKeyErr != nil {
		return nil, errMissingSecret
	}
	accessKeyID := strings.TrimSpace(string(accessKeyIDBytes))
	accessSecretKey := strings.TrimSpace(string(accessSecretKeyBytes))
	return &storageCreds{key: accessKeyID, secret: accessSecretKey}, nil
}
