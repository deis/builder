package driver

import "fmt"

// StorageDriver defines methods that a Storage Driver must implement for a
// filesystem-like key/value object storage.
type StorageDriver interface {
	// Name returns the human-readable "name" of the driver, useful in error
	// messages and logging. By convention, this will just be the registration
	// name, but drivers may provide other information here.
	Name() string

	//CheckConnectionStatus return whether the connection to the storage is still active or not
	CheckConnectionStatus() (bool, error)

	// PutContent stores the []byte content at a location designated by "path".
	// This should primarily be used for small objects.
	PutContent(path string, content []byte) error

	// GetContent retrieves the content stored at "path" as a []byte.
	// This should primarily be used for small objects.
	GetContent(path string) ([]byte, error)

	// Stat retrieves the FileInfo for the given path, including the current
	// size in bytes and the creation time.
	Stat(path string) (FileInfo, error)
}

// ErrUnsupportedMethod may be returned in the case where a StorageDriver implementation does not support an optional method.
type ErrUnsupportedMethod struct {
	DriverName string
}

func (err ErrUnsupportedMethod) Error() string {
	return fmt.Sprintf("%s: unsupported method", err.DriverName)
}

// PathNotFoundError is returned when operating on a nonexistent path.
type PathNotFoundError struct {
	Path       string
	DriverName string
}

func (err PathNotFoundError) Error() string {
	return fmt.Sprintf("%s: Path not found: %s", err.DriverName, err.Path)
}

// InvalidPathError is returned when the provided path is malformed.
type InvalidPathError struct {
	Path       string
	DriverName string
}

func (err InvalidPathError) Error() string {
	return fmt.Sprintf("%s: invalid path: %s", err.DriverName, err.Path)
}

// InvalidOffsetError is returned when attempting to read or write from an
// invalid offset.
type InvalidOffsetError struct {
	Path       string
	Offset     int64
	DriverName string
}

func (err InvalidOffsetError) Error() string {
	return fmt.Sprintf("%s: invalid offset: %d for path: %s", err.DriverName, err.Offset, err.Path)
}

// Error is a catch-all error type which captures an error string and
// the driver type on which it occurred.
type Error struct {
	DriverName string
	Enclosed   error
}

func (err Error) Error() string {
	return fmt.Sprintf("%s: %s", err.DriverName, err.Enclosed)
}
