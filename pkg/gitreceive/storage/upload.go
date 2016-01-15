package storage

import (
	"io"

	"github.com/aws/aws-sdk-go/service/s3"
)

func Upload(s3Client *s3.S3, bucketName, objKey string, reader io.Reader) error {
	// see https://godoc.org/github.com/aws/aws-sdk-go/service/s3#example-S3-PutObject
	return nil
}
