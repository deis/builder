package storage

import (
	"errors"
	"fmt"
	"time"

	storagedriver "github.com/deis/builder/pkg/storage/driver"
)

const (
	noSuchKeyCode = "NoSuchKey"

	octetStream = "application/octet-stream"
)

// ObjectExists determines whether the object in ${bucketName}/${objKey} exists, as reported by
// statter. Returns the following:
//
// - false, nil if statter succeeded and reported the object doesn't exist
// - false, err with the appropriate error if the statter failed
// - true, nil if the statter succeeded and reported the object exists
func ObjectExists(storageDriver storagedriver.StorageDriver, objKey string) (bool, error) {
	_, err := storageDriver.Stat(objKey)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// WaitForObject checks statter for the object at ${bucketName}/${objKey} right away, then at
// every tick, then once when the timeout is up.
// Returns nil if it finds the object before or at timeout. Otherwise returns a non-nil error.
func WaitForObject(storageDriver storagedriver.StorageDriver, objKey string, tick, timeout time.Duration) error {
	noExist := errors.New("object doesn't exist")
	checker := func() error {
		exists, err := ObjectExists(storageDriver, objKey)
		if err != nil {
			return err
		} else if exists {
			return nil
		} else {
			return noExist
		}
	}
	if err := checker(); err == nil {
		return nil
	}
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ticker.C:
			if err := checker(); err == nil {
				return nil
			}
		case <-timer.C:
			if err := checker(); err == nil {
				return nil
			}
			return fmt.Errorf("Object %s didn't exist after %s", objKey, timeout)
		}
	}
}
