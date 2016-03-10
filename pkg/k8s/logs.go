package k8s

import (
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	podResource = "pods"
	logResource = "log"
)

func LogStreamRequest(cl *client.Client, namespace, podName string) *client.Request {
	return cl.Get().
		Namespace(namespace).
		Name(podName).
		Resource(podResource).
		SubResource(logResource).
		VersionedParams(&api.PodLogOptions{Follow: true}, api.Scheme)
}
