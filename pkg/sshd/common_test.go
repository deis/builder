package sshd

import (
	"fmt"
	"sync"
	"time"
)

type errWGTimedOut struct {
	to time.Duration
}

func (e errWGTimedOut) Error() string {
	return fmt.Sprintf("WaitGroup wait timed out after %s", e.to)
}

// waitWithTimeout waits for wg.Done() until timeout expires. returns errWGTimedOut if timeout expired before wg.Done() returned, otherwise returns nil. this func is naturally leaky if wg.Done() never returns
func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		wg.Wait()
	}()
	select {
	case <-time.After(timeout):
		return errWGTimedOut{to: timeout}
	case <-ch:
		return nil
	}
}

// sshSessionOutput is the output from a *(golang.org/x/crypto/ssh).Session's Output() call
type sshSessionOutput struct {
	outStr string
	err    error
}
