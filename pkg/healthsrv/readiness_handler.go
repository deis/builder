package healthsrv

import (
	"log"
	"net/http"
	"time"

	"k8s.io/kubernetes/pkg/api"
)

func readinessHandler(nsLister NamespaceLister) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stopCh := make(chan struct{})

		namespaceListerCh := make(chan *api.NamespaceList)
		namespaceListerErrCh := make(chan error)
		go listNamespaces(nsLister, namespaceListerCh, namespaceListerErrCh, stopCh)

		timeoutCh := time.After(waitTimeout)
		defer close(stopCh)
		select {
		// listing k8s namespaces
		case <-namespaceListerCh:
		case err := <-namespaceListerErrCh:
			log.Printf("Healthcheck error listing namespaces (%s)", err)
			w.WriteHeader(http.StatusServiceUnavailable)
		// timeout for everything all of the above
		case <-timeoutCh:
			log.Printf("Healthcheck endpoint timed out after %s", waitTimeout)
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		w.WriteHeader(http.StatusOK)
	})
}
