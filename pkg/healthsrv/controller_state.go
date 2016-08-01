package healthsrv

import (
	"fmt"
	"net/http"

	"github.com/deis/builder/pkg/controller"
)

// GetClient is an (*net/http).Client compatible interface that provides just the Get cross-section of functionality.
// It can also be implemented for unit tests.
type GetClient interface {
	Get(string) (*http.Response, error)
}

func controllerState(client GetClient, succCh chan<- string, errCh chan<- error, stopCh <-chan struct{}) {
	url, err := controller.URLStr("healthz")
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
