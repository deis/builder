package storage

import (
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

func GetClient(regionStr string) (*s3.S3, error) {
	auth, err := getAuth()
	if err != nil {
		return nil, err
	}

	endpoint, err := getEndpoint()
	if err != nil {
		return nil, err
	}

	region := aws.Region{
		Name:       regionStr,
		S3Endpoint: endpoint,
	}

	return s3.New(*auth, region), nil
}
