package healthsrv

import (
	"github.com/docker/distribution/context"
)

// BucketLister is a *(github.com/docker/distribution/registry/storage/driver).StorageDriver compatible interface that provides just
// the List cross-section of functionality. It can also be implemented for unit tests.
type BucketLister interface {
	// List returns a list of the objects that are direct descendants of the given path.
	List(ctx context.Context, opath string) ([]string, error)
}

// listBuckets calls bl.ListBuckets(...) and sends the results back on the various given channels.
// This func is intended to be run in a goroutine and communicates via the channels it's passed.
//
// On success, it passes the bucket output on succCh, and on failure, it passes the error on errCh.
// At most one of {succCh, errCh} will be sent on. If stopCh is closed, no pending or future sends
// will occur.
func listBuckets(bl BucketLister, succCh chan<- []string, errCh chan<- error, stopCh <-chan struct{}) {
	lbOut, err := bl.List(context.Background(), "/")
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
