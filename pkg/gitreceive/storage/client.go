package storage

import (
	"github.com/deis/builder/pkg/sys"
	s3 "github.com/minio/minio-go"
)

// GetClient returns a S3 API compatible storage client
func GetClient(regionStr string, fs sys.FS, env sys.Env) (*s3.Client, error) {
	auth, err := getAuth(fs)
	if err != nil {
		return nil, err
	}

	endpoint, err := getEndpoint(env)
	if err != nil {
		return nil, err
	}

	return s3.New(endpoint, auth.accessKeyID, auth.accessKeySecret, false), nil
}
