package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func CreateBucket(s3Client *s3.S3, bucketName string) error {
	_, err := s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}
