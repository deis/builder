package main

import (
	"os"
	"runtime"

	cookoolog "github.com/Masterminds/cookoo/log"
	"github.com/codegangsta/cli"
	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/conf"
	"github.com/deis/builder/pkg/gitreceive"
	"github.com/deis/builder/pkg/gitreceive/storage"
	"github.com/deis/builder/pkg/healthsrv"
	"github.com/deis/builder/pkg/sshd"
	pkglog "github.com/deis/pkg/log"
	kcl "k8s.io/kubernetes/pkg/client/unversioned"
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
		pkglog.DefaultLogger.SetDebug(true)
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
				s3Client, err := storage.GetClient(cnf.HealthSrvTestStorageRegion)
				if err != nil {
					pkglog.Err("getting s3 client [%s]", err)
					os.Exit(1)
				}
				kubeClient, err := kcl.NewInCluster()
				if err != nil {
					pkglog.Err("getting kubernetes client [%s]", err)
					os.Exit(1)
				}
				pkglog.Info("starting health check server on port %d", cnf.HealthSrvPort)
				healthSrvCh := make(chan error)
				go func() {
					if err := healthsrv.Start(cnf.HealthSrvPort, kubeClient.Namespaces(), s3Client); err != nil {
						healthSrvCh <- err
					}
				}()

				pkglog.Info("starting SSH server on %s:%d", cnf.SSHHostIP, cnf.SSHHostPort)
				sshCh := make(chan int)
				go func() {
					sshCh <- pkg.Run(cnf.SSHHostIP, cnf.SSHHostPort, "boot")
				}()

				select {
				case err := <-healthSrvCh:
					pkglog.Err("Error running health server (%s)", err)
					os.Exit(1)
				case i := <-sshCh:
					pkglog.Err("Unexpected SSH server stop with code %d", i)
					os.Exit(i)
				}
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
