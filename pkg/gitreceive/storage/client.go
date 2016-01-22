package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func GetClient(region string) (*s3.S3, error) {
	awsCfg := new(aws.Config)
	awsCfg.Region = aws.String(region)
	endpt, err := getEndpoint()
	if err != nil {
		return nil, err
	}
	awsCfg.Endpoint = aws.String(endpt)
	creds, err := getCreds()
	if err != nil {
		return nil, err
	}
	awsCfg.Credentials = creds
	svc := s3.New(session.New(awsCfg), &aws.Config{})
	return svc, nil
}
