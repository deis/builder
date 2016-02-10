package healthsrv

import (
	"fmt"

	"github.com/deis/builder/pkg/sshd"
)

// circuitState determines whether circ.State() == sshd.ClosedState, and sends the results back on the various given channels. This func is intended to be run in a goroutine and communicates via the channels it's passed.
//
// If the circuit is closed, it passes an empty struct back on succCh. On failure, it sends an error back on errCh. At most one of {succCh, errCh} will be sent on. If stopCh is closed, no pending or future sends will occur.
func circuitState(circ *sshd.Circuit, succCh chan<- struct{}, errCh chan<- error, stopCh <-chan struct{}) {
	// There's a race between the boolean eval and the HTTP error returned (the circuit could close between the two). This function should be polled to avoid that problem. If it's being used in a k8s probe, then you're fine because k8s will repeat the health probe and effectively re-evaluate the boolean
	if circ.State() != sshd.ClosedState {
		select {
		case errCh <- fmt.Errorf("SSH Server is not yet started"):
		case <-stopCh:
		}
		return
	}
	select {
	case succCh <- struct{}{}:
	case <-stopCh:
	}
}
