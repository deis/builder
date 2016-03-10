package storage

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	s3 "github.com/minio/minio-go"
)

const (
	noSuchKeyCode = "NoSuchKey"

	octetStream = "application/octet-stream"
)

// ObjectExists determines whether the object in ${bucketName}/${objKey} exists, as reported by
// statter. Returns the following:
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

// UploadObject uploads the contents of readaer to ${bucektName}/${objectKey} using the given putter
func UploadObject(putter ObjectPutter, bucketName, objKey string, reader io.Reader) error {
	_, err := putter.PutObject(bucketName, objKey, reader, octetStream)
	return err
}

// WaitForObject checks statter for the object at ${bucketName}/${objKey} right away, then at
// every tick, then once when the timeout is up.
// Returns nil if it finds the object before or at timeout. Otherwise returns a non-nil error
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

// DownloadObject uses the given getter to download the contents the object at
// ${bucketName}/${objKey} and returns the object's contents in the given byte slice.
// Returns nil and the appropriate error if there were problems with the download
func DownloadObject(getter ObjectGetter, bucketName, objKey string) ([]byte, error) {
	reader, err := getter.GetObject(bucketName, objKey)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return data, err
}
