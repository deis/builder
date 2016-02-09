package sshd

import (
	"testing"
)

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

}
