package healthsrv

import (
	s3 "github.com/minio/minio-go"
)

// BucketLister is a *(github.com/aws/aws-sdk-go/service/s3).Client compatible interface that provides just the ListBuckets cross-section of functionality. It can also be implemented for unit tests
type BucketLister interface {
	// ListBuckets lists all the buckets in the object storage system
	ListBuckets() ([]s3.BucketInfo, error)
}

type emptyBucketLister struct{}

func (e emptyBucketLister) ListBuckets() ([]s3.BucketInfo, error) {
	return nil, nil
}

type errBucketLister struct {
	err error
}

func (e errBucketLister) ListBuckets() ([]s3.BucketInfo, error) {
	return nil, e.err
}

// listBuckets calls bl.ListBuckets(...) and sends the results back on the various given channels. This func is intended to be run in a goroutine and communicates via the channels it's passed.
//
// On success, it passes the bucket output on succCh, and on failure, it passes the error on errCh. At most one of {succCh, errCh} will be sent on. If stopCh is closed, no pending or future sends will occur.
func listBuckets(bl BucketLister, succCh chan<- []s3.BucketInfo, errCh chan<- error, stopCh <-chan struct{}) {
	lbOut, err := bl.ListBuckets()
	if err != nil {
		select {
		case errCh <- err:
		case <-stopCh:
		}
		return
	}
	select {
	case succCh <- lbOut:
	case <-stopCh:
	}
}
