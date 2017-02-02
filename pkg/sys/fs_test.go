package sys

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

const testFilename string = "sys-fs-tests"

var expected = []byte("temporary file's content")

func TestRealFS(t *testing.T) {
	fs := RealFS()

	if _, err := fs.ReadFile("/this-file-does-not-exist"); err == nil {
		t.Error("expected to receive error when real file does not exist")
	}

	tmpfile, err := ioutil.TempFile("", testFilename)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(expected); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	actual, err := fs.ReadFile(tmpfile.Name())
	if err != nil {
		t.Errorf("expected error to be nil, got '%v'", err)
	}
	if string(actual) != string(expected) {
		t.Errorf("expected '%s', got '%s'", string(expected), string(actual))
	}

	if err := fs.RemoveAll(tmpfile.Name()); err != nil {
		t.Errorf("Could not remove %s (%v)", tmpfile.Name(), err)
	}

	if _, err := os.Stat(tmpfile.Name()); err == nil {
		t.Errorf("%s was not removed from the real filesystem!", tmpfile.Name())
	}
}

func TestFakeFS(t *testing.T) {
	fs := NewFakeFS()

	if _, err := fs.ReadFile(testFilename); err == nil {
		t.Error("expected to receive error when fake file does not exist")
	}

	fs.Files[testFilename] = expected

	actual, err := fs.ReadFile(testFilename)
	if err != nil {
		t.Errorf("expected error to be nil, got '%v'", err)
	}
	if string(actual) != string(expected) {
		t.Errorf("expected '%s', got '%s'", string(expected), string(actual))
	}

	if err := fs.RemoveAll(testFilename); err != nil {
		t.Errorf("Could not remove %s (%v)", testFilename, err)
	}
	if _, ok := fs.Files[testFilename]; ok {
		t.Errorf("%s was not removed from the fake filesystem!", testFilename)
	}

	err = fs.RemoveAll(testFilename)
	if err == nil {
		t.Errorf("Expected removing %s a second time would fail", testFilename)
	}
	expectedError := fmt.Sprintf("Fake file %s not found", testFilename)
	if err.Error() != expectedError {
		t.Errorf("expected '%s', got '%s'", expectedError, err.Error())
	}
}
