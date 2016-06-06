package k8s

import (
	"k8s.io/kubernetes/pkg/api"
)

// NamespaceLister is a (k8s.io/kubernetes/pkg/client/unversioned).NamespaceInterface compatible
// interface which only has the List function. It's used in places that only need List to make
// them easier to test and more easily swappable with other implementations
// (should the need arise).
//
// Example usage:
//
//	var nsl NamespaceLister
//	nsl = kubeClient.Namespaces()
type NamespaceLister interface {
	List(opts api.ListOptions) (*api.NamespaceList, error)
}
