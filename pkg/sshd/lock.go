package sshd

import (
	"fmt"
	"sync"
	"time"
)

// RepositoryLock interface that allows the creation of a lock associated
// with a repository name to avoid simultaneous git operations
type RepositoryLock interface {
	// Lock acquires a lock for a repository. In the case the repository is already locked
	// it waits until a timeout to get the lock. If it was not possible to get the
	// lock after the timeout an error is returned
	Lock(repoName string, timeout time.Duration) error
	// Unlock releases the lock for a repository or returns an error if the specified
	// name doesn't exist. In the case the repository is already locked it waits until
	// a timeout to get the lock. If it was not possible to get the lock after the timeout
	// an error is returned
	Unlock(repoName string, timeout time.Duration) error
}

// NewInMemoryRepositoryLock returns a new instance of a RepositoryLock
func NewInMemoryRepositoryLock() RepositoryLock {
	return &inMemoryRepoLock{
		mutex:   &sync.RWMutex{},
		dataMap: make(map[string]bool),
	}
}

type inMemoryRepoLock struct {
	mutex   *sync.RWMutex
	dataMap map[string]bool
}

// Lock aquires a lock associated with the specified name.
// This implementation do not uses the timeout
func (rl *inMemoryRepoLock) Lock(repoName string, timeout time.Duration) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	_, exists := rl.dataMap[repoName]
	if !exists {
		rl.dataMap[repoName] = true
		return nil
	}

	return fmt.Errorf("repository %q already locked", repoName)
}

// Unlock releases the lock for a repository or returns
// an error if the specified name doesn't exist.
// This implementation do not uses the timeout
func (rl *inMemoryRepoLock) Unlock(repoName string, timeout time.Duration) error {
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
