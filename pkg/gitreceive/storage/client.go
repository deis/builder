package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/deis/builder/pkg/sys"
)

// GetClient returns a S3 API compatible storage client
func GetClient(regionStr string, fs sys.FS, env sys.Env) (*s3.S3, error) {
	auth, err := getAuth(fs)
	if err != nil {
		return nil, err
	}

	endpoint, err := getEndpoint(env)
	if err != nil {
		return nil, err
	}

	return s3.New(session.New(&aws.Config{
		Credentials:      auth,
		Region:           aws.String(regionStr),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})), nil
}
