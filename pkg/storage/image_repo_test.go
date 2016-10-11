package storage

import (
	"testing"
)

func TestCreateImageRepo(t *testing.T) {
	params := make(map[string]string)
	if err := CreateImageRepo("", params); err == nil {
		t.Errorf("CreateImageRepo did not fail when no region was provided")
	}
	params["region"] = ""
	if err := CreateImageRepo("", params); err == nil {
		t.Errorf("CreateImageRepo did not fail when no region was provided")
	}
}
