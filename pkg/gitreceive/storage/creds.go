package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

const (
	accessKeyIDFile     = "/var/run/secrets/object/store/access-key-id"
	accessSecretKeyFile = "/var/run/secrets/object/store/access-secret-key"
)

var (
	errMissingKey    = fmt.Errorf("missing %s", accessKeyIDFile)
	errMissingSecret = fmt.Errorf("missing %s", accessSecretKeyFile)
)

// getCreds gets storage credentials from accessKeyIDFile and accessSecretKeyFile.
// if either a key exists but not a secret or vice-versa, returns an error.
// if neither exists, returns credentials.AnonymousCredentials
func getCreds() (*credentials.Credentials, error) {
	accessKeyIDBytes, accessKeyErr := ioutil.ReadFile(accessKeyIDFile)
	accessSecretKeyBytes, accessSecretKeyErr := ioutil.ReadFile(accessSecretKeyFile)
	if accessKeyErr == os.ErrNotExist && accessSecretKeyErr == os.ErrNotExist {
		return credentials.AnonymousCredentials, nil
	}
	if accessKeyErr != nil && accessSecretKeyErr == nil {
		return nil, errMissingKey
	}
	if accessKeyErr == nil && accessSecretKeyErr != nil {
		return nil, errMissingSecret
	}
	accessKeyID := strings.TrimSpace(string(accessKeyIDBytes))
	accessSecretKey := strings.TrimSpace(string(accessSecretKeyBytes))
	return credentials.NewStaticCredentials(accessKeyID, accessSecretKey, ""), nil
}
