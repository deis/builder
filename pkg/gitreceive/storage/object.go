package storage

import (
	"io"

	s3 "github.com/minio/minio-go"
)

const (
	noSuchKeyCode = "NoSuchKey"

	octetStream = "application/octet-stream"
)

// ObjectExists determines whether the object in ${bucketName}/${objKey} exists, as reported by statter. Returns the following:
//
// - false, nil if statter succeeded and reported the object doesn't exist
// - false, err with the appropriate error if the statter failed
// - true, nil if the statter succeeded and reported the object exists
func ObjectExists(statter ObjectStatter, bucketName, objKey string) (bool, error) {
	objInfo, err := statter.StatObject(bucketName, objKey)
	if err != nil {
		minioErr := s3.ToErrorResponse(err)
		if minioErr.Code == noSuchKeyCode {
			return false, nil
		}
		return false, err
	}
	if objInfo.Err != nil {
		return false, nil
	}
	return true, nil
}

func UploadObject(putter ObjectPutter, bucketName, objKey string, reader io.Reader) error {
	_, err := putter.PutObject(bucketName, objKey, reader, octetStream)
	return err
}
