package k8s

import (
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"
)

var (
	resyncPeriod = 30 * time.Second
)

//PodWatcher is a struct which holds the return values of (k8s.io/kubernetes/pkg/controller/framework).NewIndexerInformer together.
type PodWatcher struct {
	Store      cache.StoreToPodLister
	Controller *framework.Controller
}

//NewPodWatcher creates a new BuildPodWatcher useful to list the pods using a cache which gets updated based on the watch func.
func NewPodWatcher(c *client.Client, ns string) *PodWatcher {
	pw := &PodWatcher{}

	pw.Store.Store, pw.Controller = framework.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc:  podListFunc(c, ns),
			WatchFunc: podWatchFunc(c, ns),
		},
		&api.Pod{},
		resyncPeriod,
		framework.ResourceEventHandlerFuncs{},
		cache.Indexers{},
	)

	return pw
}

func podListFunc(c *client.Client, ns string) func(options api.ListOptions) (runtime.Object, error) {
	return func(opts api.ListOptions) (runtime.Object, error) {
		return c.Pods(ns).List(api.ListOptions{
			LabelSelector: labels.Everything(),
		})
	}
}

func podWatchFunc(c *client.Client, ns string) func(options api.ListOptions) (watch.Interface, error) {
	return func(opts api.ListOptions) (watch.Interface, error) {
		return c.Pods(ns).Watch(api.ListOptions{
			LabelSelector: labels.Everything(),
		})
	}
}
