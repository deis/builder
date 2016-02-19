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
	assert.NoErr(t, err)

	expectedPackages := map[string]int{
		pkgDir + "/cleaner":    1,
		pkgDir + "/conf":       1,
		pkgDir + "/controller": 1,
		pkgDir + "/env":        1,
		pkgDir + "/git":        1,
		pkgDir + "/gitreceive": 1,
		pkgDir + "/healthsrv":  1,
		pkgDir + "/k8s":        1,
		pkgDir + "/sshd":       1,
	}

	actualPackages := map[string]int{}
	for _, lDir := range lDirs {
		actualPackages[lDir]++
	}
	assert.Equal(t, len(actualPackages), len(expectedPackages), "number of packages")
	for actualPackageName, actualNum := range actualPackages {
		if actualNum != 1 {
			t.Errorf("found %d %s packages", actualNum, actualPackageName)
			continue
		}
		expectedNum, ok := expectedPackages[actualPackageName]
		if !ok {
			t.Errorf("found unexpected package %s", actualPackageName)
			continue
		}
		if actualNum != expectedNum {
			t.Errorf("found %d %s packages, expected %d", actualNum, actualPackageName, expectedNum)
			continue
		}
	}
}

func TestStripSuffixes(t *testing.T) {
	strs := []string{"a.git", "b.git", "c.git", "d"}
	newStrs := stripSuffixes(strs, dotGitSuffix)
	assert.Equal(t, len(newStrs), len(strs), "number of strings")
	for _, str := range newStrs {
		assert.False(t, strings.HasSuffix(str, dotGitSuffix), "string %s has suffix %s", str, dotGitSuffix)
	}
}
