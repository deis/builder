package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	// ACLPublicRead default ACL for objects in the S3 API compatible storage
	ACLPublicRead = aws.String("public-read")
)

// BucketExists returns if a bucket exists in the S3 API compatible storage
func BucketExists(svc *s3.S3, bucketName string) (bool, error) {
	_, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "404" {
				return false, nil
			}
		}

		return false, err
	}
	return true, nil
}

// CreateBucket creates a new bucket in the S3 API compatible storage or
// return an error in case the bucket already exists
func CreateBucket(svc *s3.S3, bucketName string) error {
	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		ACL:    ACLPublicRead,
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "409" {
				return nil
			}
		}

		return err
	}

	return nil
}
