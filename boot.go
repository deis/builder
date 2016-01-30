package main

import (
	"os"
	"runtime"

	cookoolog "github.com/Masterminds/cookoo/log"
	"github.com/codegangsta/cli"
	"github.com/deis/builder/fetcher"
	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/conf"
	"github.com/deis/builder/pkg/gitreceive"
	pkglog "github.com/deis/builder/pkg/gitreceive/log"
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
	if os.Getenv("DEBUG") == "true" {
		pkglog.IsDebugging = true
		cookoolog.Level = cookoolog.LogDebug
	}
	pkglog.Debug("Running in debug mode")

	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "server",
			Aliases: []string{"srv"},
			Usage:   "Run the git server",
			Action: func(c *cli.Context) {
				cnf := new(sshd.Config)
				if err := conf.EnvConfig(serverConfAppName, cnf); err != nil {
					pkglog.Err("getting config for %s [%s]", serverConfAppName, err)
					os.Exit(1)
				}
				pkglog.Info("starting fetcher on port %d", cnf.FetcherPort)
				go fetcher.Serve(cnf.FetcherPort)
				pkglog.Info("starting SSH server on %s:%d", cnf.SSHHostIP, cnf.SSHHostPort)
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
					pkglog.Err("Error getting config for %s [%s]", gitReceiveConfAppName, err)
					os.Exit(1)
				}
				cnf.CheckDurations()

				if err := gitreceive.Run(cnf); err != nil {
					pkglog.Err("running git receive hook [%s]", err)
					os.Exit(1)
				}
			},
		},
	}

	app.Run(os.Args)
}
