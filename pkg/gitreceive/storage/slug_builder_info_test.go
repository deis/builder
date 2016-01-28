package storage

import (
	"testing"

	"github.com/deis/builder/pkg/gitreceive/git"
)

const (
	rawSha     = "c3b4e4ba8b7267226ff02ad07a3a2cca9c9237de"
	s3Endpoint = "http://10.1.2.3:9090"
	appName    = "myapp"
	slugName   = "myslug"
)

func TestS3Endpoint(t *testing.T) {
	sha, err := git.NewSha(rawSha)
	if err != nil {
		t.Fatalf("error building git sha (%s)", err)
	}
	sbi := NewSlugBuilderInfo(s3Endpoint, appName, slugName, sha)

	expectedPushURL := s3Endpoint + "/git/" + sbi.PushKey()
	if sbi.PushURL() != expectedPushURL {
		t.Errorf("push URL %s didn't match expected %s", sbi.PushURL(), expectedPushURL)
	}
	expectedTarURL := s3Endpoint + "/git/" + sbi.TarKey()
	if sbi.TarURL() != expectedTarURL {
		t.Errorf("tar URL %s didn't match expected %s", sbi.TarURL(), expectedTarURL)
	}
}
