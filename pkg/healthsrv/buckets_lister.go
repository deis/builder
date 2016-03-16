package healthsrv

import "github.com/docker/distribution/context"

// BucketLister is a *(github.com/minio/minio-go).Client compatible interface that provides just
// the ListBuckets cross-section of functionality. It can also be implemented for unit tests.
type BucketLister interface {
	// List returns a list of the objects that are direct descendants of the given path.
	List(ctx context.Context, opath string) ([]string, error)
}

type emptyBucketLister struct{}

func (e emptyBucketLister) ListBuckets(ctx context.Context, opath string) ([]string, error) {
	return nil, nil
}

type errBucketLister struct {
	err error
}

func (e errBucketLister) ListBuckets(ctx context.Context, opath string) ([]string, error) {
	return nil, e.err
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
