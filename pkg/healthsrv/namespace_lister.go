package healthsrv

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

// NamespaceLister is an (*k8s.io/kubernetes/pkg/client/unversioned).Client compatible interface that provides just the ListBuckets cross-section of functionality. It can also be implemented for unit tests.
type NamespaceLister interface {
	// List lists all namespaces that are selected by the given label and field selectors
	List(labels.Selector, fields.Selector) (*api.NamespaceList, error)
}

type emptyNamespaceLister struct{}

func (n noNamespacesLister) List(labels.Selector, fields.Selector) (*api.NamespaceList, error) {
	return &api.NamespaceList{}, nil
}
