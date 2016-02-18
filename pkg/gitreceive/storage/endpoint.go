package storage

import (
	"fmt"
	"os"
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

func getEndpoint() (string, error) {
	mHost := os.Getenv(minioHostEnvVar)
	mPort := os.Getenv(minioPortEnvVar)
	S3EP := os.Getenv(outsideStorageEndpoint)
	if S3EP != "" {
		return S3EP, nil
	} else if mHost != "" && mPort != "" {
		return fmt.Sprintf("http://%s:%s", mHost, mPort), nil
	} else {
		return "", errNoStorageConfig
	}
}
