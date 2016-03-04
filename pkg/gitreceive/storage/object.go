package storage

import (
	"errors"
	"fmt"
	"io"
	"time"

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
	_, err := statter.StatObject(bucketName, objKey)
	if err != nil {
		minioErr := s3.ToErrorResponse(err)
		if minioErr.Code == noSuchKeyCode {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func UploadObject(putter ObjectPutter, bucketName, objKey string, reader io.Reader) error {
	_, err := putter.PutObject(bucketName, objKey, reader, octetStream)
	return err
}

// WaitForObject checks statter for the object at ${bucketName}/${objKey} right away, then at every tick, then once when the timeout is up.
// Returns nil if it finds the object before or at timeout. Otherwise returns an error
func WaitForObject(statter ObjectStatter, bucketName, objKey string, tick, timeout time.Duration) error {
	noExist := errors.New("object doesn't exist")
	checker := func() error {
		exists, err := ObjectExists(statter, bucketName, objKey)
		if err != nil {
			return err
		} else if exists {
			return nil
		} else {
			return noExist
		}
	}
	if err := checker(); err == nil {
		return nil
	}
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ticker.C:
			if err := checker(); err == nil {
				return nil
			}
		case <-timer.C:
			if err := checker(); err == nil {
				return nil
			}
			return fmt.Errorf("Object %s/%s didn't exist after %s", bucketName, objKey, timeout)
		}
	}
}
