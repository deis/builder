package conf

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/sys"
)

func TestGetStorageParams(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Logf("could not retrieve current user: %v", err)
		t.SkipNow()
	}
	if usr.Uid != "0" {
		t.Logf("current user does not have UID of zero (got %s) "+
			"so cannot create storage cred location, skipping", usr.Uid)
		t.SkipNow()
	}

	if err := os.MkdirAll(storageCredLocation, os.ModeDir); err != nil {
		t.Fatalf("could not create storage cred location: %v", err)
	}

	// start by writing out a file to storageCredLocation
	data := []byte("hello world\n")
	if err := ioutil.WriteFile(storageCredLocation+"foo", data, 0644); err != nil {
		t.Fatalf("could not write file to storage cred location: %v", err)
	}

	params, err := GetStorageParams(sys.NewFakeEnv())
	if err != nil {
		t.Errorf("received error while retrieving storage params: %v", err)
	}

	val, ok := params["foo"]
	if !ok {
		t.Error("key foo does not exist in storage params")
	}
	if val != string(data) {
		t.Errorf("expected: %s got: %s", string(data), val)
	}

	// create a directory inside storage cred location, expecting it to pass
	if err := os.Mkdir(storageCredLocation+"bar", os.ModeDir); err != nil {
		t.Fatalf("could not create dir %s: %v", storageCredLocation+"bar", err)
	}

	_, err = GetStorageParams(sys.NewFakeEnv())
	if err != nil {
		t.Errorf("received error while retrieving storage params: %v", err)
	}

	// create the special "..data" directory symlink, expecting it to pass
	if err := os.Symlink(storageCredLocation+"bar", storageCredLocation+"..data"); err != nil {
		t.Fatalf("could not create dir symlink ..data -> %s: %v", storageCredLocation+"bar", err)
	}

	_, err = GetStorageParams(sys.NewFakeEnv())
	if err != nil {
		t.Errorf("received error while retrieving storage params: %v", err)
	}

	env := sys.NewFakeEnv()
	env.Envs = map[string]string{
		"BUILDER_STORAGE":         "minio",
		"DEIS_MINIO_SERVICE_HOST": "localhost",
		"DEIS_MINIO_SERVICE_PORT": "8088",
	}
	params, err = GetStorageParams(env)
	if err != nil {
		t.Errorf("received error while retrieving storage params: %v", err)
	}
	assert.Equal(t, params["regionendpoint"], "http://localhost:8088", "region endpoint")
	assert.Equal(t, params["secure"], false, "secure")
	assert.Equal(t, params["region"], "us-east-1", "region")
	assert.Equal(t, params["bucket"], "git", "bucket")
}

func TestGetControllerClient(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tmpdir")
	if err != nil {
		t.Fatalf("error creating temp directory (%s)", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("failed to remove builder-key from %s (%s)", tmpDir, err)
		}
	}()

	BuilderKeyLocation = filepath.Join(tmpDir, "builder-key")
	data := []byte("testbuilderkey")
	if err := ioutil.WriteFile(BuilderKeyLocation, data, 0644); err != nil {
		t.Fatalf("error creating %s (%s)", BuilderKeyLocation, err)
	}

	key, err := GetBuilderKey()
	assert.NoErr(t, err)
	assert.Equal(t, key, string(data), "data")
}

func TestGetBuilderKeyError(t *testing.T) {
	_, err := GetBuilderKey()
	assert.True(t, err != nil, "no error received when there should have been")
}
