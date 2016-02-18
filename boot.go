package main

import (
	"log"
	"os"
	"runtime"

	cookoolog "github.com/Masterminds/cookoo/log"
	"github.com/codegangsta/cli"
	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/cleaner"
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
	gitHomeDir            = "/home/git"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	if os.Getenv("DEBUG") == "true" {
		pkglog.DefaultLogger.SetDebug(true)
		cookoolog.Level = cookoolog.LogDebug
		log.Printf("Running in debug mode")
	}

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
				circ := sshd.NewCircuit()

				s3Client, err := storage.GetClient(cnf.HealthSrvTestStorageRegion)
				if err != nil {
					log.Printf("Error getting s3 client (%s)", err)
					os.Exit(1)
				}
				kubeClient, err := kcl.NewInCluster()
				if err != nil {
					log.Printf("Error getting kubernetes client [%s]", err)
					os.Exit(1)
				}
				log.Printf("Starting health check server on port %d", cnf.HealthSrvPort)
				healthSrvCh := make(chan error)
				go func() {
					if err := healthsrv.Start(cnf.HealthSrvPort, kubeClient.Namespaces(), s3Client, circ); err != nil {
						healthSrvCh <- err
					}
				}()
				log.Printf("Starting deleted app cleaner")
				cleanerErrCh := make(chan error)
				go func() {
					if err := cleaner.Run(gitHomeDir, kubeClient.Namespaces(), cnf.CleanerPollSleepDuration); err != nil {
						cleanerErrCh <- err
					}
				}()

				log.Printf("Starting SSH server on %s:%d", cnf.SSHHostIP, cnf.SSHHostPort)
				sshCh := make(chan int)
				go func() {
					sshCh <- pkg.RunBuilder(cnf.SSHHostIP, cnf.SSHHostPort, gitHomeDir, circ)
				}()

				select {
				case err := <-healthSrvCh:
					log.Printf("Error running health server (%s)", err)
					os.Exit(1)
				case i := <-sshCh:
					log.Printf("Unexpected SSH server stop with code %d", i)
					os.Exit(i)
				case err := <-cleanerErrCh:
					log.Printf("Error running the deleted app cleaner (%s)", err)
					os.Exit(1)
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
					log.Printf("Error getting config for %s [%s]", gitReceiveConfAppName, err)
					os.Exit(1)
				}
				cnf.CheckDurations()

				if err := gitreceive.Run(cnf); err != nil {
					log.Printf("Error running git receive hook [%s]", err)
					os.Exit(1)
				}
			},
		},
	}

	app.Run(os.Args)
}
