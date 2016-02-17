package k8s

import (
	"fmt"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"
)

type errPodAddedTimeout struct {
	pod     *api.Pod
	timeout time.Duration
}

func (e errPodCreateTimeout) Error() string {
	return fmt.Sprintf("pod %s timed out creating after %s", e.pod.Name, e.timeout)
}

type errUnexpectedEventType struct {
	podName string
	evtType watch.EventType
}

func (e errUnexpectedEventType) Error() string {
	return fmt.Sprintf("pod %s had unexpected event type %s", e.podName, e.evtType)
}

type errPodEventTypeTimeout struct {
	podName string
	evtType watch.EventType
	timeout time.Duration
}

func (e errPodEventStautsTimeout) Error() string {
	return fmt.Sprintf("pod %s wasn't %s within %s", e.podName, e.evtType, e.timeout)
}

type errEventTypeNotFound struct {
	t watch.EventType
}

func (e errEventTypeNotFound) Error() string {
	return fmt.Sprintf("event type %s not found", e.t)
}
