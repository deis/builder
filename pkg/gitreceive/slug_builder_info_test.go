package gitreceive

import (
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/git"
)

const (
	rawSha  = "c3b4e4ba8b7267226ff02ad07a3a2cca9c9237de"
	appName = "myapp"
)

func TestPushKey(t *testing.T) {
	sha, err := git.NewSha(rawSha)
	if err != nil {
		t.Fatalf("error building git sha (%s)", err)
	}
	sbi := NewSlugBuilderInfo(appName, sha.Short())
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
	sbi := NewSlugBuilderInfo(appName, sha.Short())
	expectedTarKey := "home/" + appName + ":git-" + sha.Short() + "/tar"
	if sbi.TarKey() != expectedTarKey {
		t.Errorf("tar key %s didn't match expected %s", sbi.TarKey(), expectedTarKey)
	}
}

func TestCacheKey(t *testing.T) {
	sha, err := git.NewSha(rawSha)
	if err != nil {
		t.Fatalf("error building git sha (%s)", err)
	}
	sbi := NewSlugBuilderInfo(appName, sha.Short())
	expectedCacheKey := "home/" + appName + "/cache"
	if sbi.CacheKey() != expectedCacheKey {
		t.Errorf("tar key %s didn't match expected %s", sbi.CacheKey(), expectedCacheKey)
	}
}

func TestAbsoluteSlugObjectKey(t *testing.T) {
	sha, err := git.NewSha(rawSha)
	assert.NoErr(t, err)
	sbi := NewSlugBuilderInfo(appName, sha.Short())
	assert.Equal(t, sbi.AbsoluteSlugObjectKey(), sbi.PushKey()+"/"+slugTGZName, "absolute slug key")
}
