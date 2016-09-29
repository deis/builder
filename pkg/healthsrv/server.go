package healthsrv

import (
	"fmt"
	"net/http"

	"github.com/deis/builder/pkg/controller"
	"github.com/deis/builder/pkg/sshd"
)

// Start starts the healthcheck server on :$port and blocks. It only returns if the server fails,
// with the indicative error.
func Start(cnf *sshd.Config, nsLister NamespaceLister, bLister BucketLister, sshServerCircuit *sshd.Circuit) error {
	mux := http.NewServeMux()
	client, err := controller.New(cnf.ControllerHost, cnf.ControllerPort)
	if err != nil {
		return err
	}
	mux.Handle("/healthz", healthZHandler(bLister, sshServerCircuit))
	mux.Handle("/readiness", readinessHandler(client, nsLister))

	hostStr := fmt.Sprintf(":%d", cnf.HealthSrvPort)
	return http.ListenAndServe(hostStr, mux)
}
