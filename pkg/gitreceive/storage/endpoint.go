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

type Endpoint struct {
	URLStr string
	Secure bool
}

func getEndpoint(env sys.Env) (*Endpoint, error) {
	mHost := env.Get(minioHostEnvVar)
	mPort := env.Get(minioPortEnvVar)
	S3EP := env.Get(outsideStorageEndpoint)
	if S3EP != "" {
		return &Endpoint{URLStr: stripScheme(S3EP), Secure: true}, nil
	} else if mHost != "" && mPort != "" {
		return &Endpoint{URLStr: fmt.Sprintf("%s:%s", mHost, mPort), Secure: false}, nil
	} else {
		return nil, errNoStorageConfig
	}
}
