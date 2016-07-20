package gitreceive

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/deis/builder/pkg/k8s"
	"github.com/deis/builder/pkg/storage"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	registrySecret = "registry-secret"
)

func getDetailsFromRegistrySecret(secretGetter k8s.SecretGetter, secret string) (map[string]string, error) {
	regSecret, err := secretGetter.Get(secret)
	if err != nil {
		return nil, err
	}
	regDetails := make(map[string]string)
	for key, value := range regSecret.Data {
		regDetails[key] = string(value)
	}
	return regDetails, nil
}

func getDetailsFromDockerConfigSecret(secretGetter k8s.SecretGetter, secret string) (map[string]string, error) {
	configSecret, err := secretGetter.Get(secret)
	if err != nil {
		return nil, err
	}
	dockerConfigJSONBytes := configSecret.Data[api.DockerConfigJsonKey]
	var secretData map[string]interface{}
	if err = json.Unmarshal(dockerConfigJSONBytes, &secretData); err != nil {
		return nil, err
	}

	var authdata map[string]interface{}
	var hostname string
	for key, value := range secretData["auths"].(map[string]interface{}) {
		hostname = key
		authdata = value.(map[string]interface{})
	}
	token := authdata["auth"].(string)
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(string(decodedToken), ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("Invalid token in docker config secret")
	}
	user := parts[0]
	password := parts[1]
	regDetails := make(map[string]string)
	regDetails["DEIS_REGISTRY_USERNAME"] = user
	regDetails["DEIS_REGISTRY_PASSWORD"] = password
	regDetails["DEIS_REGISTRY_HOSTNAME"] = hostname
	return regDetails, nil
}

func getRegistryDetails(kubeClient client.SecretsNamespacer, image *string, registryLocation, namespace, registrySecretPrefix string) (map[string]string, error) {
	registryConfigSecretInterface := kubeClient.Secrets(*image)
	privateRegistrySecretInterface := kubeClient.Secrets(namespace)
	registryEnv := make(map[string]string)
	var regSecretData map[string]string
	var err error
	if registryLocation == "off-cluster" {
		regSecretData, err = getDetailsFromRegistrySecret(privateRegistrySecretInterface, registrySecret)
		if err != nil {
			return nil, err
		}
		for key, value := range regSecretData {
			registryEnv["DEIS_REGISTRY_"+strings.ToUpper(key)] = value
		}
		if registryEnv["DEIS_REGISTRY_ORGANIZATION"] != "" {
			*image = registryEnv["DEIS_REGISTRY_ORGANIZATION"] + "/" + *image
		}
		if registryEnv["DEIS_REGISTRY_HOSTNAME"] != "" {
			*image = registryEnv["DEIS_REGISTRY_HOSTNAME"] + "/" + *image
		}
	} else if registryLocation == "ecr" {
		registryEnv, err = getDetailsFromDockerConfigSecret(registryConfigSecretInterface, registrySecretPrefix+"-"+registryLocation)
		if err != nil {
			return nil, err
		}

		regSecretData, err = getDetailsFromRegistrySecret(privateRegistrySecretInterface, registrySecret)
		if err != nil {
			return nil, err
		}
		err = storage.CreateImageRepo(*image, regSecretData)
		if err != nil {
			return nil, err
		}
		hostname := strings.Replace(registryEnv["DEIS_REGISTRY_HOSTNAME"], "https://", "", 1)
		*image = hostname + "/" + *image

	} else if registryLocation == "gcr" {
		registryEnv, err = getDetailsFromDockerConfigSecret(registryConfigSecretInterface, registrySecretPrefix+"-"+registryLocation)
		if err != nil {
			return nil, err
		}

		regSecretData, err = getDetailsFromRegistrySecret(privateRegistrySecretInterface, registrySecret)
		if err != nil {
			return nil, err
		}
		var key struct {
			ProjectID string `json:"project_id"`
		}
		jsonKey := []byte(regSecretData["key.json"])
		if err := json.Unmarshal(jsonKey, &key); err != nil {
			return nil, err
		}
		hostname := strings.Replace(registryEnv["DEIS_REGISTRY_HOSTNAME"], "https://", "", 1)
		projectID := strings.Replace(key.ProjectID, ":", "/", -1)
		registryEnv["DEIS_REGISTRY_GCS_PROJ_ID"] = projectID
		*image = strings.Replace(hostname, "https://", "", 1) + "/" + projectID + "/" + *image
	}
	return registryEnv, nil
}
