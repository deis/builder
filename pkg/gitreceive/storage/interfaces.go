package storage

import (
	"io"

	s3 "github.com/minio/minio-go"
)

// BucketCreator is a *(github.com/minio/minio-go).Client compatible interface, restricted to just the MakeBucket function. You can use it in your code for easier unit testing without any external dependencies
type BucketCreator interface {
	MakeBucket(bucketName string, acl s3.BucketACL, location string) error
}

// FakeBucketCreator is a mock function that can be swapped in for an BucketCreator, so you can unit test your code
type FakeBucketCreator func(string, s3.BucketACL, string) error

// PutObject is the interface definition
func (f FakeBucketCreator) MakeBucket(name string, acl s3.BucketACL, location string) error {
	return f(name, acl, location)
}

// ObjectStatter is a *(github.com/minio/minio-go).Client compatible interface, restricted to just the StatObject function. You can use it in your code for easier unit testing without any external dependencies
type ObjectStatter interface {
	StatObject(bucketName, objectKey string) (s3.ObjectInfo, error)
}

// FakeObjectStatter is a mock function that can be swapped in for an ObjectStatter, so you can unit test your code
type FakeObjectStatter func(string, string) (s3.ObjectInfo, error)

// PutObject is the interface definition
func (f FakeObjectStatter) StatObject(bucketName, objectKey string) (s3.ObjectInfo, error) {
	return f(bucketName, objectKey)
}

// ObjectPutter is a *(github.com/minio/minio-go).Client compatible interface, restricted to just the PutObject function. You can use it in your code for easier unit testing without any external dependencies
type ObjectPutter interface {
	PutObject(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error)
}

// FakeObjectPutter is a mock function that can be swapped in for an ObjectPutter, so you can unit test your code
type FakeObjectPutter func(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error)

// PutObject is the interface definition
func (f FakeObjectPutter) PutObject(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error) {
	return f(bucketName, objectKey, reader, contentType)
}
