package gitreceive

import (
	"fmt"

	"code.google.com/p/go-uuid/uuid"
	"k8s.io/kubernetes/pkg/api"
)

const (
	podKind   = "Pod"
	v1Version = "v1"
)

func dockerBuilderPodName(appName, shortSha) string {
	uid := uuid.New()[:8]
	return fmt.Sprintf("dockerbuild-%s-%s-%s", appName, shortSha, uid)
}

func slugBuilderPodName(appName, shortSha) string {
	uid := uuid.New()[:8]
	return fmt.Sprintf("slugbuild-%s-%s-%s", appName, shortSha, uid)
}

func dockerBuilderPod(debug, withAuth bool, name, namespace, heritageLabel, versionLabel, tarURL, imageName string) *api.Pod {
	return &api.Pod{
		Kind:       podKind,
		APIVersion: v1Version,
		Name:       name,
		Namespace:  namespace,
		Labels: map[string]string{
			"heritage": heritageLabel,
			"version":  versionLabel,
		},
		Spec: api.PodSpec{},
	}
}

func slugbuilderPod(debug, withAuth bool, name, namespace, heritageLabel, versionLabel, tarURL, putURL, buildpackURL string) *api.Pod {
	return &api.Pod{
		Kind:       podKind,
		APIVersion: v1Version,
		Name:       name,
		Namespace:  namespace,
		Labels: map[string]string{
			"heritage": heritageLabel,
			"version":  versionLabel,
		},
		Spec: api.PodSpec{},
	}
}
