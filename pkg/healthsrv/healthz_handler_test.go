package healthsrv

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
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
	assert.Equal(t, w.Body.Len(), 0, "response body length")
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
	assert.Equal(t, w.Body.Len(), 0, "response body length")
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
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}

func TestHealthZSuccess(t *testing.T) {
	nsLister := emptyNamespaceLister{}
	bLister := emptyBucketLister{}
	c := sshd.NewCircuit()
	c.Close()

	h := healthZHandler(nsLister, bLister, c)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}
