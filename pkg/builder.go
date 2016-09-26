// Package pkg provides common libraries for the Deis builder.
//
// The Deis builder is responsible for building slugs and docker images for use in the Deis
// on the Deis PaaS platform.
package pkg

import (
	"fmt"

	"github.com/deis/builder/pkg/sshd"
	"github.com/deis/pkg/log"
)

// Return codes that will be sent to the shell.
const (
	StatusOk = iota
	StatusLocalError
)

// RunBuilder starts the Builder service.
//
// The Builder service is responsible for setting up the local container
// environment and then listening for new builds. The main listening service
// is SSH. Builder listens for new Git commands and then sends those on to
// Git.
//
// Run returns on of the Status* status code constants.
func RunBuilder(cnf *sshd.Config, gitHomeDir string, sshServerCircuit *sshd.Circuit, pushLock sshd.RepositoryLock) int {
	address := fmt.Sprintf("%s:%d", cnf.SSHHostIP, cnf.SSHHostPort)
	cfg, err := sshd.Configure(cnf)
	if err != nil {
		log.Err("SSH server configuration failed: %s", err)
		return StatusLocalError
	}
	receivetype := "gitreceive"
	if err := sshd.Serve(cfg, sshServerCircuit, gitHomeDir, pushLock, address, receivetype); err != nil {
		log.Err("SSH server failed: %s", err)
		return StatusLocalError
	}

	return StatusOk
}
