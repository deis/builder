package sshd

import (
	"errors"
	"sync"
	"time"
)

var (
	errWGTimedOut = errors.New("waitgroup wait timed out")
)

// waitWithTimeout waits for wg.Done() until timeout expires. returns errWGTimedOut if timeout expired before wg.Done() returned, otherwise returns nil. this func is naturally leaky if wg.Done() never returns
func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		wg.Wait()
	}()
	select {
	case <-time.After(timeout):
		return errWGTimedOut
	case <-ch:
		return nil
	}
}
