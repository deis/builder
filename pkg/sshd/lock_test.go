package sshd

import (
	"sync"
	"testing"
	"time"

	"github.com/arschles/assert"
)

const (
	callbackTimeout = 1 * time.Second
)

func TestMultipleSameRepoLocks(t *testing.T) {
	var wg sync.WaitGroup
	const repo = "repo1"
	const numTries = 100
	lck := NewInMemoryRepositoryLock()
	assert.NoErr(t, lck.Lock(repo, 0*time.Second))
	for i := 0; i < numTries; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.True(t, lck.Lock(repo, 0*time.Second) != nil, "lock of already locked repo should return error")
		}()
	}
	assert.NoErr(t, waitWithTimeout(&wg, 1*time.Second))
	assert.NoErr(t, lck.Unlock(repo, 0*time.Second))
	for i := 0; i < numTries; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.True(t, lck.Unlock(repo, 0*time.Second) != nil, "unlock of already unlocked repo should return error")
		}()
	}
	assert.NoErr(t, waitWithTimeout(&wg, 1*time.Second))
}

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
		t.Fatal("expected error but returned nil")
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

func TestWrapInLock(t *testing.T) {
	lck := NewInMemoryRepositoryLock()
	assert.NoErr(t, wrapInLock(lck, "repo", 0*time.Second, func() error {
		return nil
	}))
	lck.Lock("repo", 0*time.Second)
	assert.Err(t, errAlreadyLocked, wrapInLock(lck, "repo", 0*time.Second, func() error {
		return nil
	}))
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
