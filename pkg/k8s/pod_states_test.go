package k8s

import (
	"sync"
	"testing"
	"time"

	"github.com/arschles/assert"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"
)

type watchLoopCase struct {
	evt watch.Event
	ch  chan watch.Event
}

func TestStartWatchLoop(t *testing.T) {
	fake := watch.NewFake()
	defer fake.Stop()
	watchers := newEventWatchers(make(map[watch.EventType][]chan<- watch.Event))
	go startWatchLoop(fake, watchers)
	cases := []watchLoopCase{
		watchLoopCase{evt: watch.Event{Type: watch.Added, Object: &api.Pod{}}, ch: make(chan watch.Event)},
		watchLoopCase{evt: watch.Event{Type: watch.Deleted, Object: &api.Pod{}}, ch: make(chan watch.Event)},
		watchLoopCase{evt: watch.Event{Type: watch.Modified, Object: &api.Pod{}}, ch: make(chan watch.Event)},
		watchLoopCase{evt: watch.Event{Type: watch.Error, Object: &api.Pod{}}, ch: make(chan watch.Event)},
	}
	var wg sync.WaitGroup
	for _, c := range cases {
		watchers.add(c.evt.Type, c.ch)
		wg.Add(2)
		go func(wlc watchLoopCase) {
			defer wg.Done()
			fake.Action(wlc.evt.Type, wlc.evt.Object)
		}(c)
		go func(wlc watchLoopCase) {
			defer wg.Done()
			select {
			case evt := <-wlc.ch:
				assert.Equal(t, evt.Type, wlc.evt.Type, "received event type")
			case <-time.After(broadcastTimeout):
				t.Errorf("event %s wasn't broadcasted after %s", wlc.evt.Type, broadcastTimeout)
			}
		}(c)
	}
	wg.Wait()
}
