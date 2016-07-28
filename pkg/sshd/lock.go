package sshd

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	errAlreadyLocked = errors.New("already locked")
)

// RepositoryLock interface that allows the creation of a lock associated
// with a repository name to avoid simultaneous git operations.
type RepositoryLock interface {
	// Lock acquires a lock for a repository.
	Lock(repoName string) error
	// Unlock releases the lock for a repository or returns an error if the specified
	// name doesn't exist.
	Unlock(repoName string) error
	// Timeout returns the time duration for which it has to hold the lock
	Timeout() time.Duration
}

func wrapInLock(lck RepositoryLock, repoName string, fn func() error) error {
	if err := lck.Lock(repoName); err != nil {
		return errAlreadyLocked
	}
	timer := time.NewTimer(lck.Timeout())
	defer timer.Stop()
	doneCh := make(chan struct{})
	fnCh := make(chan error)
	go func() {
		err := fn()
		select {
		case fnCh <- err:
		case <-doneCh:
		}
	}()
	defer lck.Unlock(repoName)
	select {
	case <-timer.C:
		defer close(doneCh)
		return fmt.Errorf("%s lock exceeded timout", repoName)
	case err := <-fnCh:
		return err
	}
}

// NewInMemoryRepositoryLock returns a new instance of a RepositoryLock.
func NewInMemoryRepositoryLock(timeout time.Duration) RepositoryLock {
	return &inMemoryRepoLock{
		mutex:   &sync.RWMutex{},
		dataMap: make(map[string]bool),
		timeout: timeout,
	}
}

type inMemoryRepoLock struct {
	mutex   *sync.RWMutex
	dataMap map[string]bool
	timeout time.Duration
}

// Lock acquires a lock associated with the specified name.
func (rl *inMemoryRepoLock) Lock(repoName string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	_, exists := rl.dataMap[repoName]
	if !exists {
		rl.dataMap[repoName] = true
		return nil
	}

	return fmt.Errorf("repository %q already locked", repoName)
}

// Unlock releases the lock for a repository or returns an error if the specified name doesn't
// exist.
func (rl *inMemoryRepoLock) Unlock(repoName string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	locked, exists := rl.dataMap[repoName]
	if !exists {
		return fmt.Errorf("repository %q not found", repoName)
	}

	if locked {
		delete(rl.dataMap, repoName)
	}

	return nil
}

// Timeout returns the time duration for which a gitpush should hold the lock
func (rl *inMemoryRepoLock) Timeout() time.Duration {
	return rl.timeout
}
