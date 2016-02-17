package k8s

import (
	"sync"
	"testing"
	"time"
)

const (
	chanReadTimeout = 100 * time.Millisecond
)

func TestWrite(t *testing.T) {
	strs := map[string]bool{"abc": false, "def": false}
	ch := make(chan []byte)
	wr := chanWriter{ch: ch}
	var wg sync.WaitGroup
	for str, _ := range strs {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			wr.Write([]byte(s))
		}(str)
	}
	for str, _ := range strs {
		select {
		case s := <-ch:
			if str == string(s) {
				strs[str] = true
			}
		case <-time.After(chanReadTimeout):
			t.Errorf("couldn't read string after %s", chanReadTimeout)
		}
	}
	wg.Wait()
	for str, read := range strs {
		if !read {
			t.Errorf("didn't read %s after written", str)
		}
	}
}
