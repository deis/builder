package healthsrv

import (
	"log"
	"net/http"
	"time"

	"github.com/deis/builder/pkg/sshd"
)

const (
	waitTimeout = 10 * time.Second
)

func healthZHandler(bLister BucketLister, serverCircuit *sshd.Circuit) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stopCh := make(chan struct{})

		numChecks := 0
		serverStateCh := make(chan struct{})
		serverStateErrCh := make(chan error)
		go circuitState(serverCircuit, serverStateCh, serverStateErrCh, stopCh)
		numChecks++

		listBucketsCh := make(chan []string)
		listBucketsErrCh := make(chan error)
		go listBuckets(bLister, listBucketsCh, listBucketsErrCh, stopCh)
		numChecks++

		timeoutCh := time.After(waitTimeout)
		defer close(stopCh)
		for i := 0; i < numChecks; i++ {
			select {
			// ensuring the SSH server has been started
			case <-serverStateCh:
			case err := <-serverStateErrCh:
				log.Printf("Healthcheck error getting server state (%s)", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				return

			// listing buckets
			case <-listBucketsCh:
			case err := <-listBucketsErrCh:
				log.Printf("Healthcheck error listing buckets (%s)", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				return

			// timeout for everything all of the above
			case <-timeoutCh:
				log.Printf("Healthcheck endpoint timed out after %s", waitTimeout)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}
