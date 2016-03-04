package storage

import (
	"errors"
	"testing"

	"github.com/arschles/assert"
	s3 "github.com/minio/minio-go"
)

const (
	bucketName = "mybucket"
)

type bucketCreate struct {
	name string
	acl  s3.BucketACL
	loc  string
}

func TestCreateBucketSuccess(t *testing.T) {
	creator := &FakeBucketCreator{
		Fn: func(name string, acl s3.BucketACL, location string) error {
			return nil
		},
	}

	assert.NoErr(t, CreateBucket(creator, bucketName))
	assert.Equal(t, len(creator.Calls), 1, "number of calls to MakeBucket")
	assert.Equal(t, creator.Calls[0].BucketName, bucketName, "bucket name")
	assert.Equal(t, creator.Calls[0].ACL, ACLPublicRead, "bucket ACL")
	assert.Equal(t, creator.Calls[0].Location, "", "bucket location")
}

func TestCreateBucketFailure(t *testing.T) {
	err := errors.New("test err")
	creator := &FakeBucketCreator{
		Fn: func(string, s3.BucketACL, string) error {
			return err
		},
	}
	assert.Err(t, CreateBucket(creator, bucketName), err)
	assert.Equal(t, len(creator.Calls), 1, "number of calls to MakeBucket")
}
