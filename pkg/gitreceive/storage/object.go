package storage

import (
	"io"
	"net/http"

	"github.com/mitchellh/goamz/s3"
)

func ObjectExists(svc *s3.S3, bucket, objName string) (bool, error) {
	resp, err := svc.Bucket(bucket).Head(objName)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil
}

func UploadObject(svc *s3.S3, bucketName, objKey string, reader io.Reader) error {
	// see https://godoc.org/github.com/aws/aws-sdk-go/service/s3#example-S3-PutObject
	return nil
}
