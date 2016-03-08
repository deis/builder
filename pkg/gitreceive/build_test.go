package gitreceive

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/gitreceive/storage"
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
	getter := &storage.RealObjectGetter{}
	procType, err := getProcFile(getter, tmpDir, bucketName, objKey)
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
	getter := &storage.RealObjectGetter{}
	_, err = getProcFile(getter, tmpDir, bucketName, objKey)

	assert.True(t, err != nil, "no error received when there should have been")
}

func TestGetProcFileFromServerSuccess(t *testing.T) {
	data := []byte("web: example-go")
	obj := &storage.FakeObject{Data: string(data)}
	getter := &storage.FakeObjectGetter{
		Fn: func(string, string) (storage.Object, error) {
			return obj, nil
		},
	}

	procType, err := getProcFile(getter, "", bucketName, objKey)
	actualData := pkg.ProcessType{}
	yaml.Unmarshal(data, &actualData)
	assert.NoErr(t, err)
	assert.Equal(t, procType, actualData, "data")
}

func TestGetProcFileFromServerFailure(t *testing.T) {
	data := []byte("web: example-go")
	obj := &storage.FakeObject{Data: string(data)}
	expectedErr := errors.New("test error")
	getter := &storage.FakeObjectGetter{
		Fn: func(string, string) (storage.Object, error) {
			return obj, expectedErr
		},
	}

	_, err := getProcFile(getter, "", bucketName, objKey)
	assert.Err(t, err, fmt.Errorf("error in reading %s/%s (%s)", bucketName, objKey, expectedErr))
	assert.True(t, err != nil, "no error received when there should have been")
}
