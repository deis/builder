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
	var res bucketCreate
	creator := FakeBucketCreator(func(name string, acl s3.BucketACL, location string) error {
		res = bucketCreate{name: name, acl: acl, loc: location}
		return nil
	})

	assert.NoErr(t, CreateBucket(creator, bucketName))
	assert.Equal(t, res.name, bucketName, "bucket name")
	assert.Equal(t, res.acl, ACLPublicRead, "bucket ACL")
	assert.Equal(t, res.loc, "", "bucket location")
}

func TestCreateBucketFailure(t *testing.T) {
	err := errors.New("test err")
	creator := FakeBucketCreator(func(string, s3.BucketACL, string) error {
		return err
	})
	assert.Err(t, CreateBucket(creator, bucketName), err)
}
