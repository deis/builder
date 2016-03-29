package gitreceive

import (
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/gitreceive/git"
)

func TestPushKey(t *testing.T) {
	sha, err := git.NewSha(rawSha)
	if err != nil {
		t.Fatalf("error building git sha (%s)", err)
	}
	sbi := NewSlugBuilderInfo(appName + ":git-" + sha.Short())
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
	slugName := appName + ":git-" + sha.Short()
	sbi := NewSlugBuilderInfo(slugName)
	expectedTarKey := "home/" + slugName + "/tar"
	if sbi.TarKey() != expectedTarKey {
		t.Errorf("tar key %s didn't match expected %s", sbi.TarKey(), expectedTarKey)
	}
}

func TestAbsoluteSlugObjectKey(t *testing.T) {
	sha, err := git.NewSha(rawSha)
	assert.NoErr(t, err)
	sbi := NewSlugBuilderInfo(appName + ":git-" + sha.Short())
	assert.Equal(t, sbi.AbsoluteSlugObjectKey(), sbi.PushKey()+"/"+slugTGZName, "absolute slug key")
}
