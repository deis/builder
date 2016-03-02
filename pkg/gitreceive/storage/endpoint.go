package storage

import (
	"fmt"
	"os"
	"strings"

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

func stripScheme(str string) string {
	schemes := []string{"http://", "https://"}
	for _, scheme := range schemes {
		if strings.HasPrefix(str, scheme) {
			str = str[len(scheme):]
		}
	}
	return str
}

type endpoint struct {
	urlStr string
	secure bool
}

func getEndpoint(env sys.Env) (*endpoint, error) {
	mHost := env.Get(minioHostEnvVar)
	mPort := env.Get(minioPortEnvVar)
	S3EP := env.Get(outsideStorageEndpoint)
	if S3EP != "" {
		return &endpoint{urlStr: stripScheme(S3EP), secure: true}, nil
	} else if mHost != "" && mPort != "" {
		return &endpoint{urlStr: fmt.Sprintf("%s:%s", mHost, mPort), secure: false}, nil
	} else {
		return nil, errNoStorageConfig
	}
}
