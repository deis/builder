package healthsrv

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"

	"github.com/docker/distribution/context"

	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/sshd"
	"github.com/deis/controller-sdk-go"
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

type failureGetClient struct{}

func (e failureGetClient) Get(url string) (*http.Response, error) {
	resp := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader("")),
		StatusCode: http.StatusServiceUnavailable,
	}
	return resp, nil
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

type fakeHTTPServer struct {
	// determines wether to return success or failure.
	Healthy bool
}

func (f fakeHTTPServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("DEIS_API_VERSION", deis.APIVersion)

	if req.URL.Path == "/healthz" {
		if f.Healthy {
			res.WriteHeader(http.StatusOK)
		} else {
			res.WriteHeader(http.StatusServiceUnavailable)
		}
		res.Write(nil)
		return
	}

	fmt.Printf("Unrecongized URL %s\n", req.URL)
	res.WriteHeader(http.StatusNotFound)
	res.Write(nil)
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
	handler := fakeHTTPServer{true}
	server := httptest.NewServer(handler)
	defer server.Close()
	client, err := deis.New(false, server.URL, "")
	if err != nil {
		t.Fatal(err)
	}

	nsLister := errNamespaceLister{err: errTest}

	h := readinessHandler(client, nsLister)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/readiness", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}

func TestReadinessControllerErr(t *testing.T) {
	handler := fakeHTTPServer{false}
	server := httptest.NewServer(handler)
	defer server.Close()
	client, err := deis.New(false, server.URL, "")
	if err != nil {
		t.Fatal(err)
	}

	nsLister := emptyNamespaceLister{}

	h := readinessHandler(client, nsLister)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/readiness", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusServiceUnavailable, "response code")
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
	handler := fakeHTTPServer{true}
	server := httptest.NewServer(handler)
	defer server.Close()
	client, err := deis.New(false, server.URL, "")
	if err != nil {
		t.Fatal(err)
	}

	nsLister := emptyNamespaceLister{}

	h := readinessHandler(client, nsLister)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/readiness", bytes.NewBuffer(nil))
	assert.NoErr(t, err)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK, "response code")
	assert.Equal(t, w.Body.Len(), 0, "response body length")
}
