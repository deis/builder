package gitreceive

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/storage"
	"github.com/docker/distribution/context"
	"gopkg.in/yaml.v2"
)

const (
	bucketName = "mybucket"
	objKey     = "myobj"
)

func TestGetProcFileFromRepoSuccess(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tmpdir")
	if err != nil {
		t.Fatalf("error creating temp directory (%s)", err)
	}
	data := []byte("web: example-go")
	if err := ioutil.WriteFile(tmpDir+"/Procfile", data, 0644); err != nil {
		t.Fatalf("error creating %s/Procfile (%s)", tmpDir, err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("failed to remove Procfile from %s (%s)", tmpDir, err)
		}
	}()
	getter := &storage.FakeObjectGetter{}
	procType, err := getProcFile(getter, tmpDir, objKey)
	actualData := pkg.ProcessType{}
	yaml.Unmarshal(data, &actualData)
	assert.NoErr(t, err)
	assert.Equal(t, procType, actualData, "data")
}

func TestGetProcFileFromRepoFailure(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tmpdir")
	if err != nil {
		t.Fatalf("error creating temp directory (%s)", err)
	}
	data := []byte("web= example-go")
	if err := ioutil.WriteFile(tmpDir+"/Procfile", data, 0644); err != nil {
		t.Fatalf("error creating %s/Procfile (%s)", tmpDir, err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("failed to remove Procfile from %s (%s)", tmpDir, err)
		}
	}()
	getter := &storage.FakeObjectGetter{}
	_, err = getProcFile(getter, tmpDir, objKey)

	assert.True(t, err != nil, "no error received when there should have been")
}

func TestGetProcFileFromServerSuccess(t *testing.T) {
	data := []byte("web: example-go")
	getter := &storage.FakeObjectGetter{
		Fn: func(context.Context, string) ([]byte, error) {
			return data, nil
		},
	}

	procType, err := getProcFile(getter, "", objKey)
	actualData := pkg.ProcessType{}
	yaml.Unmarshal(data, &actualData)
	assert.NoErr(t, err)
	assert.Equal(t, procType, actualData, "data")
}

func TestGetProcFileFromServerFailure(t *testing.T) {
	expectedErr := errors.New("test error")
	getter := &storage.FakeObjectGetter{
		Fn: func(context.Context, string) ([]byte, error) {
			return []byte("web: example-go"), expectedErr
		},
	}

	_, err := getProcFile(getter, "", objKey)
	assert.Err(t, err, fmt.Errorf("error in reading %s (%s)", objKey, expectedErr))
	assert.True(t, err != nil, "no error received when there should have been")
}
