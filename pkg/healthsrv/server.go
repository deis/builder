package healthsrv

import (
	"fmt"
	"net/http"

	"github.com/deis/builder/pkg/sshd"
)

// Start starts the healthcheck server on :$port and blocks. It only returns if the server fails,
// with the indicative error.
func Start(port int, nsLister NamespaceLister, bLister BucketLister, sshServerCircuit *sshd.Circuit) error {
	mux := http.NewServeMux()
	client := &http.Client{}
	mux.Handle("/healthz", healthZHandler(bLister, sshServerCircuit))
	mux.Handle("/readiness", readinessHandler(client, nsLister))

	hostStr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(hostStr, mux)
}
