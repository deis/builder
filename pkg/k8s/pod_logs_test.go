package k8s

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/arschles/assert"
)

type fakeLogStreamer struct {
	str string
}

func (f fakeLogStreamer) Stream() (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader(f.str)), nil
}

func TestStreamPodLogs(t *testing.T) {
	streamer := fakeLogStreamer{str: "abcdef"}
	succCh := make(chan struct{})
	logsCh := make(chan []byte)
	errCh := make(chan error)
	go streamPodLogs(streamer, succCh, logsCh, errCh)
	select {
	case s := <-logsCh:
		assert.Equal(t, streamer.str, string(s), "received string")
	case err := <-errCh:
		t.Fatalf("receiving a log from streamPodLogs (%s)", err)
	}
	select {
	case <-succCh:
	case <-time.After(chanReadTimeout):
		t.Fatalf("receiving success from streamPodLogs after %s", chanReadTimeout)
	}
}
