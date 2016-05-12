package healthsrv

import (
	"log"
	"net/http"
	"time"

	"k8s.io/kubernetes/pkg/api"
)

func readinessHandler(client GetClient, nsLister NamespaceLister) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stopCh := make(chan struct{})

		numChecks := 0
		namespaceListerCh := make(chan *api.NamespaceList)
		namespaceListerErrCh := make(chan error)
		go listNamespaces(nsLister, namespaceListerCh, namespaceListerErrCh, stopCh)
		numChecks++

		controllerStateCh := make(chan string)
		controllerStateErrCh := make(chan error)
		go controllerState(client, controllerStateCh, controllerStateErrCh, stopCh)
		numChecks++

		timeoutCh := time.After(waitTimeout)
		defer close(stopCh)
		for i := 0; i < numChecks; i++ {
			select {
			// listing k8s namespaces
			case <-namespaceListerCh:
			case err := <-namespaceListerErrCh:
				log.Printf("Readinesscheck error listing namespaces (%s)", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				return

			// listing k8s namespaces
			case <-controllerStateCh:
			case err := <-controllerStateErrCh:
				log.Printf("Readinesscheck error listning controller (%s)", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				return

			// timeout for everything all of the above
			case <-timeoutCh:
				log.Printf("Readinesscheck endpoint timed out after %s", waitTimeout)
				w.WriteHeader(http.StatusServiceUnavailable)
				return

			}
		}
		w.WriteHeader(http.StatusOK)
	})
}
