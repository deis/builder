package k8s

import (
	"sync"
	"testing"
	"time"

	"github.com/arschles/assert"
	"k8s.io/kubernetes/pkg/watch"
)

const (
	broadcastTimeout = 200 * time.Millisecond
)

func TestAddWatcher(t *testing.T) {
	ch1 := make(chan watch.Event)
	initial := map[watch.EventType][]chan<- watch.Event{
		watch.Added: []chan<- watch.Event{ch1},
	}
	w := newEventWatchers(initial)
	ch2 := make(chan watch.Event)

	w.add(watch.Added, ch2)
	assert.Equal(t, len(w.watchers[watch.Added]), 2, "length of watch.Added slice")

	w.add(watch.Deleted, ch2)
	assert.Equal(t, len(w.watchers[watch.Deleted]), 1, "length of watch.Deleted slice")
}

func TestAddWatcherConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	ch := make(chan watch.Event)
	wg.Add(2)
	w := newEventWatchers(make(map[watch.EventType][]chan<- watch.Event))
	go func() {
		defer wg.Done()
		w.add(watch.Added, ch)
	}()
	go func() {
		defer wg.Done()
		w.add(watch.Deleted, ch)
	}()
	wg.Wait()
	assert.Equal(t, len(w.watchers[watch.Added]), 1, "length of watch.Added slice")
	assert.Equal(t, len(w.watchers[watch.Deleted]), 1, "length of watch.Deleted slice")
}

func TestWatchersBroadcast(t *testing.T) {
	w := newEventWatchers(make(map[watch.EventType][]chan<- watch.Event))
	ch1 := make(chan watch.Event)
	ch2 := make(chan watch.Event)
	w.add(watch.Added, ch1)
	w.add(watch.Deleted, ch2)
	addedEvt := watch.Event{}
	go func() {
		w.broadcast(watch.Added, addedEvt)
	}()

	select {
	case e := <-ch1:
		assert.Equal(t, e, addedEvt, "returned added event")
	case <-time.After(broadcastTimeout):
		t.Errorf("didn't receive on added event chan within %s", broadcastTimeout)
	}

	deletedEvt := watch.Event{}
	go func() {
		w.broadcast(watch.Deleted, deletedEvt)
	}()
	select {
	case e := <-ch2:
		assert.Equal(t, e, deletedEvt, "returned deleted event")
	case <-time.After(broadcastTimeout):
		t.Errorf("didn't receive on deleted event chan within %s", broadcastTimeout)
	}
}

func TestWatchersBroadcastStop(t *testing.T) {
	w := newEventWatchers(make(map[watch.EventType][]chan<- watch.Event))
	w.add(watch.Added, make(chan watch.Event))
	sig := make(chan struct{})
	go func() {
		defer close(sig)
		w.broadcast(watch.Added, watch.Event{})
	}()

	select {
	case <-sig:
		t.Errorf("broadcast finished without a receiving channel")
	case <-time.After(broadcastTimeout):
	}
	w.stop()
	select {
	case <-sig:
	case <-time.After(broadcastTimeout):
		t.Errorf("broadcast didn't finish after close")
	}

}
