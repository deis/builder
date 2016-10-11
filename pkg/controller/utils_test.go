package controller

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	builderconf "github.com/deis/builder/pkg/conf"

	"github.com/arschles/assert"
	deis "github.com/deis/controller-sdk-go"
)

func TestNew(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tmpdir")
	if err != nil {
		t.Fatalf("error creating temp directory (%s)", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("failed to remove builder-key from %s (%s)", tmpDir, err)
		}
	}()

	builderconf.BuilderKeyLocation = filepath.Join(tmpDir, "builder-key")
	data := []byte("testbuilderkey")
	if err := ioutil.WriteFile(builderconf.BuilderKeyLocation, data, 0644); err != nil {
		t.Fatalf("error creating %s (%s)", builderconf.BuilderKeyLocation, err)
	}

	host := "127.0.0.1"
	port := "80"
	cli, err := New(host, port)
	assert.NoErr(t, err)
	assert.Equal(t, cli.ControllerURL.String(), fmt.Sprintf("http://%s:%s/", host, port), "data")
	assert.Equal(t, cli.HooksToken, string(data), "data")
	assert.Equal(t, cli.UserAgent, "deis-builder", "user-agent")

	port = "invalid-port-number"
	if _, err = New(host, port); err == nil {
		t.Errorf("expected error with invalid port number, got nil")
	}
}

func TestNewWithInvalidBuilderKeyPath(t *testing.T) {
	host := "127.0.0.1"
	port := "80"
	_, err := New(host, port)
	assert.True(t, err != nil, "no error received when there should have been")
}

func TestCheckAPICompat(t *testing.T) {
	client := &deis.Client{ControllerAPIVersion: deis.APIVersion}
	err := deis.ErrAPIMismatch

	if apiErr := CheckAPICompat(client, err); apiErr != nil {
		t.Errorf("api errors are non-fatal and should return nil, got '%v'", apiErr)
	}

	err = errors.New("random error")
	if apiErr := CheckAPICompat(client, err); apiErr == nil {
		t.Error("expected error to be returned, got nil")
	}
}
