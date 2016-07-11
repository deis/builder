package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arschles/assert"
)

func TestCreatePreReceiveHook(t *testing.T) {
	const gitHome = "TestGitHome"
	gopath := os.Getenv("GOPATH")
	repoPath := filepath.Join(gopath, "src", "github.com", "deis", "builder", "testdata")
	assert.NoErr(t, createPreReceiveHook(gitHome, repoPath))
	hookBytes, err := ioutil.ReadFile(filepath.Join(repoPath, "hooks", "pre-receive"))
	assert.NoErr(t, err)
	hookStr := string(hookBytes)
	gitHomeIdx := strings.Index(hookStr, fmt.Sprintf("GIT_HOME=%s", gitHome))
	assert.False(t, gitHomeIdx == -1, "GIT_HOME was not found")
}
