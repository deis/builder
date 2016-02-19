package k8s

import (
	"errors"
	"time"

	"k8s.io/kubernetes/pkg/api"
	apierrs "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/util/wait"
)

var (
	ErrPodInFailedState           = errors.New("Pod is in the Failed state")
	ErrIntervalGreaterThanTimeout = errors.New("Given interval is greater than given timeout")
)

// ConditionFunc is called in WaitForPodCondition on every successful poll event. Given a pod, it should return true if the condition is true and polling should stop successfully, false if polling should continue and an error if polling should stop with failure
type ConditionFunc func(*api.Pod) (bool, error)

// PodGetter is the abstraction used by many pod specific functions to get a pod. Example usage with the standard *(k8s.io/kubernetes/pkg/client/unversioend).Client:
//
//  var pg PodGetter
//  pg = myKubeClient.Pods("myNamespace").Get
type PodGetter func(string) (*api.Pod, error)

// WaitForPodCondition polls the state of a pod every interval until timeout. Each time it polls, it does the following:
// - Call podGetter(podName) to get the pod, and return with an error if the pod was not found
// - If the pod was found, return directly the result of condition(pod)
func WaitForPodCondition(getPod PodGetter, podName string, condition ConditionFunc, interval, timeout time.Duration) error {
	if interval > timeout {
		return ErrIntervalGreaterThanTimeout
	}
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		pod, err := getPod(podName)
		if err != nil {
			if apierrs.IsNotFound(err) {
				return true, err
			}
		}

		return condition(pod)
	})
}

// WaitForPodRunning waits for the pod with the given name to go into the Running state. Returns ErrPodInFailedState if the pod goes into the Failed state. Any other error means that the wait failed. In this case, the pod is not necessarily running on return.
func WaitForPodRunning(getPod PodGetter, podName string, interval, timeout time.Duration) error {
	condition := func(pod *api.Pod) (bool, error) {
		if pod.Status.Phase == api.PodRunning {
			return true, nil
		}
		if pod.Status.Phase == api.PodFailed {
			return true, ErrPodInFailedState
		}
		return false, nil
	}

	return WaitForPodCondition(getPod, podName, condition, interval, timeout)
}

// WaitForPodEnded waits for the a pod with the given name to report Succeeded or Failed state. It returns with an error if the waiting failed for any reason, so obviously any non-nil error means that the pod has not necessarily ended
func WaitForPodEnded(getPod PodGetter, podName string, interval, timeout time.Duration) error {
	condition := func(pod *api.Pod) (bool, error) {
		phase := pod.Status.Phase
		return (phase == api.PodSucceeded || phase == api.PodFailed), nil
	}
	return WaitForPodCondition(getPod, podName, condition, interval, timeout)
}
