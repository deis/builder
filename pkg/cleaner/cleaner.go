// Package cleaner is a background process that compares the kubernetes namespace list with the folders in the local git home directory, deleting what's not in the namespace list
package cleaner

import (
	"io/ioutil"
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

const (
	dotGitSuffix = ".git"
)

func localDirs(gitHome string) ([]string, error) {
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
		ret = append(ret, filepath.Join(gitHome, nm))
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
		} else {
			lst := make([]string, len(nsList.Items))
			for i, ns := range nsList.Items {
				lst[i] = strings.ToLower(ns.Name)
			}
			log.Debug("Cleaner found namespaces %s", lst)
		}

		gitDirs, err := localDirs(gitHome)
		if err != nil {
			log.Err("Cleaner error listing local git directories (%s)", err)
			continue
		} else {
			log.Debug("Cleaner found local git directories in %s: %s", gitHome, gitDirs)
		}

		gitDirs = stripSuffixes(gitDirs, dotGitSuffix)

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
