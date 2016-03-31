package gitreceive

import (
	"fmt"
	"time"

	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
	apierrs "k8s.io/kubernetes/pkg/api/errors"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/wait"
)

const (
	slugBuilderName   = "deis-slugbuilder"
	dockerBuilderName = "deis-dockerbuilder"

	tarPath          = "TAR_PATH"
	putPath          = "PUT_PATH"
	debugKey         = "DEBUG"
	objectStore      = "objectstorage-keyfile"
	dockerSocketName = "docker-socket"
	dockerSocketPath = "/var/run/docker.sock"
	builderStorage   = "BUILDER_STORAGE"
	objectStorePath  = "/var/run/secrets/deis/objectstore/creds"
)

func dockerBuilderPodName(appName, shortSha string) string {
	uid := uuid.New()[:8]
	return fmt.Sprintf("dockerbuild-%s-%s-%s", appName, shortSha, uid)
}

func slugBuilderPodName(appName, shortSha string) string {
	uid := uuid.New()[:8]
	return fmt.Sprintf("slugbuild-%s-%s-%s", appName, shortSha, uid)
}

func dockerBuilderPod(debug bool, name, namespace string, env map[string]interface{}, tarKey, imageName, dockerBuilderImage, storageType string) *api.Pod {
	pod := buildPod(debug, name, namespace, env)

	pod.Spec.Containers[0].Name = dockerBuilderName
	pod.Spec.Containers[0].Image = dockerBuilderImage

	addEnvToPod(pod, tarPath, tarKey)
	addEnvToPod(pod, "IMG_NAME", imageName)
	addEnvToPod(pod, builderStorage, storageType)

	pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, api.VolumeMount{
		Name:      dockerSocketName,
		MountPath: dockerSocketPath,
	})

	pod.Spec.Volumes = append(pod.Spec.Volumes, api.Volume{
		Name: dockerSocketName,
		VolumeSource: api.VolumeSource{
			HostPath: &api.HostPathVolumeSource{
				Path: dockerSocketPath,
			},
		},
	})

	return &pod
}

func slugbuilderPod(debug bool, name, namespace string, env map[string]interface{}, tarKey, putKey, buildpackURL, slugBuilderImage, storageType string) *api.Pod {
	pod := buildPod(debug, name, namespace, env)

	pod.Spec.Containers[0].Name = slugBuilderName
	pod.Spec.Containers[0].Image = slugBuilderImage

	addEnvToPod(pod, tarPath, tarKey)
	addEnvToPod(pod, putPath, putKey)
	addEnvToPod(pod, builderStorage, storageType)

	if buildpackURL != "" {
		addEnvToPod(pod, "BUILDPACK_URL", buildpackURL)
	}

	return &pod
}

func buildPod(debug bool, name, namespace string, env map[string]interface{}) api.Pod {
	pod := api.Pod{
		Spec: api.PodSpec{
			RestartPolicy: api.RestartPolicyNever,
			Containers: []api.Container{
				api.Container{
					ImagePullPolicy: api.PullAlways,
				},
			},
			Volumes: []api.Volume{},
		},
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"heritage": name,
			},
		},
	}

	pod.Spec.Volumes = append(pod.Spec.Volumes, api.Volume{
		Name: objectStore,
		VolumeSource: api.VolumeSource{
			Secret: &api.SecretVolumeSource{
				SecretName: objectStore,
			},
		},
	})

	pod.Spec.Containers[0].VolumeMounts = []api.VolumeMount{
		api.VolumeMount{
			Name:      objectStore,
			MountPath: objectStorePath,
			ReadOnly:  true,
		},
	}

	if len(pod.Spec.Containers) > 0 {
		for k, v := range env {
			pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, api.EnvVar{
				Name:  k,
				Value: fmt.Sprintf("%v", v),
			})
		}
	}

	if debug {
		addEnvToPod(pod, debugKey, "1")
	}

	return pod
}

func addEnvToPod(pod api.Pod, key, value string) {
	if len(pod.Spec.Containers) > 0 {
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, api.EnvVar{
			Name:  key,
			Value: value,
		})
	}
}

// waitForPod waits for a pod in state running or failed
func waitForPod(c *client.Client, ns, podName string, interval, timeout time.Duration) error {
	condition := func(pod *api.Pod) (bool, error) {
		if pod.Status.Phase == api.PodRunning {
			return true, nil
		}
		if pod.Status.Phase == api.PodFailed {
			return true, fmt.Errorf("Giving up; pod went into failed status: \n%s", fmt.Sprintf("%#v", pod))
		}
		return false, nil
	}

	return waitForPodCondition(c, ns, podName, condition, interval, timeout)
}

// waitForPodEnd waits for a pod in state succeeded or failed
func waitForPodEnd(c *client.Client, ns, podName string, interval, timeout time.Duration) error {
	condition := func(pod *api.Pod) (bool, error) {
		if pod.Status.Phase == api.PodSucceeded {
			return true, nil
		}
		if pod.Status.Phase == api.PodFailed {
			return true, nil
		}
		return false, nil
	}

	return waitForPodCondition(c, ns, podName, condition, interval, timeout)
}

// waitForPodCondition waits for a pod in state defined by a condition (func)
func waitForPodCondition(c *client.Client, ns, podName string, condition func(pod *api.Pod) (bool, error),
	interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		pod, err := c.Pods(ns).Get(podName)
		if err != nil {
			if apierrs.IsNotFound(err) {
				return false, nil
			}
		}

		done, err := condition(pod)
		if err != nil {
			return false, err
		}
		if done {
			return true, nil
		}

		return false, nil
	})
}
