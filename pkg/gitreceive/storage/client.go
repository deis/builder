package storage

import (
	"github.com/deis/builder/pkg/sys"
	s3 "github.com/minio/minio-go"
)

type Client struct {
	*s3.Client
	Endpoint *Endpoint
}

// GetClient returns a S3 API compatible storage client
func GetClient(regionStr string, fs sys.FS, env sys.Env) (*Client, error) {
	auth, err := getAuth(fs)
	if err != nil {
		return nil, err
	}

	endpoint, err := getEndpoint(env)
	if err != nil {
		return nil, err
	}

	// the New function call guesses which signature version to use. Currently, it correctly guesses V2 for GCS and V4 for both AWS S3 and Minio
	s3Client, err := s3.New(endpoint.URLStr, auth.accessKeyID, auth.accessKeySecret, false)
	if err != nil {
		return nil, err
	}
	return &Client{Client: s3Client, Endpoint: endpoint}, nil
}
