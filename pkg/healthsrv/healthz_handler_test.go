package healthsrv

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/sshd"
)

var (
	testErr = errors.New("test error")
)

func TestHealthZCircuitOpen(t *testing.T) {
	nsLister := emptyNamespaceLister{}
	bLister := emptyBucketLister{}
	c := sshd.NewCircuit()

	h := healthZHandler(nsLister, bLister, c)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
	expectedBody := "SSH Server is not yet started"
	assert.Equal(t, strings.TrimSpace(string(w.Body.Bytes())), expectedBody, "response body")
}

func TestHealthZBucketListErr(t *testing.T) {
	nsLister := emptyNamespaceLister{}
	bLister := errBucketLister{err: testErr}
	c := sshd.NewCircuit()
	c.Close()
	h := healthZHandler(nsLister, bLister, c)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
	expectedBody := fmt.Sprintf("Error listing buckets (%s)", testErr)
	assert.Equal(t, strings.TrimSpace(string(w.Body.Bytes())), expectedBody, "response body")
}

func TestHealthZNamespaceListErr(t *testing.T) {
	nsLister := errNamespaceLister{err: testErr}
	bLister := emptyBucketLister{}
	c := sshd.NewCircuit()
	c.Close()

	h := healthZHandler(nsLister, bLister, c)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
	expectedBody := fmt.Sprintf("Error listing namespaces (%s)", testErr)
	assert.Equal(t, strings.TrimSpace(string(w.Body.Bytes())), expectedBody, "response body")
}

func TestHealthZSuccess(t *testing.T) {
	t.Skip("TODO")
}
