// Package cleaner is a background process that compares the kubernetes namespace list with the folders in the local git home directory, deleting what's not in the namespace list
package cleaner

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deis/builder/pkg/k8s"
	"github.com/deis/builder/pkg/sshd"
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

// Run starts the deleted app cleaner. Every pollSleepDuration, it compares the result of nsLister.List with the directories in the top level of gitHome on the local file system. On any error, it uses log.Debug to output a human readable description of what happened.
func Run(gitHome string, nsLister k8s.NamespaceLister, repoLock sshd.RepositoryLock, pollSleepDuration time.Duration) error {
	for {
		nsList, err := nsLister.List(labels.Everything(), fields.Everything())
		if err != nil {
			log.Err("Cleaner error listing namespaces (%s)", err)
			continue
		}

		gitDirs, err := localDirs(gitHome)
		if err != nil {
			log.Err("Cleaner error listing local git directories (%s)", err)
		}

		dirsToDelete := getDiff(nsList.Items, gitDirs)
		if len(dirsToDelete) > 0 {
			log.Debug("Cleaner found the following git directories to delete: %s", dirsToDelete)
		} else {
			log.Debug("Cleaner found no git directories to delete")
		}
		for _, dirToDelete := range dirsToDelete {
			if err := repoLock.Lock(dirToDelete, time.Duration(0)); err != nil {
				log.Err("Cleaner error locking repository %s for deletion (%s)", dirToDelete, err)
				continue
			}
			if err := os.RemoveAll(dirToDelete); err != nil {
				log.Err("Cleaner error removing deleted app %s (%s)", dirToDelete, err)
			}
			if err := repoLock.Unlock(dirToDelete, time.Duration(0)); err != nil {
				log.Err("Cleaner error unlocking repository %s for deletion (%s)", dirToDelete, err)
				continue
			}
		}

		time.Sleep(pollSleepDuration)
	}
}
