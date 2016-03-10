package sys

import (
	"fmt"
	"io/ioutil"
	"os"
)

// FS is the interface to a file system.
type FS interface {
	// ReadAll gets the contents of filename, or an error if the file didn't exist or there was an
	// error reading it.
	ReadFile(filename string) ([]byte, error)

	// RemoveAll removes all of the files under the directory at name. It behaves similarly to the
	// func of the same name in the os package (https://godoc.org/os#RemoveAll).
	RemoveAll(name string) error
}

// RealFS returns an FS object that interacts with the real local filesystem.
func RealFS() FS {
	return &realFS{}
}

type realFS struct{}

// ReadFile is the interface implementation for FS.
func (r *realFS) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

// RemoveAll is the interface implementation for FS.
func (r *realFS) RemoveAll(name string) error {
	return os.RemoveAll(name)
}

// FakeFileNotFound is the error returned by FakeFS when a requested file isn't found.
type FakeFileNotFound struct {
	Filename string
}

// Error is the error interface implementation.
func (f FakeFileNotFound) Error() string {
	return fmt.Sprintf("Fake file %s not found", f.Filename)
}

// FakeFS is an in-memory FS implementation.
type FakeFS struct {
	Files map[string][]byte
}

// NewFakeFS returns a FakeFS with no files.
func NewFakeFS() *FakeFS {
	return &FakeFS{Files: make(map[string][]byte)}
}

// ReadFile is the FS interface implementation. It returns FakeFileNotFound if the file was not
// found in the in-memory 'filesystem' of f.
func (f *FakeFS) ReadFile(name string) ([]byte, error) {
	b, ok := f.Files[name]
	if !ok {
		return nil, FakeFileNotFound{Filename: name}
	}
	return b, nil
}

// RemoveAll is the interface implementation for FS.
func (f *FakeFS) RemoveAll(name string) error {
	_, ok := f.Files[name]
	if !ok {
		return FakeFileNotFound{Filename: name}
	}
	delete(f.Files, name)
	return nil
}
