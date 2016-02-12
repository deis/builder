package sshd

import (
	"fmt"
	"sync/atomic"
)

// CircuitState represents the state of a Circuit
type CircuitState uint32

const (
	// OpenState indicates that the circuit is in the open state, and thus non-functional
	OpenState CircuitState = 0
	// ClosedState indicates that the circuit is in the closed state, and thus functional
	ClosedState CircuitState = 1
)

// String is the fmt.Stringer interface implementation
func (c CircuitState) String() string {
	if c == OpenState {
		return "OPEN"
	} else if c == ClosedState {
		return "CLOSED"
	} else {
		return fmt.Sprintf("Unknown (%d)", c.toUint32())
	}
}

func (s CircuitState) toUint32() uint32 {
	return uint32(s)
}

// Circuit is a concurrency-safe data structure that can take one of two states at any point in time:
//
// - OpenState - non functional
// - ClosedState - functional
//
// The circuit is intended as a point-in-time indicator of functionality. It has no backoff mechanism, jitter or ramp-up/ramp-down functionality
type Circuit struct {
	bit uint32
}

// NewCircuit creates a new circuit, in the open (non-functional) state
func NewCircuit() *Circuit {
	return &Circuit{bit: OpenState.toUint32()}
}

// State returns the current state of the circuit. Note that concurrent modifications may be happening, so the state may be different than what's returned
func (c *Circuit) State() CircuitState {
	return CircuitState(atomic.LoadUint32(&c.bit))
}

// Close closes the circuit if it wasn't already closed. Returns true if it had to be closed, false if it was already closed
func (c *Circuit) Close() bool {
	return atomic.CompareAndSwapUint32(&c.bit, OpenState.toUint32(), ClosedState.toUint32())
}

// Open opens the circuit if it wasn't already closed. Returns true if it had to be opened, false if it was already open
func (c *Circuit) Open() bool {
	return atomic.CompareAndSwapUint32(&c.bit, ClosedState.toUint32(), OpenState.toUint32())
}
