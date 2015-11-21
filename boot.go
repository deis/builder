package main

import (
	"log"
	"os"
	"runtime"

	"github.com/arschles/builder/fetcher"
	"github.com/deis/builder/pkg"
	"github.com/kelseyhightower/envconfig"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

type Config struct {
	FetcherPort int    `envconfig:"fetcher_port"`
	SSHHostIP   string `envconfig:"ssh_host_ip"`
	SSHHostPort int    `envconfig:"ssh_host_port"`
}

func main() {
	var conf Config
	if err := envconfig.Process("builder", &config); err != nil {
		log.Fatalf("error fetching config [%s]", err)
		os.Exit(1)
	}
	log.Printf("starting fetcher on port %d", conf.FetcherPort)
	go fetcher.Serve(conf.FetcherPort)
	os.Exit(pkg.Run("boot"))
}
