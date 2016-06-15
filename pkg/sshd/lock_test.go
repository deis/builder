package sshd

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/arschles/assert"
)

const (
	callbackTimeout = 1 * time.Second
)

var (
	gitError = errors.New("git receive error")
)

func TestMultipleSameRepoLocks(t *testing.T) {
	var wg sync.WaitGroup
	const repo = "repo1"
	const numTries = 0
	lck := NewInMemoryRepositoryLock(0)
	assert.NoErr(t, lck.Lock(repo))
	for i := 0; i < numTries; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.True(t, lck.Lock(repo) != nil, "lock of already locked repo should return error")
		}()
	}
	assert.NoErr(t, waitWithTimeout(&wg, 1*time.Second))
	assert.NoErr(t, lck.Unlock(repo))
	for i := 0; i < numTries; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.True(t, lck.Unlock(repo) != nil, "unlock of already unlocked repo should return error")
		}()
	}
	assert.NoErr(t, waitWithTimeout(&wg, 1*time.Second))
}

func TestSingleLock(t *testing.T) {
	rl := NewInMemoryRepositoryLock(0)
	key := "fakeid"
	callbackCh := make(chan interface{})
	go lockAndCallback(rl, key, callbackCh)
	verifyCallbackHappens(t, callbackCh)
}

func TestSingleLockUnlock(t *testing.T) {
	rl := NewInMemoryRepositoryLock(0)
	key := "fakeid"
	callbackCh := make(chan interface{})
	go lockAndCallback(rl, key, callbackCh)
	verifyCallbackHappens(t, callbackCh)
	err := rl.Unlock(key)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestInvalidUnlock(t *testing.T) {
	rl := NewInMemoryRepositoryLock(0)
	key := "fakeid"
	err := rl.Unlock(key)
	if err == nil {
		t.Fatal("expected error but returned nil")
	}
}

func TestDoubleLockUnlock(t *testing.T) {
	rl := NewInMemoryRepositoryLock(0)
	key := "fakeid"
	callbackCh1stLock := make(chan interface{})
	callbackCh2ndLock := make(chan interface{})

	go lockAndCallback(rl, key, callbackCh1stLock)
	verifyCallbackHappens(t, callbackCh1stLock)
	go lockAndCallback(rl, key, callbackCh2ndLock)
	verifyCallbackDoesntHappens(t, callbackCh2ndLock)
	err := rl.Unlock(key)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	err = rl.Unlock(key)
	if err == nil {
		t.Fatalf("expected error but returned nil")
	}
}

func TestWrapInLock(t *testing.T) {
	lck := NewInMemoryRepositoryLock(0)
	assert.NoErr(t, wrapInLock(lck, "repo", func() error {
		return nil
	}))
	assert.Err(t, gitError, wrapInLock(lck, "repo", func() error {
		return gitError
	}))
	assert.NoErr(t, lck.Lock("repo"))
	assert.Err(t, errAlreadyLocked, wrapInLock(lck, "repo", func() error {
		return nil
	}))
}

func lockAndCallback(rl RepositoryLock, id string, callbackCh chan<- interface{}) {
	if err := rl.Lock(id); err == nil {
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
