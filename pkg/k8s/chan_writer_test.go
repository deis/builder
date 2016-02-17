package k8s

import (
	"testing"
	"time"
)

const (
	chanReadTimeout = 100 * time.Millisecond
)

func TestWrite(t *testing.T) {
	strLookup := map[string]bool{"abc": false, "def": false}
	strs := make([]string, len(strLookup))
	i := 0
	for str, _ := range strLookup {
		strs[i] = str
		i++
	}

	ch := make(chan []byte)
	wr := chanWriter{ch: ch}

	go func(strs []string) {
		defer close(ch)
		for _, str := range strs {
			if _, err := wr.Write([]byte(str)); err != nil {
				t.Fatalf("error writing to chan writer (%s)", err)
			}
		}
	}(strs)

	for byteStr := range ch {
		strLookup[string(byteStr)] = true
	}

	for str, ok := range strLookup {
		if !ok {
			t.Fatalf("didn't read %s after it was supposed to be written", str)
		}
	}
}
