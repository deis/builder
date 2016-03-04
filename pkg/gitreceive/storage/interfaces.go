package storage

import (
	"io"

	s3 "github.com/minio/minio-go"
)

// BucketCreator is a *(github.com/minio/minio-go).Client compatible interface, restricted to just the MakeBucket function. You can use it in your code for easier unit testing without any external dependencies
type BucketCreator interface {
	MakeBucket(bucketName string, acl s3.BucketACL, location string) error
}

type FakeMakeBucketCall struct {
	BucketName string
	ACL        s3.BucketACL
	Location   string
}

// FakeBucketCreator is a mock function that can be swapped in for an BucketCreator, so you can unit test your code
type FakeBucketCreator struct {
	Fn    func(string, s3.BucketACL, string) error
	Calls []FakeMakeBucketCall
}

// PutObject is the interface definition
func (f *FakeBucketCreator) MakeBucket(name string, acl s3.BucketACL, location string) error {
	f.Calls = append(f.Calls, FakeMakeBucketCall{BucketName: name, ACL: acl, Location: location})
	return f.Fn(name, acl, location)
}

// ObjectStatter is a *(github.com/minio/minio-go).Client compatible interface, restricted to just the StatObject function. You can use it in your code for easier unit testing without any external dependencies
type ObjectStatter interface {
	StatObject(bucketName, objectKey string) (s3.ObjectInfo, error)
}

type FakeStatObjectCall struct {
	BucketName string
	ObjectKey  string
}

// FakeObjectStatter is a mock function that can be swapped in for an ObjectStatter, so you can unit test your code
type FakeObjectStatter struct {
	Fn    func(string, string) (s3.ObjectInfo, error)
	Calls []FakeStatObjectCall
}

// PutObject is the interface definition
func (f *FakeObjectStatter) StatObject(bucketName, objectKey string) (s3.ObjectInfo, error) {
	f.Calls = append(f.Calls, FakeStatObjectCall{BucketName: bucketName, ObjectKey: objectKey})
	return f.Fn(bucketName, objectKey)
}

// ObjectPutter is a *(github.com/minio/minio-go).Client compatible interface, restricted to just the PutObject function. You can use it in your code for easier unit testing without any external dependencies
type ObjectPutter interface {
	PutObject(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error)
}

type FakePutObjectCall struct {
	BucketName  string
	ObjectKey   string
	Reader      io.Reader
	ContentType string
}

// FakeObjectPutter is a mock function that can be swapped in for an ObjectPutter, so you can unit test your code
type FakeObjectPutter struct {
	Fn    func(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error)
	Calls []FakePutObjectCall
}

// PutObject is the interface definition
func (f *FakeObjectPutter) PutObject(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error) {
	f.Calls = append(f.Calls, FakePutObjectCall{
		BucketName:  bucketName,
		ObjectKey:   objectKey,
		Reader:      reader,
		ContentType: contentType,
	})
	return f.Fn(bucketName, objectKey, reader, contentType)
}
