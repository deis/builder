package storage

import (
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/sys"
)

type getEndpointTestCase struct {
	envVars     map[string]string
	expectedOut string
	expectedErr error
}

func TestGetEndpoint(t *testing.T) {
	testCases := []getEndpointTestCase{
		getEndpointTestCase{
			envVars:     map[string]string{"DEIS_OUTSIDE_STORAGE": "http://outside.storage.com"},
			expectedOut: "http://outside.storage.com",
		},
		getEndpointTestCase{
			envVars:     map[string]string{"DEIS_OUTSIDE_STORAGE": "https://outside.com"},
			expectedOut: "https://outside.com",
		},
		getEndpointTestCase{
			envVars: map[string]string{
				"DEIS_OUTSIDE_STORAGE":    "outside.com",
				"DEIS_MINIO_SERVICE_HOST": "minio.com",
				"DEIS_MINIO_SERVICE_PORT": "8888",
			},
			expectedOut: "outside.com",
		},
		getEndpointTestCase{
			envVars: map[string]string{
				"DEIS_MINIO_SERVICE_HOST": "minio.com",
				"DEIS_MINIO_SERVICE_PORT": "8888",
			},
			expectedOut: "http://minio.com:8888",
		},
		getEndpointTestCase{
			envVars: map[string]string{
				"DEIS_MINIO_SERVICE_HOST": "minio.com",
			},
			expectedErr: errNoStorageConfig,
		},
		getEndpointTestCase{
			envVars: map[string]string{
				"DEIS_MINIO_SERVICE_PORT": "9999",
			},
			expectedErr: errNoStorageConfig,
		},
	}
	for _, testCase := range testCases {
		fe := sys.NewFakeEnv()
		fe.Envs = testCase.envVars
		str, err := getEndpoint(fe)
		assert.Equal(t, str, testCase.expectedOut, "output")
		if testCase.expectedErr == nil {
			assert.NoErr(t, err)
		} else {
			assert.Equal(t, err, testCase.expectedErr, "error")
		}
	}
}
