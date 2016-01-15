package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func GetClient() (*s3.S3, error) {
	awsCfg := new(aws.Config)
	endpt, err := getEndpoint()
	if err != nil {
		return nil, err
	}
	awsCfg.Endpoint = endpt
	creds, err := getCreds()
	if err != nil {
		return nil, err
	}
	awsCfg.Credentials = creds
	svc := s3.New(session.New(awsCfg), &aws.Config{})
	return svc, nil
}
