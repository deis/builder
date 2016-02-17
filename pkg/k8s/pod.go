package k8s

// PodHasAnyContainerStatus iterates the exit codes of all containers in podName. If fn(containerStatus) returns true, returns true and containerStatus. If none returns true, returns false and 0. On any error, returns false, 0 and the error
func PodHasAnyContainerStatus(podGetter PodGetter, podName string, fn func(int) bool) (bool, int, error) {
	pod, err := podGetter.Get(podName)
	if err != nil {
		return false, 0, err
	}
	for _, containerStatus := range pod.Status.ContainerStatuses {
		state := containerStatus.State.Terminated
		if fn(state.ExitCode) {
			return true, state.ExitCode, nil
		}
	}
	return false, 0, nil
}
