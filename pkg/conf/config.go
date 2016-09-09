package conf

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/deis/builder/pkg/sys"
	"github.com/kelseyhightower/envconfig"
)

const (
	builderKeyLocation  = "/var/run/secrets/api/auth/builder-key"
	storageCredLocation = "/var/run/secrets/deis/objectstore/creds/"
	minioHostEnvVar     = "DEIS_MINIO_SERVICE_HOST"
	minioPortEnvVar     = "DEIS_MINIO_SERVICE_PORT"
	gcsKey              = "key.json"
)

// Parameters is map which contains storage params
type Parameters map[string]interface{}

// EnvConfig is a convenience function to process the envconfig (
// https://github.com/kelseyhightower/envconfig) based configuration environment variables into
// conf. Additional notes:
//
// - appName will be passed as the first parameter to envconfig.Process
// - conf should be a pointer to an envconfig compatible struct. If you'd like to use struct
// 	 	tags to customize your struct, see
// 		https://github.com/kelseyhightower/envconfig#struct-tag-support
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
	builderKey := strings.Trim(string(builderKeyBytes), "\n")
	return builderKey, nil
}

// GetStorageParams returns the credentials required for connecting to object storage
func GetStorageParams(env sys.Env) (Parameters, error) {
	params := make(map[string]interface{})
	files, err := ioutil.ReadDir(storageCredLocation)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || file.Name() == "..data" {
			continue
		}
		data, err := ioutil.ReadFile(storageCredLocation + file.Name())
		if err != nil {
			return nil, err
		}
		//GCS expect the to have the location of the service account credential json file
		if file.Name() == gcsKey {
			params["keyfile"] = storageCredLocation + file.Name()
		} else {
			params[file.Name()] = string(data)
		}
	}
	params["bucket"] = params["builder-bucket"]
	params["container"] = params["builder-container"]
	if env.Get("BUILDER_STORAGE") == "minio" {
		mHost := env.Get(minioHostEnvVar)
		mPort := env.Get(minioPortEnvVar)
		params["regionendpoint"] = fmt.Sprintf("http://%s:%s", mHost, mPort)
		params["secure"] = false
		params["region"] = "us-east-1"
		params["bucket"] = "git"
	}

	return params, nil
}
