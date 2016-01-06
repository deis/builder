package etcd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

const (
	hostEnvVar         = "DEIS_ETCD_1_SERVICE_HOST"
	portEnvVar         = "DEIS_ETCD_1_SERVICE_PORT_CLIENT"
	defaultHost        = "localhost"
	defaultPort        = "4001"
	defaultRetryCycles = 2
)

var (
	defaultRetrySleep = 200 * time.Millisecond
)

func CreateClientFromEnv() (*etcd.Client, error) {
	host := os.Getenv(hostEnvVar)
	port := os.Getenv(portEnvVar)
	if host == "" {
		host = defaultHost
	}
	if port == "" {
		port = defaultPort
	}
	hosts := []string{
		fmt.Sprintf("http://%s:%s", host, port),
	}
	client := etcd.NewClient(hosts)
	client.CheckRetry = getCheckRetryFunc(defaultRetryCycles, defaultRetrySleep)

	return client, nil
}

func getCheckRetryFunc(retryCycles int, retrySleep time.Duration) func(*etcd.Cluster, int, http.Response, error) error {
	return func(c *etcd.Cluster, numReqs int, last http.Response, err error) error {
		if numReqs > retryCycles*len(c.Machines) {
			return fmt.Errorf("Tried and failed %d cluster connections: %s", retryCycles, err)
		}

		switch last.StatusCode {
		case 0:
			return nil
		case 500:
			time.Sleep(retrySleep)
			return nil
		case 200:
			return nil
		default:
			return fmt.Errorf("Unhandled HTTP Error: %s %d", last.Status, last.StatusCode)
		}
	}
}
