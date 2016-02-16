package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	bucketAlreadyExistsCode = "BucketAlreadyExists"
)

var (
	// ACLPublicRead default ACL for objects in the S3 API compatible storage
	ACLPublicRead = aws.String("public-read")
)

// CreateBucket creates a new bucket in the S3 API compatible storage or
// return an error in case the bucket already exists
func CreateBucket(svc *s3.S3, bucketName string) error {
	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		ACL:    ACLPublicRead,
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == bucketAlreadyExistsCode {
				return nil
			}
		}

		return err
	}

	return nil
}
