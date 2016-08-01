package healthsrv

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"

	"github.com/docker/distribution/context"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/sshd"
)

var (
	errTest = errors.New("test error")
)

type emptyBucketLister struct{}

func (e emptyBucketLister) List(ctx context.Context, opath string) ([]string, error) {
	return nil, nil
}

type errBucketLister struct {
	err error
}

func (e errBucketLister) List(ctx context.Context, opath string) ([]string, error) {
	return nil, e.err
}

type successGetClient struct{}

func (e successGetClient) Get(url string) (*http.Response, error) {
	resp := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader("")),
		StatusCode: http.StatusOK,
	}
	return resp, nil
}

type failureGetClient struct{}

func (e failureGetClient) Get(url string) (*http.Response, error) {
	resp := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader("")),
		StatusCode: http.StatusServiceUnavailable,
	}
	return resp, nil
}

type errGetClient struct {
	err error
}

func (e errGetClient) Get(url string) (*http.Response, error) {
	return nil, e.err
}

type emptyNamespaceLister struct{}

func (n emptyNamespaceLister) List(opts api.ListOptions) (*api.NamespaceList, error) {
	return &api.NamespaceList{}, nil
}

type errNamespaceLister struct {
	err error
}

func (e errNamespaceLister) List(opts api.ListOptions) (*api.NamespaceList, error) {
	return nil, e.err
}

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
	client := successGetClient{}
	os.Setenv("DEIS_CONTROLLER_SERVICE_HOST", "127.0.0.1")
	os.Setenv("DEIS_CONTROLLER_SERVICE_PORT", "8000")

	h := readinessHandler(client, nsLister)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/readiness", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}

func TestReadinessControllerErr(t *testing.T) {
	nsLister := emptyNamespaceLister{}
	client := failureGetClient{}

	h := readinessHandler(client, nsLister)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/readiness", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}

func TestReadinessControllerGetErr(t *testing.T) {
	nsLister := emptyNamespaceLister{}
	client := errGetClient{err: errTest}

	h := readinessHandler(client, nsLister)
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
	client := successGetClient{}

	h := readinessHandler(client, nsLister)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/readiness", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}
