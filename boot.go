package main

import (
	"log"
	"os"
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/deis/builder/pkg"
	"github.com/deis/builder/pkg/cleaner"
	"github.com/deis/builder/pkg/conf"
	"github.com/deis/builder/pkg/gitreceive"
	"github.com/deis/builder/pkg/healthsrv"
	"github.com/deis/builder/pkg/sshd"
	"github.com/deis/builder/pkg/sys"
	pkglog "github.com/deis/pkg/log"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	_ "github.com/docker/distribution/registry/storage/driver/azure"
	"github.com/docker/distribution/registry/storage/driver/factory"
	_ "github.com/docker/distribution/registry/storage/driver/gcs"
	_ "github.com/docker/distribution/registry/storage/driver/s3-aws"
	_ "github.com/docker/distribution/registry/storage/driver/swift"
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
				fs := sys.RealFS()
				env := sys.RealEnv()
				pushLock := sshd.NewInMemoryRepositoryLock(cnf.GitLockTimeout())
				circ := sshd.NewCircuit()

				storageParams, err := conf.GetStorageParams(env)
				if err != nil {
					log.Printf("Error getting storage parameters (%s)", err)
					os.Exit(1)
				}
				var storageDriver storagedriver.StorageDriver
				if cnf.StorageType == "minio" {
					storageDriver, err = factory.Create("s3", storageParams)
				} else {
					storageDriver, err = factory.Create(cnf.StorageType, storageParams)
				}
				if err != nil {
					log.Printf("Error creating storage driver (%s)", err)
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
					if err := healthsrv.Start(cnf.HealthSrvPort, kubeClient.Namespaces(), storageDriver, circ); err != nil {
						healthSrvCh <- err
					}
				}()
				log.Printf("Starting deleted app cleaner")
				cleanerErrCh := make(chan error)
				go func() {
					if err := cleaner.Run(gitHomeDir, kubeClient.Namespaces(), fs, cnf.CleanerPollSleepDuration()); err != nil {
						cleanerErrCh <- err
					}
				}()

				log.Printf("Starting SSH server on %s:%d", cnf.SSHHostIP, cnf.SSHHostPort)
				sshCh := make(chan int)
				go func() {
					sshCh <- pkg.RunBuilder(cnf.SSHHostIP, cnf.SSHHostPort, gitHomeDir, circ, pushLock)
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
				fs := sys.RealFS()
				env := sys.RealEnv()
				storageParams, err := conf.GetStorageParams(env)
				if err != nil {
					log.Printf("Error getting storage parameters (%s)", err)
					os.Exit(1)
				}
				var storageDriver storagedriver.StorageDriver
				if cnf.StorageType == "minio" {
					storageDriver, err = factory.Create("s3", storageParams)
				} else {
					storageDriver, err = factory.Create(cnf.StorageType, storageParams)
				}
				if err != nil {
					log.Printf("Error creating storage driver (%s)", err)
					os.Exit(1)
				}

				if err := gitreceive.Run(cnf, fs, env, storageDriver); err != nil {
					log.Printf("Error running git receive hook [%s]", err)
					os.Exit(1)
				}
			},
		},
	}

	app.Run(os.Args)
}
