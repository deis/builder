package conf

import (
	"github.com/kelseyhightower/envconfig"
)

func EnvConfig(appName string, conf interface{}) error {
	if err := envconfig.Process(appName, conf); err != nil {
		return err
	}
	return nil
}
