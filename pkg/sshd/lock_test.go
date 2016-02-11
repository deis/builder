package sshd

import (
	"testing"
	"time"
)

const (
	callbackTimeout = 1 * time.Second
)

func TestSingleLock(t *testing.T) {
	rl := NewInMemoryRepositoryLock()
	key := "fakeid"
	callbackCh := make(chan interface{})
	go lockAndCallback(rl, key, callbackCh)
	verifyCallbackHappens(t, callbackCh)
}

func TestSingleLockUnlock(t *testing.T) {
	rl := NewInMemoryRepositoryLock()
	key := "fakeid"
	callbackCh := make(chan interface{})
	go lockAndCallback(rl, key, callbackCh)
	verifyCallbackHappens(t, callbackCh)
	err := rl.Unlock(key, time.Duration(0))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestInvalidUnlock(t *testing.T) {
	rl := NewInMemoryRepositoryLock()
	key := "fakeid"
	err := rl.Unlock(key, time.Duration(0))
	if err == nil {
		t.Fatalf("expected error but returned nil", err)
	}
}

func TestDoubleLockUnlock(t *testing.T) {
	rl := NewInMemoryRepositoryLock()
	key := "fakeid"
	callbackCh1stLock := make(chan interface{})
	callbackCh2ndLock := make(chan interface{})

	go lockAndCallback(rl, key, callbackCh1stLock)
	verifyCallbackHappens(t, callbackCh1stLock)
	go lockAndCallback(rl, key, callbackCh2ndLock)
	verifyCallbackDoesntHappens(t, callbackCh2ndLock)
	err := rl.Unlock(key, time.Duration(0))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	err = rl.Unlock(key, time.Duration(0))
	if err == nil {
		t.Fatalf("expected error but returned nil")
	}
}

func lockAndCallback(rl RepositoryLock, id string, callbackCh chan<- interface{}) {
	if err := rl.Lock(id, time.Duration(0)); err == nil {
		callbackCh <- true
	}
}

func verifyCallbackHappens(t *testing.T, callbackCh <-chan interface{}) bool {
	select {
	case <-callbackCh:
		return true
	case <-time.After(callbackTimeout):
		t.Fatalf("Timed out waiting for callback.")
		return false
	}
}

func verifyCallbackDoesntHappens(t *testing.T, callbackCh <-chan interface{}) bool {
	select {
	case <-callbackCh:
		t.Fatalf("Unexpected callback.")
		return false
	case <-time.After(callbackTimeout):
		return true
	}
}
