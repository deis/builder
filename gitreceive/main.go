package main

import (
	"bufio"
	"fmt"
	"github.com/helm/helm/log"
	"io"
	"os"
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

const newline = "\n"

// readLine reads from bio until it reaches a "\n". returns the line, not including the "\n"
func getLine(bio *bufio.Reader) (string, error) {
	line, err := bio.ReadString(newline)
	if err != nil {
		return "", err
	}
	if strings.HasSuffix(line, newline) {
		return line[0 : len(line)-len(newline)]
	}
}

func readLine(line string) (string, string, string, error) {
	spl := strings.Split(line, " ")
	if len(spl) != 3 {
		return "", "", "", fmt.Errorf("malformed line [%s]", line)
	}
	return spl[0], spl[1], spl[2], nil
}

func main() {
	conf, err := getConfig("gitreceive")
	if err != nil {
		log.Msg("config error [%s]", err)
		os.Exit(1)
	}
	bio := bufio.NewReader(os.Stdin)
	for line, err := getLine(bio); err != nil; {
		oldRev, newRev, refName, err := readLine(line)
		if err := receive(conf, newRev); err != nil {
			log.Die("failed on rev %s - push denied", newRev)
			os.Exit(1)
		}
		if err := build(conf, newRev); err != nil {
			log.Die("error building %s [%s]", newRev, err)
			os.Exit(1)
		}
	}
}
