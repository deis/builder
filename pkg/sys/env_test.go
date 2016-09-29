package sys

import (
	"os"
	"testing"
)

const expectedEnv string = "mmmcoffee"

func TestRealEnvGet(t *testing.T) {
	e := RealEnv()
	os.Setenv("DEIS_BUILDER_REAL_ENV_TEST", expectedEnv)
	if actual := e.Get("DEIS_BUILDER_REAL_ENV_TEST"); actual != expectedEnv {
		t.Errorf("expected '%s', got '%s'", expectedEnv, actual)
	}
}

func TestFakeEnvGet(t *testing.T) {
	e := NewFakeEnv()
	e.Envs["DEIS_BUILDER_FAKE_ENV_TEST"] = expectedEnv
	if actual := e.Get("DEIS_BUILDER_FAKE_ENV_TEST"); actual != expectedEnv {
		t.Errorf("expected '%s', got '%s'", expectedEnv, actual)
	}
}
