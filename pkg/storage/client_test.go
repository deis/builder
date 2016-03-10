package storage

import (
	"fmt"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/sys"
)

type getClientTestCase struct {
	fileSystem   map[string][]byte
	envs         map[string]string
	expectClient bool
	expectErr    bool
}

func TestGetClientEmptyAuth(t *testing.T) {
	testCases := []getClientTestCase{
		// no auth & outside storage
		getClientTestCase{
			fileSystem:   make(map[string][]byte),
			envs:         map[string]string{"DEIS_OUTSIDE_STORAGE": "http://outside.com"},
			expectClient: true,
			expectErr:    false,
		},
		// invalid auth - missing secret
		getClientTestCase{
			fileSystem:   map[string][]byte{accessKeyIDFile: []byte("access key")},
			envs:         map[string]string{"DEIS_OUTSIDE_STORAGE": "http://outside.com"},
			expectClient: false,
			expectErr:    true,
		},
		// invalid auth - missing key
		getClientTestCase{
			fileSystem:   map[string][]byte{accessSecretKeyFile: []byte("access secret")},
			envs:         map[string]string{"DEIS_OUTSIDE_STORAGE": "http://outside.com"},
			expectClient: false,
			expectErr:    true,
		},
		// invalid endpoint
		getClientTestCase{
			fileSystem:   make(map[string][]byte),
			envs:         make(map[string]string),
			expectClient: false,
			expectErr:    true,
		},
		// valid auth, outside storage
		getClientTestCase{
			fileSystem:   map[string][]byte{accessKeyIDFile: []byte("access key"), accessSecretKeyFile: []byte("access secret")},
			envs:         map[string]string{"DEIS_OUTSIDE_STORAGE": "http://outside.com"},
			expectClient: true,
			expectErr:    false,
		},
	}
	for i, testCase := range testCases {
		fs := sys.NewFakeFS()
		fs.Files = testCase.fileSystem
		env := sys.NewFakeEnv()
		env.Envs = testCase.envs
		cl, err := GetClient("myrevion", fs, env)
		assert.Equal(t, testCase.expectClient, cl != nil, fmt.Sprintf("returned client %d", i))
		assert.Equal(t, testCase.expectErr, err != nil, fmt.Sprintf("returned error %d", i))
	}
}
