// Package cleaner is a background process that compares the kubernetes namespace list with the
// folders in the local git home directory, deleting what's not in the namespace list.
package cleaner

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/deis/builder/pkg/gitreceive"
	"github.com/deis/builder/pkg/k8s"
	"github.com/deis/builder/pkg/sys"
	"github.com/deis/pkg/log"
	"github.com/docker/distribution/context"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

const (
	dotGitSuffix = ".git"
)

// localDirs returns all of the local directories immediately under gitHome that filter returns true for.
// filter will receive only the names of each of the top level directories (not their fully qualified paths), and should return true if it should be included in the output
func localDirs(gitHome string, filter func(string) bool) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(gitHome)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, fileInfo := range fileInfos {
		nm := fileInfo.Name()
		if len(nm) <= 0 || nm == "." || !fileInfo.IsDir() {
			continue
		}
		if filter(nm) {
			ret = append(ret, nm)
		}
	}
	return ret, nil
}

// getDiff gets the directories that are not in namespaceList
func getDiff(namespaceList []api.Namespace, dirs []string) []string {
	var ret []string

	// create a set of lowercase namespace names
	namespacesSet := make(map[string]struct{})
	for _, ns := range namespaceList {
		lowerName := strings.ToLower(ns.Name)
		namespacesSet[lowerName] = struct{}{}
	}

	// get dirs not in the namespaces set
	for _, dir := range dirs {
		lowerName := strings.ToLower(dir)
		if _, ok := namespacesSet[lowerName]; !ok {
			ret = append(ret, lowerName)
		}
	}

	return ret
}

func stripSuffixes(strs []string, suffix string) []string {
	ret := make([]string, len(strs))
	for i, str := range strs {
		idx := strings.LastIndex(str, suffix)
		if idx >= 0 {
			ret[i] = str[:idx]
		} else {
			ret[i] = str
		}
	}
	return ret
}

func dirHasGitSuffix(dir string) bool {
	return strings.HasSuffix(dir, dotGitSuffix)
}

func deleteFromObjectStore(app string, storageDriver storagedriver.StorageDriver) error {

	cacheKey := fmt.Sprintf(gitreceive.CacheKeyPattern, app)

	// if cache file exists, delete it
	if _, err := storageDriver.Stat(context.Background(), cacheKey); err == nil {
		log.Info("Cleaner deleting cache %s for app %s", cacheKey, app)
		if err := storageDriver.Delete(context.Background(), cacheKey); err != nil {
			return err
		}
	}

	// delete all slug files matching app
	objs, err := storageDriver.List(context.Background(), "home")
	if err != nil {
		return err
	}

	// regex needs prepended / to match output of List()
	gitRegex, err := regexp.Compile(`^/` + fmt.Sprintf(gitreceive.GitKeyPattern, app, ".{8}") + "$")
	if err != nil {
		return err
	}

	for _, obj := range objs {
		if gitRegex.MatchString(obj) {
			log.Info("Cleaner deleting slug %s for app %s", obj, app)
			if err := storageDriver.Delete(context.Background(), obj); err != nil {
				return err
			}
		}
	}
	return nil
}

// Run starts the deleted app cleaner. Every pollSleepDuration, it compares the result of nsLister.List with the directories in the top level of gitHome on the local file system.
// On any error, it uses log messages to output a human readable description of what happened.
func Run(gitHome string, nsLister k8s.NamespaceLister, fs sys.FS, pollSleepDuration time.Duration, storageDriver storagedriver.StorageDriver) error {
	for {
		nsList, err := nsLister.List(api.ListOptions{LabelSelector: labels.Everything(), FieldSelector: fields.Everything()})
		if err != nil {
			log.Err("Cleaner error listing namespaces (%s)", err)
			continue
		}

		gitDirs, err := localDirs(gitHome, dirHasGitSuffix)
		if err != nil {
			log.Err("Cleaner error listing local git directories (%s)", err)
			continue
		}

		gitDirs = stripSuffixes(gitDirs, dotGitSuffix)

		appsToDelete := getDiff(nsList.Items, gitDirs)

		for _, appToDelete := range appsToDelete {
			dirToDelete := filepath.Join(gitHome, appToDelete+dotGitSuffix)
			if err := fs.RemoveAll(dirToDelete); err != nil {
				log.Err("Cleaner error removing local files for deleted app %s (%s)", dirToDelete, err)
			}
			if err := deleteFromObjectStore(appToDelete, storageDriver); err != nil {
				log.Err("Cleaner error removing object store files for deleted app %s (%s)", appToDelete, err)
			}
		}

		time.Sleep(pollSleepDuration)
	}
}
