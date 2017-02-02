package gitreceive

import (
	"fmt"
)

const (
	slugTGZName = "slug.tgz"
	// CacheKeyPattern is the template for location cache dirs.
	CacheKeyPattern = "home/%s/cache"
	// GitKeyPattern is the template for storing git key files.
	GitKeyPattern = "home/%s:git-%s"
)

// SlugBuilderInfo contains all of the object storage related information needed to pass to a
// slug builder.
type SlugBuilderInfo struct {
	pushKey        string
	tarKey         string
	cacheKey       string
	disableCaching bool
}

// NewSlugBuilderInfo creates and populates a new SlugBuilderInfo based on the given data
func NewSlugBuilderInfo(appName string, shortSha string, disableCaching bool) *SlugBuilderInfo {
	basePath := fmt.Sprintf(GitKeyPattern, appName, shortSha)
	tarKey := fmt.Sprintf("%s/tar", basePath)
	// this is where workflow tells slugrunner to download the slug from, so we have to tell slugbuilder to upload it to here
	pushKey := fmt.Sprintf("%s/push", basePath)

	cacheKey := fmt.Sprintf(CacheKeyPattern, appName)

	return &SlugBuilderInfo{
		pushKey:        pushKey,
		tarKey:         tarKey,
		cacheKey:       cacheKey,
		disableCaching: disableCaching,
	}
}

// PushKey returns the object storage key that the slug builder will store the slug in.
// The returned value only contains the path to the folder, not including the final filename.
func (s SlugBuilderInfo) PushKey() string { return s.pushKey }

// TarKey returns the object storage key from which the slug builder will download for the tarball
// (from which it uses to build the slug). The returned value only contains the path to the
// folder, not including the final filename.
func (s SlugBuilderInfo) TarKey() string { return s.tarKey }

// CacheKey returns the object storage key that the slug builder will use to store the cache in
// it's application specific and persisted between deploys (doesn't contain git-sha)
func (s SlugBuilderInfo) CacheKey() string { return s.cacheKey }

// DisableCaching dictates whether or not the slugbuilder should persist the buildpack cache.
func (s SlugBuilderInfo) DisableCaching() bool { return s.disableCaching }

// AbsoluteSlugObjectKey returns the PushKey plus the final filename of the slug.
func (s SlugBuilderInfo) AbsoluteSlugObjectKey() string { return s.PushKey() + "/" + slugTGZName }

// AbsoluteProcfileKey returns the PushKey plus the standard procfile name.
func (s SlugBuilderInfo) AbsoluteProcfileKey() string { return s.PushKey() + "/Procfile" }
