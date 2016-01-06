package main

import (
	"os"
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/deis/builder/fetcher"
	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/conf"
	"github.com/deis/builder/pkg/gitreceive"
	"github.com/deis/builder/pkg/log"
	"github.com/deis/builder/pkg/sshd"
)

const (
	serverConfAppName     = "deis-builder-server"
	gitReceiveConfAppName = "deis-builder-git-receive"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "server",
			Aliases: []string{"srv"},
			Usage:   "Run the git server",
			Action: func(c *cli.Context) {
				cnf := new(sshd.Config)
				if err := conf.EnvConfig(serverConfAppName, cnf); err != nil {
					log.Err("getting config for %s [%s]", serverConfAppName, err)
					os.Exit(1)
				}
				log.Info("starting fetcher on port %d", cnf.FetcherPort)
				go fetcher.Serve(cnf.FetcherPort)
				log.Info("starting SSH server on %s:%d", cnf.SSHHostIP, cnf.SSHHostPort)
				os.Exit(pkg.Run(cnf.SSHHostIP, cnf.SSHHostPort, "boot"))
			},
		},
		{
			Name:    "git-receive",
			Aliases: []string{"gr"},
			Usage:   "Run the git-receive hook",
			Action: func(c *cli.Context) {
				cnf := new(gitreceive.Config)
				if err := conf.EnvConfig(gitReceiveConfAppName, cnf); err != nil {
					log.Err("Error getting config for %s [%s]", gitReceiveConfAppName, err)
					os.Exit(1)
				}
				if err := gitreceive.Run(cnf); err != nil {
					log.Err("running git receive hook [%s]", err)
					os.Exit(1)
				}
			},
		},
	}

	app.Run(os.Args)
}
