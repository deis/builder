package gitreceive

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/k8s"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	testSecret    = "test-secret"
	deisNamespace = "deis"
)

func TestGetDetailsFromRegistrySecretErr(t *testing.T) {
	expectedErr := errors.New("get secret error")
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &api.Secret{}, expectedErr
		},
	}
	_, err := getDetailsFromRegistrySecret(getter, testSecret)
	assert.Err(t, err, expectedErr)
}

func TestGetDetailsFromRegistrySecretSuccess(t *testing.T) {
	data := map[string][]byte{"test": []byte("test")}
	expectedData := map[string]string{"test": "test"}
	secret := api.Secret{Data: data}
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &secret, nil
		},
	}
	secretData, err := getDetailsFromRegistrySecret(getter, testSecret)
	assert.NoErr(t, err)
	assert.Equal(t, secretData, expectedData, "secret data")
}

func TestGetDetailsFromDockerConfigSecretErr(t *testing.T) {
	expectedErr := errors.New("get secret error")
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &api.Secret{}, expectedErr
		},
	}
	_, err := getDetailsFromDockerConfigSecret(getter, testSecret)
	assert.Err(t, expectedErr, err)
}

func TestGetDetailsFromDockerConfigSecretJsonErr(t *testing.T) {
	expectedErr := errors.New("invalid character 'o' in literal null (expecting 'u')")
	data := map[string][]byte{api.DockerConfigJsonKey: []byte("not a json")}
	secret := api.Secret{Data: data}
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &secret, nil
		},
	}
	_, err := getDetailsFromDockerConfigSecret(getter, testSecret)
	assert.Equal(t, expectedErr.Error(), err.Error(), "error")
}

func TestGetDetailsFromDockerConfigSecretTokenerr(t *testing.T) {
	expectedErr := errors.New("Invalid token in docker config secret")
	auth := []byte(`
    {
    "auths": {
              "https://test.io": {
                  "auth": "test",
                  "email": "none@none.com"
              }
          }
    }
`)
	data := make(map[string][]byte)
	data[api.DockerConfigJsonKey] = auth
	secret := api.Secret{Data: data}
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &secret, nil
		},
	}
	_, err := getDetailsFromDockerConfigSecret(getter, testSecret)
	assert.Err(t, expectedErr, err)
}

func TestGetDetailsFromDockerConfigSecretSuccess(t *testing.T) {
	encToken := base64.StdEncoding.EncodeToString([]byte("testuser:testpassword"))
	auth := []byte(`
    {
    "auths": {
              "https://test.io": {
                  "auth": "` + encToken + `",
                  "email": "none@none.com"
              }
          }
    }
`)
	expectedData := map[string]string{"DEIS_REGISTRY_USERNAME": "testuser", "DEIS_REGISTRY_PASSWORD": "testpassword", "DEIS_REGISTRY_HOSTNAME": "https://test.io"}
	data := make(map[string][]byte)
	data[api.DockerConfigJsonKey] = auth
	secret := api.Secret{Data: data}
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &secret, nil
		},
	}
	regData, err := getDetailsFromDockerConfigSecret(getter, testSecret)
	assert.NoErr(t, err)
	assert.Equal(t, expectedData, regData, "registry details")

}

func TestGetRegistryDetailsOffclusterErr(t *testing.T) {
	expectedErr := errors.New("get secret error")
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &api.Secret{}, expectedErr
		},
	}

	kubeClient := &k8s.FakeSecretsNamespacer{
		Fn: func(string) client.SecretsInterface {
			return getter
		},
	}
	image := "test-image"
	_, err := getRegistryDetails(kubeClient, &image, "off-cluster", deisNamespace, "private-registry")
	assert.Err(t, err, expectedErr)
}

func TestGetRegistryDetailsOffclusterSuccess(t *testing.T) {
	data := map[string][]byte{"organization": []byte("kmala"), "hostname": []byte("quay.io")}
	expectedData := map[string]string{"DEIS_REGISTRY_HOSTNAME": "quay.io", "DEIS_REGISTRY_ORGANIZATION": "kmala"}
	expectedImage := "quay.io/kmala/test-image"
	secret := api.Secret{Data: data}
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &secret, nil
		},
	}

	kubeClient := &k8s.FakeSecretsNamespacer{
		Fn: func(string) client.SecretsInterface {
			return getter
		},
	}
	image := "test-image"
	regDetails, err := getRegistryDetails(kubeClient, &image, "off-cluster", deisNamespace, "private-registry")
	assert.NoErr(t, err)
	assert.Equal(t, expectedData, regDetails, "registry details")
	assert.Equal(t, expectedImage, image, "image")
}

