package cleaner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arschles/assert"
	"k8s.io/kubernetes/pkg/api"
)

func TestGetDiff(t *testing.T) {
	nsList := []api.Namespace{
		api.Namespace{ObjectMeta: api.ObjectMeta{Name: "app1"}},
		api.Namespace{ObjectMeta: api.ObjectMeta{Name: "app2"}},
	}
	dirList := []string{"app1", "app3"}
	diff := getDiff(nsList, dirList)
	assert.Equal(t, len(diff), 1, "number of items in the disjunction")
}

func TestLocalDirs(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoErr(t, err)
	pkgDir, err := filepath.Abs(wd + "/..")
	assert.NoErr(t, err)
	lDirs, err := localDirs(pkgDir)
	for _, dir := range lDirs {
		rel, err := filepath.Rel(pkgDir, dir)
		assert.NoErr(t, err)
		spl := strings.Split(rel, "/")
		assert.Equal(t, len(spl), 1, "directory depth")
	}
}
