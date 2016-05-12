package healthsrv

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/deis/builder/pkg/controller"
)

// GetClient is an (*net/http).Client compatible interface that provides just the Get cross-section of functionality.
// It can also be implemented for unit tests.
type GetClient interface {
	Get(string) (*http.Response, error)
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

func controllerState(client GetClient, succCh chan<- string, errCh chan<- error, stopCh <-chan struct{}) {
	url, err := controller.ControllerURLStr("healthz")
	if err != nil {
		select {
		case errCh <- err:
		case <-stopCh:
		}
		return
	}
	resp, err := client.Get(url)
	if err != nil {
		select {
		case errCh <- err:
		case <-stopCh:
		}
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		select {
		case errCh <- fmt.Errorf("Failed to get controller health status"):
		case <-stopCh:
		}
		return
	}
	select {
	case succCh <- "controller is healthy":
	case <-stopCh:
		return
	}
}
