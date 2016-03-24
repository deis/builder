package storage

import (
	"github.com/docker/distribution/context"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
)

// ObjectStatter is a *(github.com/docker/distribution/registry/storage/driver).StorageDriver compatible interface, restricted to
// just the StatObject function. You can use it in your code for easier unit testing without
// any external dependencies (like access to S3).
type ObjectStatter interface {
	Stat(ctx context.Context, path string) (storagedriver.FileInfo, error)
}

// FakeStatObjectCall represents a single call to StatObject on the FakeObjectStatter.
type FakeStatObjectCall struct {
	Path string
}

// FakeObjectStatter is a mock function that can be swapped in for an ObjectStatter, so you can
// unit test your code.
type FakeObjectStatter struct {
	Fn    func(context.Context, string) (storagedriver.FileInfo, error)
	Calls []FakeStatObjectCall
}

// Stat is the interface definition.
func (f *FakeObjectStatter) Stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	f.Calls = append(f.Calls, FakeStatObjectCall{Path: path})
	return f.Fn(ctx, path)
}

// ObjectGetter is a *(github.com/docker/distribution/registry/storage/driver).StorageDriver compatible interface, restricted to
// just the GetContent function. You can use it in your code for easier unit testing without
// any external dependencies (like access to S3).
type ObjectGetter interface {
	GetContent(ctx context.Context, path string) ([]byte, error)
}

// FakeGetObjectCall represents a single call to GetContent on the FakeObjectStatter.
type FakeGetObjectCall struct {
	Path string
}

// FakeObjectGetter is a mock function that can be swapped in for an ObjectGetter, so you can
// unit test your code.
type FakeObjectGetter struct {
	Fn    func(context.Context, string) ([]byte, error)
	Calls []FakeGetObjectCall
}

// GetContent is the interface definition.
func (f *FakeObjectGetter) GetContent(ctx context.Context, path string) ([]byte, error) {
	f.Calls = append(f.Calls, FakeGetObjectCall{Path: path})
	return f.Fn(ctx, path)
}
