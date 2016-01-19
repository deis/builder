package storage

import (
	"fmt"
	"os"
)

const (
	minioHostEnvVar          = "DEIS_MINIO_SERVICE_HOST"
	minioPortEnvVar          = "DEIS_MINIO_SERVICE_PORT"
	outsideStorageHostEnvVar = "DEIS_OUTSIDE_STORAGE_HOST"
	outsideStoragePortEnvVar = "DEIS_OUTSIDE_STORAGE_PORT"
)

var (
	errNoStorageConfig = fmt.Errorf(
		"no storage config variables found (%s:%s or %s:%s)",
		minioHostEnvVar,
		minioPortEnvVar,
		outsideStorageHostEnvVar,
		outsideStoragePortEnvVar,
	)
)

func getEndpoint() (string, error) {
	mHost := os.Getenv(minioHostEnvVar)
	mPort := os.Getenv(minioPortEnvVar)
	oHost := os.Getenv(outsideStorageHostEnvVar)
	oPort := os.Getenv(outsideStoragePortEnvVar)
	if mHost != "" && mPort != "" {
		return fmt.Sprintf("http://%s:%s", mHost, mPort), nil
	} else if oHost != "" && oPort != "" {
		return fmt.Sprintf("https://%s:%s", oHost, oPort), nil
	} else {
		return "", errNoStorageConfig
	}
}
