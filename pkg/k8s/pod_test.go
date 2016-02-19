package k8s

import (
	"testing"
	"time"

	"github.com/arschles/assert"
	"k8s.io/kubernetes/pkg/api"
)

const (
	podName = "myPod"
)

var nilPodGetter = func(string) (*api.Pod, error) {
	return nil, nil
}

var alwaysTrueCondition = func(*api.Pod) (bool, error) {
	return true, nil
}

func TestWaitForPodCondition(t *testing.T) {
	err := WaitForPodCondition(nilPodGetter, podName, alwaysTrueCondition, 2*time.Millisecond, 1*time.Millisecond)
	assert.Err(t, ErrIntervalGreaterThanTimeout, err)
	assert.NoErr(t, WaitForPodCondition(nilPodGetter, podName, alwaysTrueCondition, 1*time.Millisecond, 2*time.Millisecond))
}

func TestWaitForPodRunning(t *testing.T) {
	t.Skip("TODO")
}

func TestWaitForPodEnded(t *testing.T) {
	t.Skip("TODO")
}
