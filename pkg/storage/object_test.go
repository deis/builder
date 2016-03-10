package storage

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

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

func TestObjectExistsOtherErr(t *testing.T) {
	expectedErr := errors.New("other error")
	statter := &FakeObjectStatter{
		Fn: func(string, string) (s3.ObjectInfo, error) {
			return s3.ObjectInfo{}, expectedErr
		},
	}
	exists, err := ObjectExists(statter, bucketName, objKey)
	assert.Err(t, err, expectedErr)
	assert.False(t, exists, "object found when the statter errored")
}

func TestUploadObjectSuccess(t *testing.T) {
	rdr := strings.NewReader("hello world!")
	putter := &FakeObjectPutter{
		Fn: func(string, string, io.Reader, string) (int64, error) {
			return 0, nil
		},
	}
	assert.NoErr(t, UploadObject(putter, bucketName, objKey, rdr))
	assert.Equal(t, len(putter.Calls), 1, "number of calls to PutObject")
	assert.Equal(t, putter.Calls[0].BucketName, bucketName, "the bucket name")
	assert.Equal(t, putter.Calls[0].ObjectKey, objKey, "the object key")
	assert.Equal(t, putter.Calls[0].ContentType, octetStream, "the content type")
}

func TestUploadObjectFailure(t *testing.T) {
	rdr := strings.NewReader("hello world")
	err := errors.New("test error")
	putter := &FakeObjectPutter{
		Fn: func(string, string, io.Reader, string) (int64, error) {
			return 0, err
		},
	}
	assert.Err(t, UploadObject(putter, bucketName, objKey, rdr), err)
	assert.Equal(t, len(putter.Calls), 1, "number of calls to PutObject")
}

func TestWaitForObjectMissing(t *testing.T) {
	// Skipped for now. See https://github.com/deis/builder/issues/238
	t.SkipNow()
	statter := &FakeObjectStatter{
		Fn: func(string, string) (s3.ObjectInfo, error) {
			return s3.ObjectInfo{}, s3.ErrorResponse{Code: noSuchKeyCode}
		},
	}
	err := WaitForObject(statter, bucketName, objKey, 1*time.Millisecond, 2*time.Millisecond)
	assert.True(t, err != nil, "no error received when there should have been")
	// it should make 1 call immediately, then calls at 1ms and 2ms
	assert.True(
		t,
		len(statter.Calls) >= 1,
		"the statter was not called, but should have been called at least once",
	)
}

func TestWaitForObjectExists(t *testing.T) {
	statter := &FakeObjectStatter{
		Fn: func(string, string) (s3.ObjectInfo, error) {
			return s3.ObjectInfo{}, nil
		},
	}
	assert.NoErr(t, WaitForObject(statter, bucketName, objKey, 1*time.Millisecond, 2*time.Millisecond))
	// it should make 1 call immediately, then immediateley succeed
	assert.Equal(t, len(statter.Calls), 1, "number of calls to the statter")
}

func TestDownloadObjectSuccess(t *testing.T) {
	obj := &FakeObject{Data: "web: example-go"}
	getter := &FakeObjectGetter{
		Fn: func(string, string) (Object, error) {
			return obj, nil
		},
	}
	data, err := DownloadObject(getter, bucketName, objKey)
	assert.NoErr(t, err)
	assert.Equal(t, string(data), "web: example-go", "data")
	assert.Equal(t, len(getter.Calls), 1, "number of calls to GetObject")
	assert.Equal(t, getter.Calls[0].BucketName, bucketName, "the bucket name")
	assert.Equal(t, getter.Calls[0].ObjectKey, objKey, "the object key")
}
