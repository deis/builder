package healthsrv

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

// NamespaceLister is an (*k8s.io/kubernetes/pkg/client/unversioned).Client compatible interface
// that provides just the ListBuckets cross-section of functionality. It can also be implemented
// for unit tests.
type NamespaceLister interface {
	// List lists all namespaces that are selected by the given label and field selectors.
	List(labels.Selector, fields.Selector) (*api.NamespaceList, error)
}

type emptyNamespaceLister struct{}

func (n emptyNamespaceLister) List(labels.Selector, fields.Selector) (*api.NamespaceList, error) {
	return &api.NamespaceList{}, nil
}

type errNamespaceLister struct {
	err error
}

func (e errNamespaceLister) List(labels.Selector, fields.Selector) (*api.NamespaceList, error) {
	return nil, e.err
}

// listNamespaces calls nl.List(...) and sends the results back on the various given channels.
// This func is intended to be run in a goroutine and communicates via the channels it's passed.
//
// On success, it passes the namespace list on succCh, and on failure, it passes the error on
// errCh. At most one of {succCh, errCh} will be sent on. If stopCh is closed, no pending or
// future sends will occur.
func listNamespaces(nl NamespaceLister, succCh chan<- *api.NamespaceList, errCh chan<- error, stopCh <-chan struct{}) {
	nsList, err := nl.List(labels.Everything(), fields.Everything())
	if err != nil {
		select {
		case errCh <- err:
		case <-stopCh:
		}
		return
	}
	select {
	case succCh <- nsList:
	case <-stopCh:
		return
	}
}
