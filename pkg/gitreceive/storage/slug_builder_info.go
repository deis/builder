package storage

import (
	"fmt"

	"github.com/deis/builder/pkg/gitreceive/git"
)

// SlugBuilderInfo contains all of the object storage related information needed to pass to a slug builder
type SlugBuilderInfo struct {
	PushKey string
	PushURL string
	TarKey  string
	TarURL  string
}

// NewSlugBuilderInfo creates and populates a new SlugBuilderInfo based on the given data
func NewSlugBuilderInfo(s3Endpoint, appName, slugName string, gitSha *git.SHA) *SlugBuilderInfo {
	tarKey := fmt.Sprintf("home/%s/tar", slugName)
	// this is where workflow tells slugrunner to download the slug from, so we have to tell slugbuilder to upload it to here
	pushKey := fmt.Sprintf("home/%s/push", fmt.Sprintf("%s:git-%s", appName, gitSha.Short))

	return &SlugBuilderInfo{
		PushKey: pushKey,
		PushURL: fmt.Sprintf("%s/%s", s3Endpoint, pushKey),
		TarKey:  tarKey,
		TarURL:  fmt.Sprintf("%s/git/%s", s3Endpoint, tarKey),
	}
}
