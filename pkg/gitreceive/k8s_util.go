package gitreceive

import (
	"fmt"
	"io/ioutil"

	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
	utilyaml "k8s.io/kubernetes/pkg/util/yaml"
)

var (
	// yamlTempl path to deis yaml files (to allow override during tests)
	yamlTempl = "/etc/deis-%v%v.yaml"
)

func dockerBuilderPodName(appName, shortSha string) string {
	uid := uuid.New()[:8]
	return fmt.Sprintf("dockerbuild-%s-%s-%s", appName, shortSha, uid)
}

func slugBuilderPodName(appName, shortSha string) string {
	uid := uuid.New()[:8]
	return fmt.Sprintf("slugbuild-%s-%s-%s", appName, shortSha, uid)
}

func dockerBuilderPod(debug, withAuth bool, name, namespace string, env map[string]interface{}, tarURL, imageName string) *api.Pod {
	pod := buildPod("dockerbuilder", withAuth, name, namespace, env)
	return &pod
}

func slugbuilderPod(debug, withAuth bool, name, namespace string, env map[string]interface{}, tarURL, putURL, buildpackURL string) *api.Pod {
	pod := buildPod("slugbuilder", withAuth, name, namespace, env)
	return &pod
}

func buildPod(buildType string, withAuth bool, name, namespace string, env map[string]interface{}) api.Pod {
	useCreds := "-no-creds"
	if withAuth {
		useCreds = ""
	}
	fileName := fmt.Sprintf(yamlTempl, buildType, useCreds)

	pod := podFromFile(fileName)

	addEnvToPod(pod, env)

	pod.ObjectMeta.Name = name
	pod.ObjectMeta.Namespace = namespace

	return pod
}

func podFromFile(fileName string) api.Pod {
	var pod api.Pod
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Errorf("missing pod template %v (%v)", fileName, err)
	}

	json, err := utilyaml.ToJSON(data)
	if err != nil {
		fmt.Errorf("invalid pod template %v (%v)", fileName, err)
	}

	api.Scheme.DecodeInto(json, &pod)

	return pod
}

func addEnvToPod(pod api.Pod, env map[string]interface{}) {
	if len(pod.Spec.Containers) > 0 {
		for k, v := range env {
			pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, api.EnvVar{
				Name:  k,
				Value: fmt.Sprintf("%v", v),
			})
		}
	}
}
