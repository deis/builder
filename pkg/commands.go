package pkg

import (
	"os"
	"os/signal"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/log"
	"github.com/Masterminds/cookoo/safely"
)

// KillOnExit kills PIDs when the program exits.
//
// Otherwise, this blocks until an os.Interrupt or os.Kill is received.
//
// Params:
//  This treats Params as a map of process names (unimportant) to PIDs. It then
// attempts to kill all of the pids that it receives.
func KillOnExit(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	safely.GoDo(c, func() {
		log.Info(c, "Builder is running.")

		<-sigs

		c.Log("info", "Builder received signal to stop.")
		pids := p.AsMap()
		killed := 0
		for name, pid := range pids {
			if pid, ok := pid.(int); ok {
				if proc, err := os.FindProcess(pid); err == nil {
					log.Infof(c, "Killing %s (pid=%d)", name, pid)
					proc.Kill()
					killed++
				}
			}
		}
	})
	return nil, nil
}
