package storage

import (
	"testing"

	"github.com/arschles/assert"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/deis/builder/pkg/sys"
)

func TestGetAuthEmptyAuth(t *testing.T) {
	fs := sys.NewFakeFS()
	creds, err := getAuth(fs)
	assert.NoErr(t, err)
	assert.Equal(t, creds, credentials.AnonymousCredentials, "returned credentials")
}

func TestGetAuthMissingSecret(t *testing.T) {
	fs := sys.NewFakeFS()
	fs.Files[accessSecretKeyFile] = []byte("hello world")
	creds, err := getAuth(fs)
	assert.Err(t, err, errMissingKey)
	assert.True(t, creds == nil, "returned credentials were not nil")
}

func TestGetAuthMissingKey(t *testing.T) {
	fs := sys.NewFakeFS()
	fs.Files[accessKeyIDFile] = []byte("hello world")
	creds, err := getAuth(fs)
	assert.Err(t, err, errMissingSecret)
	assert.True(t, creds == nil, "returned credentials were not nil")
}

func TestGetAuthSuccess(t *testing.T) {
	fs := sys.NewFakeFS()
	fs.Files[accessKeyIDFile] = []byte("invalid")
	fs.Files[accessSecretKeyFile] = []byte("also invalid")
	creds, err := getAuth(fs)
	assert.NoErr(t, err)
	assert.True(t, creds != nil, "creds were nil when they shouldn't have been")
}
