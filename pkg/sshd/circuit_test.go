package sshd

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	numConcurrents = 1000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestOpenCloseSerial(t *testing.T) {
	c := NewCircuit()
	if c.State() != OpenState {
		t.Fatalf("unexpected initial circuit state. want %s, got %s", OpenState, c.State())
	}
	if b := c.Close(); !b {
		t.Fatalf("tried to close the circuit but it said it was already closed")
	}
	if c.State() != ClosedState {
		t.Fatalf("unexpected circuit state. want %s, got %s", ClosedState, c.State())
	}
	if b := c.Open(); !b {
		t.Fatalf("tried to open the circuit but it said it was already opened")
	}
	if c.State() != OpenState {
		t.Fatalf("unexpected circuit state. want %s, got %s", OpenState, c.State())
	}
}

func TestOpenCloseConcurrent(t *testing.T) {
	c := NewCircuit()
	var wg sync.WaitGroup
	state := uint32(0)
	for i := 0; i < numConcurrents; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := rand.Int() % 2
			if r == 0 {
				c.Open()
				atomic.StoreUint32(&state, OpenState.toUint32())
			} else {
				c.Close()
				atomic.StoreUint32(&state, ClosedState.toUint32())
			}
		}(i)
	}
	wg.Wait()
	if state != c.State().toUint32() {
		t.Fatalf("expected state %d wasn't equal to actual %d", state, c.State().toUint32())
	}
}
