package gitreceive

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
)

func TestDockerBuilderPodName(t *testing.T) {
	name := dockerBuilderPodName("demo", "12345678")
	if !strings.HasPrefix(name, "dockerbuild-demo-12345678-") {
		t.Fatalf("expected pod name dockerbuild-demo-12345678-*, got %s", name)
	}
}

func TestSlugBuilderPodName(t *testing.T) {
	name := slugBuilderPodName("demo", "12345678")
	if !strings.HasPrefix(name, "slugbuild-demo-12345678-") {
		t.Fatalf("expected pod name slugbuild-demo-12345678-*, got %s", name)
	}
}

type slugBuildCase struct {
	debug            bool
	withAuth         bool
	name             string
	namespace        string
	env              map[string]interface{}
	tarURL           string
	putURL           string
	buildPack        string
	slugBuilderImage string
}

type dockerBuildCase struct {
	debug              bool
	withAuth           bool
	name               string
	namespace          string
	env                map[string]interface{}
	tarURL             string
	imgName            string
	region             string
	dockerBuilderImage string
}

func TestBuildPod(t *testing.T) {
	emptyEnv := make(map[string]interface{})

	env := make(map[string]interface{})
	env["KEY"] = "VALUE"

	var pod *api.Pod

	slugBuilds := []slugBuildCase{
		{true, true, "test", "default", emptyEnv, "tar", "put-url", "", ""},
		{true, false, "test", "default", emptyEnv, "tar", "put-url", "", ""},
		{true, true, "test", "default", env, "tar", "put-url", "", ""},
		{true, false, "test", "default", env, "tar", "put-url", "", ""},
		{true, true, "test", "default", emptyEnv, "tar", "put-url", "buildpack", ""},
		{true, false, "test", "default", emptyEnv, "tar", "put-url", "buildpack", ""},
		{true, true, "test", "default", env, "tar", "put-url", "buildpack", ""},
		{true, false, "test", "default", env, "tar", "put-url", "buildpack", "customimage"},
	}

	for _, build := range slugBuilds {
		pod = slugbuilderPod(build.debug, build.withAuth, build.name, build.namespace, build.env, build.tarURL, build.putURL, build.buildPack, build.slugBuilderImage)

		if pod.ObjectMeta.Name != build.name {
			t.Errorf("expected %v but returned %v ", build.name, pod.ObjectMeta.Name)
		}

		if pod.ObjectMeta.Namespace != build.namespace {
			t.Errorf("expected %v but returned %v ", build.namespace, pod.ObjectMeta.Namespace)
		}

		checkForEnv(t, pod, "TAR_URL", build.tarURL)
		checkForEnv(t, pod, "put_url", build.putURL)

		if build.buildPack != "" {
			checkForEnv(t, pod, "BUILDPACK_URL", build.buildPack)
		}

		if build.slugBuilderImage != "" {
			if pod.Spec.Containers[0].Image != build.slugBuilderImage {
				t.Errorf("expected %v but returned %v ", build.slugBuilderImage, pod.Spec.Containers[0].Image)
			}
		}
	}

	dockerBuilds := []dockerBuildCase{
		{true, true, "test", "default", emptyEnv, "tar", "", "us-east-1", ""},
		{true, false, "test", "default", emptyEnv, "tar", "", "us-east-1", ""},
		{true, true, "test", "default", env, "tar", "", "us-east-1", ""},
		{true, false, "test", "default", env, "tar", "", "us-east-1", ""},
		{true, true, "test", "default", emptyEnv, "tar", "img", "us-east-1", ""},
		{true, false, "test", "default", emptyEnv, "tar", "img", "us-east-1", ""},
		{true, true, "test", "default", env, "tar", "img", "us-east-1", ""},
		{true, false, "test", "default", env, "tar", "img", "us-east-1", "customimage"},
	}

	for _, build := range dockerBuilds {
		pod = dockerBuilderPod(build.debug, build.withAuth, build.name, build.namespace, build.env, build.tarURL, build.imgName, build.region, build.dockerBuilderImage)

		if pod.ObjectMeta.Name != build.name {
			t.Errorf("expected %v but returned %v ", build.name, pod.ObjectMeta.Name)
		}
		if pod.ObjectMeta.Namespace != build.namespace {
			t.Errorf("expected %v but returned %v ", build.namespace, pod.ObjectMeta.Namespace)
		}
		if !build.withAuth {
			checkForEnv(t, pod, "TAR_URL", build.tarURL)
			checkForEnv(t, pod, "IMG_NAME", build.imgName)
		}
		if build.dockerBuilderImage != "" {
			if pod.Spec.Containers[0].Image != build.dockerBuilderImage {
				t.Errorf("expected %v but returned %v ", build.dockerBuilderImage, pod.Spec.Containers[0].Image)
			}
		}
	}
}

func checkForEnv(t *testing.T, pod *api.Pod, key, expVal string) {
	val, err := envValueFromKey(pod, key)
	if err != nil {
		t.Errorf("%v", err)
	}
	if val != val {
		t.Errorf("expected %v but returned %v ", expVal, val)
	}
}

func envValueFromKey(pod *api.Pod, key string) (string, error) {
	for _, env := range pod.Spec.Containers[0].Env {
		if env.Name == key {
			return env.Value, nil
		}
	}

	return "", fmt.Errorf("no key with name %v found in pod env", key)
}
