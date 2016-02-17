package k8s

import (
	"k8s.io/kubernetes/pkg/watch"
)

// PodStates is a state machine that can notify an arbitrary number of clients about the states of a given watch.Interface.
type PodStates struct {
	watcher  watch.Interface
	watchers *eventWatchers
}

// NewPodStates creates a new PodStates instance that watches watcher for incoming states. It populates the initial set of watchers with initialWatchers
func NewPodStates(watcher watch.Interface, initialWatchers map[watch.EventType][]chan<- watch.Event) *PodStates {
	watchers := newEventWatchers(initialWatchers)
	ps := &PodStates{
		watcher:  watcher,
		watchers: watchers,
	}
	go startWatchLoop(watcher, watchers)
	return ps
}

func (p *PodStates) Stop() {
	p.watchers.stop()
}

func startWatchLoop(watcher watch.Interface, watchers *eventWatchers) {
	for evt := range watcher.ResultChan() {
		watchers.broadcast(evt.Type, evt)
	}
}
