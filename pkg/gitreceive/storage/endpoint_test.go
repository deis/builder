package storage

import (
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/sys"
)

type getEndpointTestCase struct {
	envVars     map[string]string
	expectedOut *Endpoint
	expectedErr error
}

func TestGetEndpoint(t *testing.T) {
	testCases := []getEndpointTestCase{
		getEndpointTestCase{
			envVars:     map[string]string{"DEIS_OUTSIDE_STORAGE": "http://outside.storage.com"},
			expectedOut: &Endpoint{URLStr: "outside.storage.com", Secure: true},
		},
		getEndpointTestCase{
			envVars:     map[string]string{"DEIS_OUTSIDE_STORAGE": "https://outside.com"},
			expectedOut: &Endpoint{URLStr: "outside.com", Secure: true},
		},
		getEndpointTestCase{
			envVars: map[string]string{
				"DEIS_OUTSIDE_STORAGE":    "outside.com",
				"DEIS_MINIO_SERVICE_HOST": "minio.com",
				"DEIS_MINIO_SERVICE_PORT": "8888",
			},
			expectedOut: &Endpoint{URLStr: "outside.com", Secure: true},
		},
		getEndpointTestCase{
			envVars: map[string]string{
				"DEIS_MINIO_SERVICE_HOST": "minio.com",
				"DEIS_MINIO_SERVICE_PORT": "8888",
			},
			expectedOut: &Endpoint{URLStr: "minio.com:8888", Secure: false},
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
		ep, err := getEndpoint(fe)

		if testCase.expectedOut != nil {
			assert.Equal(t, ep.URLStr, testCase.expectedOut.URLStr, "url string")
			assert.Equal(t, ep.Secure, testCase.expectedOut.Secure, "secure boolean")
		} else {
			assert.True(t, ep == nil, "endpoint was non-nil when it should have been")
		}

		if testCase.expectedErr == nil {
			assert.NoErr(t, err)
		} else {
			assert.Equal(t, err, testCase.expectedErr, "error")
		}
	}
}

type schemeTestCase struct {
	before string
	after  string
}

func TestStripScheme(t *testing.T) {
	schemes := []schemeTestCase{
		schemeTestCase{before: "https://deis.com", after: "deis.com"},
		schemeTestCase{before: "http://deis.com", after: "deis.com"},
		schemeTestCase{before: "deis.com", after: "deis.com"},
		schemeTestCase{before: "://deis.com", after: "://deis.com"},
	}
	for i, scheme := range schemes {
		assert.Equal(t, stripScheme(scheme.before), scheme.after, fmt.Sprintf("scheme %s (# %d)", scheme.before, i))
	}
}
