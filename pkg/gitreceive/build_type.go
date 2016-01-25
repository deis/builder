package gitreceive

import (
	"fmt"
	"os"
)

type buildType string

func (b buildType) String() string {
	return string(b)
}

const (
	buildTypeProcfile   buildType = "procfile"
	buildTypeDockerfile buildType = "dockerfile"
)

func getBuildTypeForDir(dirName string) buildType {
	_, err := os.Stat(fmt.Sprintf("%s/Dockerfile", dirName))
	if err == nil {
		return buildTypeDockerfile
	}
	return buildTypeProcfile
}
