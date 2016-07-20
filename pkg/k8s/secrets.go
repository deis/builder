package k8s

import (
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

// SecretGetter is a (k8s.io/kubernetes/pkg/client/unversioned).SecretsInterface compatible
// interface which only has the Get function. It's used in places that only need Get to make
// them easier to test and more easily swappable with other implementations
// (should the need arise).
type SecretGetter interface {
	Get(name string) (*api.Secret, error)
}

// FakeSecretGetter is a mock function that can be swapped in for an SecretGetter or
// (k8s.io/kubernetes/pkg/client/unversioned).SecretsInterface, so you can
// unit test your code.
type FakeSecretGetter struct {
	Fn func(string) (*api.Secret, error)
}

// Get is the interface definition.
func (f *FakeSecretGetter) Get(name string) (*api.Secret, error) {
	return f.Fn(name)
}

// Delete is the interface definition.
func (f *FakeSecretGetter) Delete(name string) error {
	return nil
}

// Create is the interface definition.
func (f *FakeSecretGetter) Create(secret *api.Secret) (*api.Secret, error) {
	return &api.Secret{}, nil
}

// Update is the interface definition.
func (f *FakeSecretGetter) Update(secret *api.Secret) (*api.Secret, error) {
	return &api.Secret{}, nil
}

// List is the interface definition.
func (f *FakeSecretGetter) List(opts api.ListOptions) (*api.SecretList, error) {
	return &api.SecretList{}, nil
}

// Watch is the interface definition.
func (f *FakeSecretGetter) Watch(opts api.ListOptions) (watch.Interface, error) {
	return nil, nil
}

// FakeSecretsNamespacer is a mock function that can be swapped in for an
// (k8s.io/kubernetes/pkg/client/unversioned).SecretsNamespacer, so you can unit test you code
type FakeSecretsNamespacer struct {
	Fn func(string) client.SecretsInterface
}

// Secrets is the interface definition.
func (f *FakeSecretsNamespacer) Secrets(namespace string) client.SecretsInterface {
	return f.Fn(namespace)
}
