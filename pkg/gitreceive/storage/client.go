package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// GetClient returns a S3 API compatible storage client
func GetClient(regionStr string) (*s3.S3, error) {
	auth, err := getAuth()
	if err != nil {
		return nil, err
	}

	endpoint, err := getEndpoint()
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
