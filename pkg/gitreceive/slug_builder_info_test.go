package gitreceive

import (
	"testing"

	"github.com/arschles/assert"
)

func TestSlugBuilderInfo(t *testing.T) {
	sbi := NewSlugBuilderInfo("myapp", "c3b4e4ba", false)
	assert.Equal(t, "home/myapp:git-c3b4e4ba/push", sbi.PushKey(), "key")
	assert.Equal(t, "home/myapp:git-c3b4e4ba/tar", sbi.TarKey(), "key")
	assert.Equal(t, "home/myapp/cache", sbi.CacheKey(), "key")
	assert.Equal(t, "home/myapp:git-c3b4e4ba/push/slug.tgz", sbi.AbsoluteSlugObjectKey(), "key")
	assert.Equal(t, "home/myapp:git-c3b4e4ba/push/Procfile", sbi.AbsoluteProcfileKey(), "key")
	assert.Equal(t, false, sbi.DisableCaching(), "key")
}
