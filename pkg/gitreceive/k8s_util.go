package gitreceive

import (
	"fmt"

	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
)

const (
	slugBuilderName   = "deis-slugbuilder"
	dockerBuilderName = "deis-dockerbuilder"

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

func dockerBuilderPod(debug, withAuth bool, name, namespace string, env map[string]interface{}, tarURL, imageName, region, dockerBuilderImage string) *api.Pod {
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

func slugbuilderPod(debug, withAuth bool, name, namespace string, env map[string]interface{}, tarURL, putURL, buildpackURL, slugBuilderImage string) *api.Pod {
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
				"heritage": name,
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
