package k8s

import (
	"io"

	"k8s.io/kubernetes/pkg/api"
)

// PodCreator is a Kubernetes PodInterface (https://godoc.org/k8s.io/kubernetes/pkg/client/unversioned#PodInterface) compatible interface, strictly for creating pods.
//
// Example usage:
//
//  var pc PodCreator
//  pc = myKubeClient.Pods("myNamespace")
type PodCreator interface {
	Create(pod *api.Pod) (*api.Pod, error)
}

// PodGetter is a Kubernetes PodInterface (https://godoc.org/k8s.io/kubernetes/pkg/client/unversioned#PodInterface) compatible interface, strictly for getting a single pod by name pods.
//
// Example usage:
//
//  var pg PodGetter
//  pg = myKubeClient.Pods("myNamespace")
// type PodGetter interface {
// 	Get(name string) (*api.Pod, error)
// }

// PodLister is a Kubernetes PodInterface (https://godoc.org/k8s.io/kubernetes/pkg/client/unversioned#PodInterface) compatible interface, strictly for listing all pods in a group (usually a namespace).
//
// Example usage:
//
//  var pl PodLister
//  pl = myKubeClient.Pods("myNamespace")
// type PodLister interface {
// 	List(opts api.ListOptions) (*api.PodList, error)
// }
//
// type PodGetterLister interface {
// 	PodGetter
// 	PodLister
// }

// PodCreatorGetterLister is a compound interface composed of Pod{Creator,Getter,Lister} (as the name implies). Since all of the composed interfaces adhere to (k8s.io/kubernetes/pkg/api).PodInterface, this does too.
//
// Example usage:
//
//  var pcgl PodCreatorGetterLister
//  pcgl = myKubeClient.Pods("myNamespace")
// type PodCreatorGetterLister interface {
// 	PodCreator
// 	PodGetter
// 	PodLister
// }

type PodLogStreamer interface {
	Stream() (io.ReadCloser, error)
}
