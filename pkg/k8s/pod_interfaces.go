package k8s

import (
	"io"
	"time"

	"k8s.io/kubernetes/pkg/api"
	apiunver "k8s.io/kubernetes/pkg/api/unversioned"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

// PodGetter is a *(k8s.io/kubernetes/pkg/client/unversioned).PodInterface compatible interface that only contains the Get func. It is intended for use in getting a single pod
//
// Example usage:
//
//	var pg PodGetter
//	pg = myKubeClient.Pods("myNamespace")
type PodGetter interface {
	Get(name string) (*api.Pod, error)
}

// PodLogStreamer is a *(k8s.io/kubernetes/pkg/client/unversioned).Request compatible interface that only contains the Stream func. It is intended for use in streaming logs, but can be used for other streaming purposes.
//
// Example usage:
//
//	var pls PodLogStreamer
//	pls = StreamerFromKubeClient(myKubeClient, "myNamespace", "myPod")
//	rc, err := pls.Stream()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer rc.Close()
//	io.Copy(os.Stdout, rc)
type PodLogStreamer interface {
	Stream() (io.ReadCloser, error)
}

// LogStreamerFromKubeClient returns a PodLogStreamer using the given *(k8s.io/kubernetes/pkg/client/unversioned).Client. The returned PodLogStreamer will stream logs for the given pod in the given namespace.
// If previous is true, the streamer will include logs even if the pod is terminated. the sinceTime parameter determines the earliest log to show
func LogStreamerFromKubeClient(kc *client.Client, namespace, podName string, previous bool, sinceTime time.Time) PodLogStreamer {
	t := apiunver.NewTime(sinceTime)
	podLogOptions := &api.PodLogOptions{
		Follow:    true,
		Previous:  previous,
		SinceTime: &t,
	}
	return kc.Get().Namespace(namespace).Name(podName).Resource("pods").SubResource("log").VersionedParams(podLogOptions, api.Scheme)
}
