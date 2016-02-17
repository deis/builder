package k8s

// waitForPodCondition waits for a pod in state defined by a condition (func).
func waitForPodCondition(podGetter PodGetter, podName string, interval, timeout time.Duration) error {
	podInterface := c.Pods(ns)
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		podList, err := podInterface.List(labels.Everything(), fields.Everything())
		// error listing pods
		if err != nil {
			return false, err
		}
		pod := findPod(podName, podList.Items)
		// pod not found
		if pod == nil {
			return false, nil
		}

		done, err := condition(pod)
		if done {
			return true, nil
		}

		return false, nil
	})
}
