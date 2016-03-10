package base

import storagedriver "github.com/deis/builder/pkg/storage/driver"

// Base provides a wrapper around a storagedriver implementation that provides
// common path and bounds checking.
type Base struct {
	storagedriver.StorageDriver
}

// Format errors received from the storage driver
func (base *Base) setDriverName(e error) error {
	switch actual := e.(type) {
	case nil:
		return nil
	case storagedriver.ErrUnsupportedMethod:
		actual.DriverName = base.StorageDriver.Name()
		return actual
	case storagedriver.PathNotFoundError:
		actual.DriverName = base.StorageDriver.Name()
		return actual
	case storagedriver.InvalidPathError:
		actual.DriverName = base.StorageDriver.Name()
		return actual
	case storagedriver.InvalidOffsetError:
		actual.DriverName = base.StorageDriver.Name()
		return actual
	default:
		storageError := storagedriver.Error{
			DriverName: base.StorageDriver.Name(),
			Enclosed:   e,
		}

		return storageError
	}
}

// CheckConnectionStatus wraps CheckConnectionStatus of underlying storage driver.
func (base *Base) CheckConnectionStatus() (bool, error) {

	b, e := base.StorageDriver.CheckConnectionStatus()
	return b, e
}

// PutContent wraps PutContent of underlying storage driver.
func (base *Base) PutContent(path string, content []byte) error {

	return base.setDriverName(base.StorageDriver.PutContent(path, content))
}

// GetContent wraps GetContent of underlying storage driver.
func (base *Base) GetContent(path string) ([]byte, error) {

	b, e := base.StorageDriver.GetContent(path)
	return b, base.setDriverName(e)
}

// Stat wraps Stat of underlying storage driver.
func (base *Base) Stat(path string) (storagedriver.FileInfo, error) {
	fi, e := base.StorageDriver.Stat(path)
	return fi, base.setDriverName(e)
}
