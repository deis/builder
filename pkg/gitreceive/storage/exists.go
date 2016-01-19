package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func ObjectExists(s3Client *s3.S3, bucket, objName string) bool {
	// see https://godoc.org/github.com/aws/aws-sdk-go/service/s3#example-S3-HeadObject
	in := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objName),
	}
	resp, err := s3Client.HeadObject(in)
	if err != nil {
		return false
	}
	return resp != nil
}
