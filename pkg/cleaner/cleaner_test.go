package cleaner

import (
	"testing"

	"time"

	"github.com/deis/builder/pkg/sys"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type namespace struct {
}

// Watch returns a watch.Interface that watches the requested namespaces.
func (r *namespace) Watch(label labels.Selector, field fields.Selector, resourceversion string) (watch.Interface, error) {
	nst := watch.NewFake()
	go func() {
		nst.Add(&api.Namespace{ObjectMeta: api.ObjectMeta{Name: "dir1"}})
		nst.Modify(&api.Namespace{ObjectMeta: api.ObjectMeta{Name: "dir1"}})
		nst.Delete(&api.Namespace{ObjectMeta: api.ObjectMeta{Name: "dir1"}})
		nst.Modify(&api.Namespace{ObjectMeta: api.ObjectMeta{Name: "dir2"}})
		nst.Delete(nil)
		nst.Add(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "dir3"}})
		nst.Delete(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "dir3"}})
		nst.Stop()
	}()
	return nst, nil
}

func (r *namespace) IsAnAPIObject() {}

func TestCleanerRun(t *testing.T) {
	ns := &namespace{}
	fs := sys.NewFakeFS()
	dirhome := "/home/git"
	fs.Files["/home/git/dir1.git"] = []byte{}
	fs.Files["/home/git/dir2.git"] = []byte{}
	fs.Files["/home/git/dir3.git"] = []byte{}
	go Run(dirhome, ns, fs)
	time.Sleep(5 * time.Second)
	// Namespace with name dir1 got deleted directory should be deleted
	_, ok := fs.Files["/home/git/dir1.git"]
	if ok {
		t.Fatal("Test failed: Deleting a namespace should delete respective direcotry")
	}
	// Namespace with name dir2 got modified directory should not be deleted
	_, ok = fs.Files["/home/git/dir2.git"]
	if !ok {
		t.Fatal("Test failed: Modyfiying a namespace should not delete respective direcotry")
	}
	// Pod with name dir3 got deleted directory should not be deleted
	_, ok = fs.Files["/home/git/dir3.git"]
	if !ok {
		t.Fatal("Test failed: Deleting a pod should not delete respective direcotry")
	}
}
