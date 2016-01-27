package storage

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/deis/builder/pkg/gitreceive/git"
)

type SlugBuilderInfo struct {
	PushKey string
	PushURL string
	TarKey  string
	TarURL  string
}

func NewSlugBuilderInfo(s3Client *s3.S3, appName, slugName string, gitSha *git.SHA) *SlugBuilderInfo {
	tarKey := fmt.Sprintf("home/%s/tar", slugName)
	// this is where workflow tells slugrunner to download the slug from, so we have to tell slugbuilder to upload it to here
	pushKey := fmt.Sprintf("home/%s/push", fmt.Sprintf("%s:git-%s", appName, gitSha.Short))

	return &SlugBuilderInfo{
		PushKey: pushKey,
		PushURL: fmt.Sprintf("%s/%s", s3Client.Endpoint, pushKey),
		TarKey:  tarKey,
		TarURL:  fmt.Sprintf("%s/git/%s", s3Client.Endpoint, tarKey),
	}
}
