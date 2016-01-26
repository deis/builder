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
	emptyAuth        = credentials.AnonymousCredentials
)

// getAuth gets storage credentials from accessKeyIDFile and accessSecretKeyFile.
// if a key exists but not a secret, or vice-versa, returns an error.
// if both don't exist returns emptyAuth.
// otherwise returns a valid auth
func getAuth() (*credentials.Credentials, error) {
	accessKeyIDBytes, accessKeyErr := ioutil.ReadFile(accessKeyIDFile)
	accessSecretKeyBytes, accessSecretKeyErr := ioutil.ReadFile(accessSecretKeyFile)
	if accessKeyErr == os.ErrNotExist && accessSecretKeyErr == os.ErrNotExist {
		return emptyAuth, nil
	}
	if accessKeyErr != nil && accessSecretKeyErr == nil {
		return nil, errMissingKey
	}
	if accessKeyErr == nil && accessSecretKeyErr != nil {
		return nil, errMissingSecret
	}

	id := strings.TrimSpace(string(accessKeyIDBytes))
	secret := strings.TrimSpace(string(accessSecretKeyBytes))
	return credentials.NewStaticCredentials(id, secret, ""), nil
}

// CredsOK checks if the required credentials to make a request exist
func CredsOK() bool {
	cred, err := getAuth()
	if err != nil {
		return false
	}

	auth, _ := cred.Get()
	if auth.AccessKeyID == "" && auth.SecretAccessKey == "" {
		return false
	}

	return true
}
