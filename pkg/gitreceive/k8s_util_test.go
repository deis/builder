package gitreceive

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
)

func DockerBuilderPodName(t *testing.T) {
	name := dockerBuilderPodName("demo", "12345678")
	if !strings.HasPrefix(name, "dockerbuild-demo-12345678-") {
		t.Fatalf("expected pod name dockerbuild-demo-12345678-, got %s", name)
	}
}

func SlugBuilderPodName(t *testing.T) {
	name := dockerBuilderPodName("demo", "12345678")
	if !strings.HasPrefix(name, "slugbuild-demo-12345678-") {
		t.Fatalf("expected pod name slugbuild-demo-12345678-, got %s", name)
	}
}

type buildCase struct {
	buildType string
	withAuth  bool
	name      string
	namespace string
	env       map[string]interface{}
	templ     string
}

func TestBuildPod(t *testing.T) {
	emptyEnv := make(map[string]interface{})

	env := make(map[string]interface{})
	env["KEY"] = "VALUE"

	var pod api.Pod
	var expPod api.Pod

	yamlTempl = "../../rootfs/etc/deis-%v%v.yaml"

	builds := []buildCase{
		{"slugbuilder", true, "test", "default", emptyEnv, "../../rootfs/etc/deis-slugbuilder.yaml"},
		{"slugbuilder", false, "test", "default", emptyEnv, "../../rootfs/etc/deis-slugbuilder-no-creds.yaml"},
		{"slugbuilder", true, "test", "default", env, "../../rootfs/etc/deis-slugbuilder.yaml"},
		{"slugbuilder", false, "test", "default", env, "../../rootfs/etc/deis-slugbuilder-no-creds.yaml"},
		{"dockerbuilder", true, "test", "default", emptyEnv, "../../rootfs/etc/deis-dockerbuilder.yaml"},
		{"dockerbuilder", false, "test", "default", emptyEnv, "../../rootfs/etc/deis-dockerbuilder-no-creds.yaml"},
		{"dockerbuilder", true, "test", "default", env, "../../rootfs/etc/deis-dockerbuilder.yaml"},
		{"dockerbuilder", false, "test", "default", env, "../../rootfs/etc/deis-dockerbuilder-no-creds.yaml"},
	}

	for _, build := range builds {
		pod = buildPod(build.buildType, build.withAuth, build.name, build.namespace, build.env)
		expPod = buildTestPod(build.name, build.namespace, build.env, build.templ)
		if !reflect.DeepEqual(expPod, pod) {
			t.Errorf("%v and %v are not equal", expPod, pod)
		}
	}
}

func buildTestPod(name, namespace string, env map[string]interface{}, templ string) api.Pod {
	pod := podFromFile(templ)
	pod.ObjectMeta.Name = name
	pod.ObjectMeta.Namespace = namespace
	addEnvToPod(pod, env)
	return pod
}
