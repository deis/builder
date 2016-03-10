package storage

import (
	"fmt"

	"github.com/deis/builder/pkg/gitreceive/git"
)

const (
	slugTGZName = "slug.tgz"
)

// SlugBuilderInfo contains all of the object storage related information needed to pass to a slug builder
type SlugBuilderInfo struct {
	pushKey string
	pushURL string
	tarKey  string
	tarURL  string
}

// NewSlugBuilderInfo creates and populates a new SlugBuilderInfo based on the given data
func NewSlugBuilderInfo(s3Endpoint *Endpoint, bucket, appName, slugName string, gitSha *git.SHA) *SlugBuilderInfo {
	tarKey := fmt.Sprintf("home/%s/tar", slugName)
	// this is where the controller tells slugrunner to download the slug from, so we have to tell slugbuilder to upload it to here
	pushKey := fmt.Sprintf("home/%s:git-%s/push", appName, gitSha.Short())

	return &SlugBuilderInfo{
		pushKey: pushKey,
		pushURL: fmt.Sprintf("%s/%s/%s", s3Endpoint.FullURL(), bucket, pushKey),
		tarKey:  tarKey,
		tarURL:  fmt.Sprintf("%s/%s/%s", s3Endpoint.FullURL(), bucket, tarKey),
	}
}

// PushKey returns the object storage key that the slug builder will store the slug in. The returned value only contains the path to the folder, not including the final filename.
func (s SlugBuilderInfo) PushKey() string { return s.pushKey }

// PushURL returns the complete object storage URL that the slug builder will store the slug in. The returned value only contains the URL to the folder, not including the final filename.
func (s SlugBuilderInfo) PushURL() string { return s.pushURL }

// TarKey returns the object storage key from which the slug builder will download for the tarball (from which it uses to build the slug). The returned value only contains the path to the folder, not including the final filename.
func (s SlugBuilderInfo) TarKey() string { return s.tarKey }

// TarURL returns the complete object storage URL that the slug builder will download the tarball from. The returned value only contains the URL to the folder, not including the final filename.
func (s SlugBuilderInfo) TarURL() string { return s.tarURL }

// AbsoluteSlugObjectKey returns the PushKey plus the final filename of the slug
func (s SlugBuilderInfo) AbsoluteSlugObjectKey() string { return s.PushKey() + "/" + slugTGZName }

// AbsoluteProcfileKey returns the PushKey plus the standard procfile name
func (s SlugBuilderInfo) AbsoluteProcfileKey() string { return s.PushKey() + "/Procfile" }

// AbsoluteSlugURL returns the PushURL plus the final filename of the slug
func (s SlugBuilderInfo) AbsoluteSlugURL() string {
	return s.PushURL() + "/" + slugTGZName
}
