package k8s

import (
	"fmt"
	"reflect"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"
)

func streamEvents(watcher watch.Interface, ch chan<- watch.Event, filter func(watch.Event) bool) {
	for evt := range watcher.ResultChan() {
		if filter(evt) {
			ch <- evt
		}
	}
	close(ch)
}

func addedDeletedFilter(evt watch.Event) bool {
	return evt.Type == watch.Added || evt.Type == watch.Deleted
}

func CreateAndStreamLogs(
	creator PodCreator,
	logStreamer PodLogStreamer,
	watcher watch.Interface,
	pod *api.Pod,
	timeout time.Duration,
	evtsCh chan<- watch.Event,
	logsCh chan<- io.ReadCloser,
	errCh chan<- error,
) {
	timeoutCh := time.After(timeout)

	go streamEvents(watcher, evtsCh, addedDeletedFilter)

	select {
	case evt <- evtsCh:
		evt <- evtsCh
		if evt.Type != watch.Added {
			errCh <- errUnexpectedEventType{podName: pod.Name, evtType: evt.Type}
		}

	}

	// create pod and make sure it was in the Added state
	newPod, err := creator.Create(pod)
	if err != nil {
		return err
	}

	select {
	case <-addedCh:
	case err := <-waitErrCh:
		return nil, err
	case <-timeoutCh:
		return nil, errPodEventTypeTimeout{podName: pod.Name, evtType: watch.Added, timeout: timeout}
	}

	rc, err := PodLogStreamer.Stream()
	if err != nil {
		return nil, err
	}
	req := kubeClient.Get().Namespace(conf.PodNamespace).Name(buildPodName).Resource("pods").SubResource("log").VersionedParams(
		&api.PodLogOptions{
			Follow: true,
		}, api.Scheme)

	rc, err := req.Stream()
	if err != nil {
		return fmt.Errorf("attempting to stream logs (%s)", err)
	}
	defer rc.Close()

}
