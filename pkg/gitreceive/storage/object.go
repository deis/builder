package storage

import (
	"io"

	s3 "github.com/minio/minio-go"
)

const (
	noSuchKeyCode = "NoSuchKey"

	octetStream = "application/octet-stream"
)

func ObjectExists(statter ObjectStatter, bucket, objName string) (bool, error) {
	objInfo, err := statter.StatObject(bucketName, objKey)
	if err != nil {
		return false, err
	}
	if objInfo.Code == noSuchKeyCode || objInfo.Err != nil {
		return false, nil
	}
	return true, nil
}

func UploadObject(putter ObjectPutter, bucketName, objKey string, reader io.Reader) error {
	_, err := putter.PutObject(bucketName, objKey, reader, octetStream)
	return err
}
