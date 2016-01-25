package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mitchellh/goamz/aws"
)

const (
	accessKeyIDFile     = "/var/run/secrets/object/store/access-key-id"
	accessSecretKeyFile = "/var/run/secrets/object/store/access-secret-key"
)

var (
	errMissingKey    = fmt.Errorf("missing %s", accessKeyIDFile)
	errMissingSecret = fmt.Errorf("missing %s", accessSecretKeyFile)
	emptyAuth        = &aws.Auth{}
)

// getCreds gets storage credentials from accessKeyIDFile and accessSecretKeyFile.
// if a key exists but not a secret, or vice-versa, returns an error.
// if both don't exist returns emptyAuth.
// otherwise returns a valid auth
func getAuth() (*aws.Auth, error) {
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

	return &aws.Auth{
		AccessKey: strings.TrimSpace(string(accessKeyIDBytes)),
		SecretKey: strings.TrimSpace(string(accessSecretKeyBytes)),
	}, nil
}
