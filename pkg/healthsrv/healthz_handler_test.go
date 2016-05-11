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
	errTest = errors.New("test error")
)

func TestHealthZCircuitOpen(t *testing.T) {
	bLister := emptyBucketLister{}
	c := sshd.NewCircuit()

	h := healthZHandler(bLister, c)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}

func TestHealthZBucketListErr(t *testing.T) {
	bLister := errBucketLister{err: errTest}
	c := sshd.NewCircuit()
	c.Close()

	h := healthZHandler(bLister, c)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}

func TestReadinessNamespaceListErr(t *testing.T) {
	nsLister := errNamespaceLister{err: errTest}

	h := readinessHandler(nsLister)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/readiness", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}

func TestHealthZSuccess(t *testing.T) {
	bLister := emptyBucketLister{}
	c := sshd.NewCircuit()
	c.Close()

	h := healthZHandler(bLister, c)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}

func TestReadinessSuccess(t *testing.T) {
	nsLister := emptyNamespaceLister{}

	h := readinessHandler(nsLister)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/readiness", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}
