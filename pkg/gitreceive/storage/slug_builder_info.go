package storage

import (
	"fmt"

	"github.com/deis/builder/pkg/gitreceive/git"
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
	// this is where workflow tells slugrunner to download the slug from, so we have to tell slugbuilder to upload it to here
	pushKey := fmt.Sprintf("home/%s:git-%s/push", appName, gitSha.Short())

	return &SlugBuilderInfo{
		pushKey: pushKey,
		pushURL: fmt.Sprintf("%s/%s/%s", s3Endpoint.FullURL(), bucket, pushKey),
		tarKey:  tarKey,
		tarURL:  fmt.Sprintf("%s/%s/%s", s3Endpoint.FullURL(), bucket, tarKey),
	}
}

func (s SlugBuilderInfo) PushKey() string { return s.pushKey }
func (s SlugBuilderInfo) PushURL() string { return s.pushURL }
func (s SlugBuilderInfo) TarKey() string  { return s.tarKey }
func (s SlugBuilderInfo) TarURL() string  { return s.tarURL }
