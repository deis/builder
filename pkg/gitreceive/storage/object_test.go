package storage

import (
	"testing"

	"github.com/arschles/assert"
	s3 "github.com/minio/minio-go"
)

const (
	objKey = "myobj"
)

func TestObjectExistsSuccess(t *testing.T) {
	objInfo := s3.ObjectInfo{Key: objKey, Err: nil, Size: 1234}
	statter := &FakeObjectStatter{
		Fn: func(string, string) (s3.ObjectInfo, error) {
			return objInfo, nil
		},
	}
	exists, err := ObjectExists(statter, bucketName, objKey)
	assert.NoErr(t, err)
	assert.True(t, exists, "object not found when it should be present")
	assert.Equal(t, len(statter.Calls), 1, "number of StatObject calls")
	assert.Equal(t, statter.Calls[0].BucketName, bucketName, "bucket name")
	assert.Equal(t, statter.Calls[0].ObjectKey, objKey, "object key")
}

func TestObjectExistsNoObject(t *testing.T) {
	statter := &FakeObjectStatter{
		Fn: func(string, string) (s3.ObjectInfo, error) {
			return s3.ObjectInfo{}, s3.ErrorResponse{Code: noSuchKeyCode}
		},
	}
	exists, err := ObjectExists(statter, bucketName, objKey)
	assert.NoErr(t, err)
	assert.False(t, exists, "object found when it should be missing")
	assert.Equal(t, len(statter.Calls), 1, "number of StatObject calls")
}
