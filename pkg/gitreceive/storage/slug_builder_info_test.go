package storage

import (
	"strings"
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
	if !strings.HasPrefix(sbi.PushURL, s3Endpoint) {
		t.Errorf("push URL %s didn't have expected s3 endpoint prefix %s", sbi.PushURL, s3Endpoint)
	}
	if !strings.HasPrefix(sbi.TarURL, s3Endpoint) {
		t.Errorf("tar URL %s didn't have expected s3 endpoint prefix %s", sbi.TarURL, s3Endpoint)
	}
}
