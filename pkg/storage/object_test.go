package storage

import (
	"errors"
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/docker/distribution/context"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
)

const (
	objPath = "myobj"
)

func TestObjectExistsSuccess(t *testing.T) {
	objInfo := storagedriver.FileInfoInternal{storagedriver.FileInfoFields{Path: objPath, Size: 1234}}
	statter := &FakeObjectStatter{
		Fn: func(context.Context, string) (storagedriver.FileInfo, error) {
			return objInfo, nil
		},
	}
	exists, err := ObjectExists(statter, objPath)
	assert.NoErr(t, err)
	assert.True(t, exists, "object not found when it should be present")
	assert.Equal(t, len(statter.Calls), 1, "number of StatObject calls")
	assert.Equal(t, statter.Calls[0].Path, objPath, "object key")
}

func TestObjectExistsNoObject(t *testing.T) {
	statter := &FakeObjectStatter{
		Fn: func(context.Context, string) (storagedriver.FileInfo, error) {
			return storagedriver.FileInfoInternal{FileInfoFields: storagedriver.FileInfoFields{}}, storagedriver.PathNotFoundError{Path: objPath}
		},
	}
	exists, err := ObjectExists(statter, objPath)
	assert.NoErr(t, err)
	assert.False(t, exists, "object found when it should be missing")
	assert.Equal(t, len(statter.Calls), 1, "number of StatObject calls")
}

func TestObjectExistsOtherErr(t *testing.T) {
	expectedErr := errors.New("other error")
	statter := &FakeObjectStatter{
		Fn: func(context.Context, string) (storagedriver.FileInfo, error) {
			return storagedriver.FileInfoInternal{FileInfoFields: storagedriver.FileInfoFields{}}, expectedErr
		},
	}
	exists, err := ObjectExists(statter, objPath)
	assert.Err(t, err, expectedErr)
	assert.False(t, exists, "object found when the statter errored")
}

func TestWaitForObjectMissing(t *testing.T) {
	// Skipped for now. See https://github.com/deis/builder/issues/238
	t.SkipNow()
	statter := &FakeObjectStatter{
		Fn: func(context.Context, string) (storagedriver.FileInfo, error) {
			return storagedriver.FileInfoInternal{FileInfoFields: storagedriver.FileInfoFields{}}, storagedriver.PathNotFoundError{Path: objPath}
		},
	}
	err := WaitForObject(statter, objPath, 1*time.Millisecond, 2*time.Millisecond)
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
		Fn: func(context.Context, string) (storagedriver.FileInfo, error) {
			return storagedriver.FileInfoInternal{FileInfoFields: storagedriver.FileInfoFields{}}, nil
		},
	}
	assert.NoErr(t, WaitForObject(statter, objPath, 1*time.Millisecond, 2*time.Millisecond))
	// it should make 1 call immediately, then immediateley succeed
	assert.Equal(t, len(statter.Calls), 1, "number of calls to the statter")
}
