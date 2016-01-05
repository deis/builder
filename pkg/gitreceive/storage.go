package gitreceive

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

type storageConfig interface {
	schema() string
	host() string
	port() string
}

func getStorageConfig() (storageConfig, error) {
	mHost := os.Getenv(minioHostEnvVar)
	mPort := os.Getenv(minioPortEnvVar)
	oHost := os.Getenv(outsideStorageHostEnvVar)
	oPort := os.Getenv(outsideStoragePortEnvVar)
	if mHost != "" && mPort != "" {
		return minioConfig{hst: mHost, prt: mPort}, nil
	} else if oHost != "" && oPort != "" {
		return outsideConfig{hst: oHost, prt: oPort}, nil
	} else {
		return nil, errNoStorageConfig
	}
}

type minioConfig struct {
	hst string
	prt string
}

func (m minioConfig) schema() string { return "http" }
func (m minioConfig) host() string   { return m.hst }
func (m minioConfig) port() string   { return m.prt }

type outsideConfig struct {
	hst string
	prt string
}

func (o outsideConfig) schema() string { return "https" }
func (o outsideConfig) host() string   { return o.hst }
func (o outsideConfig) port() string   { return o.prt }
