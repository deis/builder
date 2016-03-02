package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/deis/builder/pkg/sys"
)

const (
	accessKeyIDFile     = "/var/run/secrets/object/store/access-key-id"
	accessSecretKeyFile = "/var/run/secrets/object/store/access-secret-key"
)

var (
	errMissingKey    = fmt.Errorf("missing %s", accessKeyIDFile)
	errMissingSecret = fmt.Errorf("missing %s", accessSecretKeyFile)
	emptyCreds       = creds{}
)

type creds struct {
	accessKeyID     string
	accessKeySecret string
}

func (c *creds) isZero() bool {
	return c.accessKeyID == "" && c.accessKeySecret == ""
}

// getAuth gets storage credentials from accessKeyIDFile and accessSecretKeyFile.
// if a key exists but not a secret, or vice-versa, returns an error.
// if both don't exist returns emptyAuth.
// otherwise returns a valid auth
func getAuth(fs sys.FS) (*credentials.Credentials, error) {
	accessKeyIDBytes, accessKeyErr := fs.ReadFile(accessKeyIDFile)
	accessSecretKeyBytes, accessSecretKeyErr := fs.ReadFile(accessSecretKeyFile)
	if accessKeyErr == os.ErrNotExist && accessSecretKeyErr == os.ErrNotExist {
		return &emptyCreds, nil
	}
	if accessKeyErr != nil && accessSecretKeyErr == nil {
		return nil, errMissingKey
	}
	if accessKeyErr == nil && accessSecretKeyErr != nil {
		return nil, errMissingSecret
	}

	id := strings.TrimSpace(string(accessKeyIDBytes))
	secret := strings.TrimSpace(string(accessSecretKeyBytes))
	return &creds{accessKeyID: id, accessKeySecret: secret}, nil
}

// CredsOK checks if the required credentials to make a request exist
func CredsOK(fs sys.FS) bool {
	cred, err := getAuth(fs)
	if err != nil {
		return false
	}

	if creds.accessKeyID == "" && creds.accessKeySecret == "" {
		return false
	}

	return true
}
