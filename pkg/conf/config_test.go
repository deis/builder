package conf

import (
	"io/ioutil"
	"os"
	"os/user"
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
}

func TestGetControllerClient(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tmpdir")
	if err != nil {
		t.Fatalf("error creating temp directory (%s)", err)
	}
	data := []byte("testbuilderkey")
	if err := ioutil.WriteFile(tmpDir+"/builder-key", data, 0644); err != nil {
		t.Fatalf("error creating %s/builder-key (%s)", tmpDir, err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("failed to remove builder-key from %s (%s)", tmpDir, err)
		}
	}()
	key, err := GetBuilderKey(tmpDir + "/builder-key")
	assert.NoErr(t, err)
	assert.Equal(t, key, string(data), "data")
}

func TestGetBuilderKeyError(t *testing.T) {
	_, err := GetBuilderKey("/builder-key")
	assert.True(t, err != nil, "no error received when there should have been")
}
