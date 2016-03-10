package gitreceive

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/deis/builder/pkg/gitreceive/storage"
	"github.com/deis/builder/pkg/sys"
	"github.com/deis/pkg/log"

	builderconf "github.com/deis/builder/pkg/conf"

	client "k8s.io/kubernetes/pkg/client/unversioned"
)

func readLine(line string) (string, string, string, error) {
	spl := strings.Split(line, " ")
	if len(spl) != 3 {
		return "", "", "", fmt.Errorf("malformed line [%s]", line)
	}
	return spl[0], spl[1], spl[2], nil
}

// Run runs the git-receive hook. This func is effectively the main for the git-receive hook,
// although it is called from the main in boot.go.
func Run(conf *Config, fs sys.FS, env sys.Env) error {
	log.Debug("Running git hook")

	builderKey, err := builderconf.GetBuilderKey()
	if err != nil {
		return err
	}

	s3Client, err := storage.GetClient(conf.StorageRegion, fs, env)
	if err != nil {
		return fmt.Errorf("configuring S3 client (%s)", err)
	}

	kubeClient, err := client.NewInCluster()
	if err != nil {
		return fmt.Errorf("couldn't reach the api server (%s)", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		oldRev, newRev, refName, err := readLine(line)

		if err != nil {
			return fmt.Errorf("reading STDIN (%s)", err)
		}

		log.Debug("read [%s,%s,%s]", oldRev, newRev, refName)

		if err := receive(conf, builderKey, newRev); err != nil {
			return err
		}
		// if we're processing a receive-pack on an existing repo, run a build
		if strings.HasPrefix(conf.SSHOriginalCommand, "git-receive-pack") {
			if err := build(conf, s3Client, kubeClient, fs, env, builderKey, newRev); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
