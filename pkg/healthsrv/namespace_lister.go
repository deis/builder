package healthsrv

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

type NamespaceLister interface {
	List(labels.Selector, fields.Selector) (*api.NamespaceList, error)
}
