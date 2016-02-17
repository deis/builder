package k8s

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/arschles/assert"
)

type fakeLogStreamer struct {
	str string
}

func (f fakeLogStreamer) Stream() (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader(f.str)), nil
}

func TestWritePodLogs(t *testing.T) {
	buf := new(bytes.Buffer)
	streamer := fakeLogStreamer{str: "abcdef"}
	num, err := WritePodLogs(streamer, buf)
	assert.NoErr(t, err)
	assert.Equal(t, num, int64(len(streamer.str)), "length of bytes written")
	assert.Equal(t, string(buf.Bytes()), streamer.str, "logs written")
}
