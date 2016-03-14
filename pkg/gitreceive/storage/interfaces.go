package storage

import (
	"io"

	s3 "github.com/minio/minio-go"
)

// BucketCreator is a *(github.com/minio/minio-go).Client compatible interface, restricted to
// just the MakeBucket function. You can use it in your code for easier unit testing without
// any external dependencies.
type BucketCreator interface {
	MakeBucket(bucketName string, acl s3.BucketACL, location string) error
}

// FakeMakeBucketCall represents a single call to MakeBucket on a FakeBucketCreator.
type FakeMakeBucketCall struct {
	BucketName string
	ACL        s3.BucketACL
	Location   string
}

// FakeBucketCreator is a mock function that can be swapped in for an BucketCreator, so you
// can unit test your code.
type FakeBucketCreator struct {
	Fn    func(string, s3.BucketACL, string) error
	Calls []FakeMakeBucketCall
}

// MakeBucket is the interface definition for BucketCreator.
func (f *FakeBucketCreator) MakeBucket(name string, acl s3.BucketACL, location string) error {
	f.Calls = append(f.Calls, FakeMakeBucketCall{BucketName: name, ACL: acl, Location: location})
	return f.Fn(name, acl, location)
}

// ObjectStatter is a *(github.com/minio/minio-go).Client compatible interface, restricted to
// just the StatObject function. You can use it in your code for easier unit testing without
// any external dependencies (like access to S3).
type ObjectStatter interface {
	StatObject(bucketName, objectKey string) (s3.ObjectInfo, error)
}

// FakeStatObjectCall represents a single call to StatObject on the FakeObjectStatter.
type FakeStatObjectCall struct {
	BucketName string
	ObjectKey  string
}

// FakeObjectStatter is a mock function that can be swapped in for an ObjectStatter, so you can
// unit test your code.
type FakeObjectStatter struct {
	Fn    func(string, string) (s3.ObjectInfo, error)
	Calls []FakeStatObjectCall
}

// StatObject is the interface definition.
func (f *FakeObjectStatter) StatObject(bucketName, objectKey string) (s3.ObjectInfo, error) {
	f.Calls = append(f.Calls, FakeStatObjectCall{BucketName: bucketName, ObjectKey: objectKey})
	return f.Fn(bucketName, objectKey)
}

// ObjectPutter is a *(github.com/minio/minio-go).Client compatible interface, restricted to just
// the PutObject function. You can use it in your code for easier unit testing without any
// external dependencies.
type ObjectPutter interface {
	PutObject(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error)
}

// FakePutObjectCall represents a single call to PutObject on a FakeObjectPutter.
type FakePutObjectCall struct {
	BucketName  string
	ObjectKey   string
	Reader      io.Reader
	ContentType string
}

// FakeObjectPutter is a mock function that can be swapped in for an ObjectPutter, so you can
// unit test your code.
type FakeObjectPutter struct {
	Fn    func(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error)
	Calls []FakePutObjectCall
}

// PutObject is the interface definition.
func (f *FakeObjectPutter) PutObject(bucketName, objectKey string, reader io.Reader, contentType string) (int64, error) {
	f.Calls = append(f.Calls, FakePutObjectCall{
		BucketName:  bucketName,
		ObjectKey:   objectKey,
		Reader:      reader,
		ContentType: contentType,
	})
	return f.Fn(bucketName, objectKey, reader, contentType)
}

// Object is a *(github.com/minio/minio-go).Object compatible interface. Currently, ObjectGetter
// returns these so that FakeObjectGetters can return mock implementations.
type Object interface {
	// This is called an interface composition - it automatically gives your interface the function
	// in io.Reader (https://godoc.org/io#Reader) and the function in io.Closer
	// (https://godoc.org/io#Closer).
	io.ReadCloser
	// This is also an interface composition. It gives your interface the function in
	// io.Seeker (https://godoc.org/io#Seeker).
	io.Seeker
	// This is also interface composition. It gives your interface the function in
	// io.ReaderAt (https://godoc.org/io#ReaderAt).
	io.ReaderAt
	// This function is the last one we have to add to make this interface have all the same
	// functions as s3.Object.
	Stat() (s3.ObjectInfo, error)
}

// ObjectGetter is the interface to get an object from object storage. The minio client doesn't
// already satisfy this interface, because the GetObject func
// (https://godoc.org/github.com/minio/minio-go#Client.GetObject) doesn't return an Object.
// Instead, it returns a *s3.Object. Use the RealObjectGetter below to use the minio client.
type ObjectGetter interface {
	// GetObject is *almost* the same function as the GetObject func in the minio client, but it
	// returns Object instead of *s3.Object.
	GetObject(string, string) (Object, error)
}

// RealObjectGetter is an adapter to make the *s3.Client GetObject function compatible with the
// ObjectGetter interface.
type RealObjectGetter struct {
	Client *s3.Client
}

// GetObject is the interface implementation for ObjectGetter.
func (r *RealObjectGetter) GetObject(bucket, objKey string) (Object, error) {
	obj, err := r.Client.GetObject(bucket, objKey)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// FakeGetObjectCall represents a single call a single call to GetObject on a FakeObjectGetter.
type FakeGetObjectCall struct {
	BucketName string
	ObjectKey  string
}

// FakeObjectGetter is a mock function that can be swapped in for an ObjectGetter, so you can
// unit test your code.
type FakeObjectGetter struct {
	Fn    func(string, string) (Object, error)
	Calls []FakeGetObjectCall
}

// GetObject is the interface definition.
func (f *FakeObjectGetter) GetObject(bucketName, objectKey string) (Object, error) {
	f.Calls = append(f.Calls, FakeGetObjectCall{BucketName: bucketName, ObjectKey: objectKey})
	return f.Fn(bucketName, objectKey)
}

// FakeObject is a mock function that can be swapped in for an *s3.Object, so you can unit test
// your code.
type FakeObject struct {
	Data string
}

// Read is the interface definition for Object.
func (f *FakeObject) Read(b []byte) (n int, err error) {
	copy(b, f.Data[:])
	return len(f.Data), io.EOF
}

// Close is the interface definition for Object.
func (f *FakeObject) Close() (err error) {
	return nil
}

// ReadAt is the interface definition for Object.
func (f *FakeObject) ReadAt(b []byte, offset int64) (n int, err error) {
	return 0, nil
}

// Seek is the interface definition for Object.
func (f *FakeObject) Seek(offset int64, whence int) (n int64, err error) {
	return 0, nil
}

// Stat is the interface definition for Object.
func (f *FakeObject) Stat() (s3.ObjectInfo, error) {
	return s3.ObjectInfo{}, nil
}
