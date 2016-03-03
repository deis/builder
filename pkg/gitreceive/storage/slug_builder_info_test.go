package storage

import (
	"testing"

	"github.com/deis/builder/pkg/gitreceive/git"
)

const (
	rawSha   = "c3b4e4ba8b7267226ff02ad07a3a2cca9c9237de"
	appName  = "myapp"
	slugName = "myslug"
	bucket   = "git"
)

var (
	s3Endpoint = &Endpoint{URLStr: "10.1.2.3:9090", Secure: false}
)

func TestS3Endpoint(t *testing.T) {
	sha, err := git.NewSha(rawSha)
	if err != nil {
		t.Fatalf("error building git sha (%s)", err)
	}
	sbi := NewSlugBuilderInfo(s3Endpoint, bucket, appName, slugName, sha)

	expectedPushURL := s3Endpoint.FullURL() + "/" + bucket + "/" + sbi.PushKey()
	if sbi.PushURL() != expectedPushURL {
		t.Errorf("push URL %s didn't match expected %s", sbi.PushURL(), expectedPushURL)
	}
	expectedTarURL := s3Endpoint.FullURL() + "/" + bucket + "/" + sbi.TarKey()
	if sbi.TarURL() != expectedTarURL {
		t.Errorf("tar URL %s didn't match expected %s", sbi.TarURL(), expectedTarURL)
	}
}

func TestPushKey(t *testing.T) {
	sha, err := git.NewSha(rawSha)
	if err != nil {
		t.Fatalf("error building git sha (%s)", err)
	}
	sbi := NewSlugBuilderInfo(s3Endpoint, bucket, appName, slugName, sha)
	expectedPushKey := "home/" + appName + ":git-" + sha.Short() + "/push"
	if sbi.PushKey() != expectedPushKey {
		t.Errorf("push key %s didn't match expected %s", sbi.PushKey(), expectedPushKey)
	}
}

func TestTarKey(t *testing.T) {
	sha, err := git.NewSha(rawSha)
	if err != nil {
		t.Fatalf("error building git sha (%s)", err)
	}
	sbi := NewSlugBuilderInfo(s3Endpoint, bucket, appName, slugName, sha)
	expectedTarKey := "home/" + slugName + "/tar"
	if sbi.TarKey() != expectedTarKey {
		t.Errorf("tar key %s didn't match expected %s", sbi.TarKey(), expectedTarKey)
	}
}
