package controller

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/arschles/assert"
)

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
	host := "127.0.0.1"
	port := "80"
	cli, err := New(host, port, tmpDir+"/builder-key")
	assert.NoErr(t, err)
	assert.Equal(t, cli.ControllerURL.String(), fmt.Sprintf("http://%s:%s/", host, port), "data")
	assert.Equal(t, cli.HooksToken, string(data), "data")
}

func TestGetControllerClientError(t *testing.T) {
	host := "127.0.0.1"
	port := "80"
	_, err := New(host, port, "/builder-key")
	assert.True(t, err != nil, "no error received when there should have been")
}
