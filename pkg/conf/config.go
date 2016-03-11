package conf

import (
	"fmt"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
)

const (
	builderKeyLocation = "/var/run/secrets/api/auth/builder-key"
)

func EnvConfig(appName string, conf interface{}) error {
	if err := envconfig.Process(appName, conf); err != nil {
		return err
	}
	return nil
}

// GetBuilderKey returns the key to be used as token to interact with deis-controller
func GetBuilderKey() (string, error) {
	builderKeyBytes, err := ioutil.ReadFile(builderKeyLocation)
	if err != nil {
		return "", fmt.Errorf("couldn't get builder key from %s (%s)", builderKeyLocation, err)
	}
	builderKey := string(builderKeyBytes)
	return builderKey, nil
}
