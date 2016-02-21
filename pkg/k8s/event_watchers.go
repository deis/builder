package k8s

import (
	"sync"

	"k8s.io/kubernetes/pkg/watch"
)

// eventWatchers manages sets of channels that are registered to receive on a given watch.EventType
type eventWatchers struct {
	stopSendCh chan struct{}
	rwm        *sync.RWMutex
	watchers   map[watch.EventType][]chan<- watch.Event
}

func newEventWatchers(initial map[watch.EventType][]chan<- watch.Event) *eventWatchers {
	return &eventWatchers{stopSendCh: make(chan struct{}), rwm: &sync.RWMutex{}, watchers: initial}
}

// add registers ch to receive in the future when a given evtType is received. Future calls to broadcast with evtType will cause ch to recieve from this point on. This function will block if any broadcast or other add operation is in progress
func (w *eventWatchers) add(evtType watch.EventType, ch chan<- watch.Event) {
	w.rwm.Lock()
	defer w.rwm.Unlock()
	w.watchers[evtType] = append(w.watchers[evtType], ch)
}

// broadcast sends to all channels watching on evtType. Each send happens in its own goroutine, but this function will block until all sends are complete. It will block all add operations also block all other operations on w until all sends complete or w.stop() is called
func (w *eventWatchers) broadcast(evtType watch.EventType, evt watch.Event) {
	var wg sync.WaitGroup
	w.rwm.RLock()
	defer w.rwm.RUnlock()
	for _, ch := range w.watchers[evtType] {
		wg.Add(1)
		go func(c chan<- watch.Event) {
			defer wg.Done()
			select {
			case c <- evt:
			case <-w.stopSendCh:
			}
		}(ch)
	}
	wg.Wait()
}

// stop halts all currently ongoing broadcast operations, causes all current broadcast operations to immediately return and renders w unusable after it returns.
func (w *eventWatchers) stop() {
	close(w.stopSendCh)
}
