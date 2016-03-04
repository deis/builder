package sys

import (
	"fmt"
	"io/ioutil"
)

// FS is the interface to a file system
type FS interface {
	// ReadAll gets the contents of filename, or an error if the file didn't exist or there was an error reading it
	ReadFile(filename string) ([]byte, error)
}

// RealFS returns an FS object that interacts with the real local filesystem
func RealFS() FS {
	return &realFS{}
}

type realFS struct{}

func (r *realFS) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

// FakeFileNotFound is the error returned by FakeFS when a requested file isn't found
type FakeFileNotFound struct {
	Filename string
}

// Error is the error interface implementation
func (f FakeFileNotFound) Error() string {
	return fmt.Sprintf("Fake file %s not found", f.Filename)
}

type FakeFS struct {
	Files map[string][]byte
}

func NewFakeFS() *FakeFS {
	return &FakeFS{Files: make(map[string][]byte)}
}

func (f *FakeFS) ReadFile(name string) ([]byte, error) {
	b, ok := f.Files[name]
	if !ok {
		return nil, FakeFileNotFound{Filename: name}
	}
	return b, nil
}
