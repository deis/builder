// Package cleaner is a background process that compares the kubernetes namespace list with the folders in the local git home directory, deleting what's not in the namespace list
package cleaner

import (
	"log"
	"os"

	"k8s.io/kubernetes/pkg/api"

	"github.com/deis/builder/pkg/k8s"
)

const (
	dotGitSuffix = ".git"
)

// Run starts the deleted app cleaner. Every pollSleepDuration, it compares the result of nsLister.List with the directories in the top level of gitHome on the local file system. On any error, it uses log messages to output a human readable description of what happened.
func Run(gitHome string, nsLister k8s.NamespaceWatcher) error {

	watcher, err := nsLister.Watch(nil, nil, "")
	if err != nil {
		log.Printf("unable to get watch events (%s)", err)
	}
	for {
		event := <-watcher.ResultChan()
		if event.Type == "DELETED" {
			namespace := event.Object.(*api.Namespace)
			appToDelete := gitHome + "/" + namespace.ObjectMeta.Name + dotGitSuffix
			if err := os.RemoveAll(appToDelete); err != nil {
				log.Printf("Cleaner error removing deleted app %s (%s)", appToDelete, err)
			}
		}

	}

}
