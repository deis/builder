package gitreceive

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/deis/builder/pkg/gitreceive/etcd"
	"github.com/deis/builder/pkg/log"
)

// #!/bin/bash
// strip_remote_prefix() {
//     stdbuf -i0 -o0 -e0 sed "s/^/"$'\e[1G'"/"
// }
//
// while read oldrev newrev refname
// do
//   LOCKFILE="/tmp/$RECEIVE_REPO.lock"
//   if ( set -o noclobber; echo "$$" > "$LOCKFILE" ) 2> /dev/null; then
// 	trap 'rm -f "$LOCKFILE"; exit 1' INT TERM EXIT
//
// 	# check for authorization on this repo
// 	{{.GitHome}}/receiver "$RECEIVE_REPO" "$newrev" "$RECEIVE_USER" "$RECEIVE_FINGERPRINT"
// 	rc=$?
// 	if [[ $rc != 0 ]] ; then
// 	  echo "      ERROR: failed on rev $newrev - push denied"
// 	  exit $rc
// 	fi
// 	# builder assumes that we are running this script from $GITHOME
// 	cd {{.GitHome}}
// 	# if we're processing a receive-pack on an existing repo, run a build
// 	if [[ $SSH_ORIGINAL_COMMAND == git-receive-pack* ]]; then
// 		{{.GitHome}}/builder "$RECEIVE_USER" "$RECEIVE_REPO" "$newrev" 2>&1 | strip_remote_prefix
// 	fi
//
// 	rm -f "$LOCKFILE"
// 	trap - INT TERM EXIT
//   else
// 	echo "Another git push is ongoing. Aborting..."
// 	exit 1
//   fi
// done

func readLine(line string) (string, string, string, error) {
	spl := strings.Split(line, " ")
	if len(spl) != 3 {
		return "", "", "", fmt.Errorf("malformed line [%s]", line)
	}
	return spl[0], spl[1], spl[2], nil
}

func Run(conf *Config) error {
	log.Debug("Running git hook")

	etcdClient, err := etcd.CreateClientFromEnv()
	if err != nil {
		return err
	}
	// TODO: replace etcd usage here with something else. See https://github.com/deis/builder/issues/81
	builderKey, err := getBuilderKey(etcdClient)
	if err != nil {
		return fmt.Errorf("couldn't get builder key %s (%s)", builderKey, err)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// oldRev, newRev, refName, err := readLine(line)
		_, newRev, _, err := readLine(line)
		if err != nil {
			return err
		}
		if err := receive(conf, newRev); err != nil {
			return err
		}
		if err := build(conf, builderKey, newRev); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
