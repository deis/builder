package sys

import (
	"os"
)

// Env is an interface to a set of environment variables.
type Env interface {
	// Get gets the environment variable of the given name.
	Get(name string) string
}

type realEnv struct{}

func (r realEnv) Get(name string) string {
	return os.Getenv(name)
}

// RealEnv returns an Env implementation that uses os.Getenv every time Get is called.
func RealEnv() Env {
	return realEnv{}
}

// FakeEnv is an Env implementation that stores the environment in a map.
type FakeEnv struct {
	// Envs is the map from which Get calls will look to retrieve environment variables.
	Envs map[string]string
}

// NewFakeEnv returns a new FakeEnv with no values in Envs.
func NewFakeEnv() *FakeEnv {
	return &FakeEnv{Envs: make(map[string]string)}
}

// Get is the Env interface implementation for Env.
func (f *FakeEnv) Get(name string) string {
	return f.Envs[name]
}
