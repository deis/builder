package storage

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func ObjectExists(svc *s3.S3, bucket, objName string) (bool, error) {
	_, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func UploadObject(svc *s3.S3, bucketName, objKey string, reader io.Reader) error {
	params := &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(reader),
		Bucket: aws.String(bucketName),
		Key:    aws.String(objKey),
		ACL:    ACLPublicRead,
	}
	_, err := svc.PutObject(params)
	return err
}
