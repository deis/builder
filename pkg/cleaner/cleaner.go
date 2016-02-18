// Package cleaner is a background process that compares the kubernetes namespace list with the folders in the local git home directory, deleting what's not in the namespace list
package cleaner

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deis/builder/pkg/k8s"
	"github.com/deis/pkg/log"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

func localDirs(gitHome string) ([]string, error) {
	var ret []string
	err := filepath.Walk(gitHome, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return filepath.SkipDir
		}
		ret = append(ret, filepath.Join(gitHome, path))
		return nil
	})

	if err != nil {
		return nil, err
	}
	return ret, nil
}

// getDisjunction gets the items that are in namespaceList and not in dirs or vice versa
func getDisjunction(namespaceList []api.Namespace, dirs []string) []string {
	var ret []string
	namespacesSet := make(map[string]struct{})
	dirsSet := make(map[string]struct{})

	// create sets of the namespaces and dirs
	for _, ns := range namespaceList {
		lowerName := strings.ToLower(ns.Name)
		namespacesSet[lowerName] = struct{}{}
	}

	for _, dir := range dirs {
		lowerName := strings.ToLower(dir)
		dirsSet[lowerName] = struct{}{}
	}

	// get dirs not in the namespaces set
	for _, dir := range dirs {
		lowerName := strings.ToLower(dir)
		if _, ok := namespacesSet[lowerName]; !ok {
			ret = append(ret, lowerName)
		}
	}

	// get namespaces not in the dirs set
	for _, ns := range namespaceList {
		lowerName := strings.ToLower(ns.Name)
		if _, ok := dirsSet[lowerName]; !ok {
			ret = append(ret, lowerName)
		}
	}

	return ret
}

// Run starts the deleted app cleaner. Every pollInterval, it compares the result of nsLister.List with the directories in the top level of gitHome on the local file system. On any error, it uses log.Debug to output a human readable description of what happened.
// TODO: locking mechanism on repositories. Nobody should be able to push to a repo while one is being deleted
func Run(gitHome string, nsLister k8s.NamespaceLister, pollInterval time.Duration) error {
	for {
		nsList, err := nsLister.List(labels.Everything(), fields.Everything())
		if err != nil {
			log.Debug("Cleaner error listing namespaces (%s)", err)
			continue
		}

		gitDirs, err := localDirs(gitHome)
		if err != nil {
			log.Debug("Cleaner error listing local git directories (%s)", err)
		}

		disjunctions := getDisjunction(nsList.Items, gitDirs)
		for _, disj := range disjunctions {
			if err := os.RemoveAll(disj); err != nil {
				log.Debug("Cleaner error removing deleted app %s (%s)", disj, err)
			}
		}

		time.Sleep(pollInterval)
	}
}