func TestGetRegistryDetailsGCRSuccess(t *testing.T) {
	encToken := base64.StdEncoding.EncodeToString([]byte("testuser:testpassword"))
	auth := []byte(`
    {
    "auths": {
              "https://test.io": {
                  "auth": "` + encToken + `",
                  "email": "none@none.com"
              }
          }
    }
`)
	configData := make(map[string][]byte)
	configData[api.DockerConfigJsonKey] = auth
	configSecret := api.Secret{Data: configData}
	configGetter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &configSecret, nil
		},
	}

	srvAccount := []byte(`
		{
		"project_id": "deis-test"
	}
		`)
	data := map[string][]byte{"key.json": srvAccount}
	secret := api.Secret{Data: data}
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &secret, nil
		},
	}

	kubeClient := &k8s.FakeSecretsNamespacer{
		Fn: func(namespace string) client.SecretsInterface {
			if namespace == "deis" {
				return getter
			}
			return configGetter
		},
	}

	expectedData := map[string]string{"DEIS_REGISTRY_USERNAME": "testuser", "DEIS_REGISTRY_PASSWORD": "testpassword", "DEIS_REGISTRY_HOSTNAME": "https://test.io", "DEIS_REGISTRY_GCS_PROJ_ID": "deis-test"}
	expectedImage := "test.io/deis-test/test-image"

	image := "test-image"
	regDetails, err := getRegistryDetails(kubeClient, &image, "gcr", deisNamespace, "private-registry")

	assert.NoErr(t, err)
	assert.Equal(t, expectedData, regDetails, "registry details")
	assert.Equal(t, expectedImage, image, "image")
}

func TestGetRegistryDetailsGCRConfigErr(t *testing.T) {
	expectedErr := errors.New("get secret error")
	configGetter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &api.Secret{}, expectedErr
		},
	}

	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &api.Secret{}, nil
		},
	}

	kubeClient := &k8s.FakeSecretsNamespacer{
		Fn: func(namespace string) client.SecretsInterface {
			if namespace == "deis" {
				return getter
			}
			return configGetter
		},
	}

	image := "test-image"
	_, err := getRegistryDetails(kubeClient, &image, "gcr", deisNamespace, "private-registry")

	assert.Err(t, err, expectedErr)
}

func TestGetRegistryDetailsGCRSecretErr(t *testing.T) {
	expectedErr := errors.New("get secret error")
	encToken := base64.StdEncoding.EncodeToString([]byte("testuser:testpassword"))
	auth := []byte(`
		{
		"auths": {
							"https://test.io": {
									"auth": "` + encToken + `",
									"email": "none@none.com"
							}
					}
		}
`)
	configData := make(map[string][]byte)
	configData[api.DockerConfigJsonKey] = auth
	configSecret := api.Secret{Data: configData}
	configGetter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &configSecret, nil
		},
	}

	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &api.Secret{}, expectedErr
		},
	}

	kubeClient := &k8s.FakeSecretsNamespacer{
		Fn: func(namespace string) client.SecretsInterface {
			if namespace == "deis" {
				return getter
			}
			return configGetter
		},
	}

	image := "test-image"
	_, err := getRegistryDetails(kubeClient, &image, "gcr", deisNamespace, "private-registry")

	assert.Err(t, err, expectedErr)
}

func TestGetRegistryDetailsGCRJsonErr(t *testing.T) {
	expectedErr := errors.New("invalid character 'e' in literal true (expecting 'r')")
	encToken := base64.StdEncoding.EncodeToString([]byte("testuser:testpassword"))
	auth := []byte(`
    {
    "auths": {
              "https://test.io": {
                  "auth": "` + encToken + `",
                  "email": "none@none.com"
              }
          }
    }
`)
	configData := make(map[string][]byte)
	configData[api.DockerConfigJsonKey] = auth
	configSecret := api.Secret{Data: configData}
	configGetter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &configSecret, nil
		},
	}

	data := map[string][]byte{"key.json": []byte("test")}
	secret := api.Secret{Data: data}
	getter := &k8s.FakeSecret{
		FnGet: func(string) (*api.Secret, error) {
			return &secret, nil
		},
	}

	kubeClient := &k8s.FakeSecretsNamespacer{
		Fn: func(namespace string) client.SecretsInterface {
			if namespace == "deis" {
				return getter
			}
			return configGetter
		},
	}

	image := "test-image"
	_, err := getRegistryDetails(kubeClient, &image, "gcr", deisNamespace, "private-registry")

	assert.Equal(t, expectedErr.Error(), err.Error(), "error")
}
