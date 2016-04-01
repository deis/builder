// Package cleaner is a background process that compares the kubernetes namespace list with the
// folders in the local git home directory, deleting what's not in the namespace list.
package cleaner

import (
	"log"

	"k8s.io/kubernetes/pkg/api"

	"github.com/deis/builder/pkg/k8s"
	"github.com/deis/builder/pkg/sys"
)

const (
	dotGitSuffix = ".git"
)

// Run starts the deleted app cleaner. This function listens to the Kubernetes event stream for
// events that indicate namespaces that were `DELETED`.
// If the namespace name matches a folder on the local filesystem, this func deletes that folder.
// Note that this function blocks until the watcher returned by `nsLister.Watch` is closed, so
// you should launch it in a goroutine.
func Run(gitHome string, nsLister k8s.NamespaceWatcher, fs sys.FS) error {
	watcher, err := nsLister.Watch(nil, nil, "")
	if err != nil {
		log.Printf("unable to get watch events (%s)", err)
	}
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				break
			}
			if event.Type == "DELETED" {
				switch event.Object.(type) {
				case (*api.Namespace):
					namespace := event.Object.(*api.Namespace)
					appToDelete := gitHome + "/" + namespace.ObjectMeta.Name + dotGitSuffix
					if err := fs.RemoveAll(appToDelete); err != nil {
						log.Printf("Cleaner error removing deleted app %s (%s)", appToDelete, err)
					}
				default:
				}
			}
		}
	}

}
