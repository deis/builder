package k8s

import (
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

// FakeSecret is a mock function that can be swapped in for
// (k8s.io/kubernetes/pkg/client/unversioned).SecretsInterface,
// so you can unit test your code.
type FakeSecret struct {
	FnGet    func(string) (*api.Secret, error)
	FnCreate func(*api.Secret) (*api.Secret, error)
	FnUpdate func(*api.Secret) (*api.Secret, error)
}

// Get is the interface definition.
func (f *FakeSecret) Get(name string) (*api.Secret, error) {
	return f.FnGet(name)
}

// Delete is the interface definition.
func (f *FakeSecret) Delete(name string) error {
	return nil
}

// Create is the interface definition.
func (f *FakeSecret) Create(secret *api.Secret) (*api.Secret, error) {
	return f.FnCreate(secret)
}

// Update is the interface definition.
func (f *FakeSecret) Update(secret *api.Secret) (*api.Secret, error) {
	return f.FnUpdate(secret)
}

// List is the interface definition.
func (f *FakeSecret) List(opts api.ListOptions) (*api.SecretList, error) {
	return &api.SecretList{}, nil
}

// Watch is the interface definition.
func (f *FakeSecret) Watch(opts api.ListOptions) (watch.Interface, error) {
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
