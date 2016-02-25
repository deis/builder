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
	slugBuilderName    = "deis-slugbuilder"
	slugBuilderImage   = "quay.io/deisci/slugbuilder:v2-beta"
	dockerBuilderName  = "deis-dockerbuilder"
	dockerBuilderImage = "quay.io/deisci/dockerbuilder:v2-beta"

	tarURLKey        = "TAR_URL"
	putURLKey        = "put_url"
	debugKey         = "DEBUG"
	minioUser        = "minio-user"
	dockerSocketName = "docker-socket"
	dockerSocketPath = "/var/run/docker.sock"
)

func dockerBuilderPodName(appName, shortSha string) string {
	uid := uuid.New()[:8]
	return fmt.Sprintf("dockerbuild-%s-%s-%s", appName, shortSha, uid)
}

func slugBuilderPodName(appName, shortSha string) string {
	uid := uuid.New()[:8]
	return fmt.Sprintf("slugbuild-%s-%s-%s", appName, shortSha, uid)
}

func dockerBuilderPod(debug, withAuth bool, name, namespace string, env map[string]interface{}, tarURL, imageName, region string) *api.Pod {
	pod := buildPod(debug, withAuth, name, namespace, env)

	pod.Spec.Containers[0].Name = dockerBuilderName
	pod.Spec.Containers[0].Image = dockerBuilderImage

	addEnvToPod(pod, "ACCESS_KEY_FILE", "/var/run/secrets/object/store/access_key")
	addEnvToPod(pod, "ACCESS_SECRET_FILE", "/var/run/secrets/object/store/access_secret")

	addEnvToPod(pod, tarURLKey, tarURL)
	addEnvToPod(pod, "IMG_NAME", imageName)
	addEnvToPod(pod, "REGION", region)

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

func slugbuilderPod(debug, withAuth bool, name, namespace string, env map[string]interface{}, tarURL, putURL, buildpackURL string) *api.Pod {
	pod := buildPod(debug, withAuth, name, namespace, env)

	pod.Spec.Containers[0].Name = slugBuilderName
	pod.Spec.Containers[0].Image = slugBuilderImage

	addEnvToPod(pod, tarURLKey, tarURL)
	addEnvToPod(pod, putURLKey, putURL)

	if buildpackURL != "" {
		addEnvToPod(pod, "BUILDPACK_URL", buildpackURL)
	}

	return &pod
}

func buildPod(debug, withAuth bool, name, namespace string, env map[string]interface{}) api.Pod {
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
				"heritage": "deis",
				"version":  "2.0.0-beta",
			},
		},
	}

	if withAuth {
		pod.Spec.Volumes = append(pod.Spec.Volumes, api.Volume{
			Name: minioUser,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: minioUser,
				},
			},
		})

		pod.Spec.Containers[0].VolumeMounts = []api.VolumeMount{
			api.VolumeMount{
				Name:      minioUser,
				MountPath: "/var/run/secrets/object/store",
				ReadOnly:  true,
			},
		}
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
		if pod.Status.Phase == api.PodSucceeded {
			return true, nil
		}
		if pod.Status.Phase == api.PodFailed {
			return true, fmt.Errorf("Giving up; pod went into failed status: \n%s", fmt.Sprintf("%#v", pod))
		}
		return false, nil
	}

	return waitForPodCondition(c, false, ns, podName, condition, interval, timeout)
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

	return waitForPodCondition(c, true, ns, podName, condition, interval, timeout)
}

// waitForPodCondition waits for a pod in state defined by a condition (func)
func waitForPodCondition(c *client.Client, ret bool, ns, podName string, condition func(pod *api.Pod) (bool, error),
	interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		pod, err := c.Pods(ns).Get(podName)
		if err != nil {
			if apierrs.IsNotFound(err) {
				return ret, nil
			}
		}

		done, err := condition(pod)
		if done {
			return true, nil
		}

		return false, nil
	})
}
