package storage

import (
	"fmt"

	"github.com/deis/builder/pkg/sys"
)

const (
	minioHostEnvVar        = "DEIS_MINIO_SERVICE_HOST"
	minioPortEnvVar        = "DEIS_MINIO_SERVICE_PORT"
	outsideStorageEndpoint = "DEIS_OUTSIDE_STORAGE"
)

var (
	errNoStorageConfig = fmt.Errorf(
		"no storage config variables found (%s:%s or %s)",
		minioHostEnvVar,
		minioPortEnvVar,
		outsideStorageEndpoint,
	)
)

func getEndpoint(env sys.Env) (string, error) {
	mHost := env.Get(minioHostEnvVar)
	mPort := env.Get(minioPortEnvVar)
	S3EP := env.Get(outsideStorageEndpoint)
	if S3EP != "" {
		return S3EP, nil
	} else if mHost != "" && mPort != "" {
		return fmt.Sprintf("http://%s:%s", mHost, mPort), nil
	} else {
		return "", errNoStorageConfig
	}
}
