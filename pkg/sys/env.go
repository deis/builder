package sys

import (
	"os"
)

// Env is an interface to a set of environment variables
type Env interface {
	// Get gets the environment variable of the given name
	Get(name string) string
}

type realEnv struct{}

func (r realEnv) Get(name string) string {
	return os.Getenv(name)
}

func RealEnv() Env {
	return realEnv{}
}

type FakeEnv struct {
	Envs map[string]string
}

func NewFakeEnv() *FakeEnv {
	return &FakeEnv{Envs: make(map[string]string)}
}

func (f *FakeEnv) Get(name string) string {
	return f.Envs[name]
}
