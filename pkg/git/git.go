package git

// This file just contains the Git-specific portions of sshd.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/deis/pkg/log"
	"golang.org/x/crypto/ssh"
)

// prereceiveHookTplStr is the template for a pre-receive hook. The following template variables
// are passed into it:
//
// - .GitHome: the path to Git's home directory
const preReceiveHookTplStr = `#!/bin/bash
set -eo pipefail
strip_remote_prefix() {
    stdbuf -i0 -o0 -e0 sed "s/^/"$'\e[1G'"/"
}

GIT_HOME={{.GitHome}} \
SSH_CONNECTION="$SSH_CONNECTION" \
SSH_ORIGINAL_COMMAND="$SSH_ORIGINAL_COMMAND" \
REPOSITORY="$RECEIVE_REPO" \
USERNAME="$RECEIVE_USER" \
FINGERPRINT="$RECEIVE_FINGERPRINT" \
POD_NAMESPACE="$POD_NAMESPACE" \
boot git-receive | strip_remote_prefix
`

var preReceiveHookTpl = template.Must(template.New("hooks").Parse(preReceiveHookTplStr))

// Receive receives a Git repo.
// This will only work for git-receive-pack.
func Receive(
	repo, operation, gitHome string,
	channel ssh.Channel,
	fingerprint, username, conndata, receivetype string) error {

	log.Info("receiving git repo name: %s, operation: %s, fingerprint: %s, user: %s", repo, operation, fingerprint, username)

	if receivetype == "mock" {
		channel.Write([]byte("OK"))
		return nil
	}
	repoPath := filepath.Join(gitHome, repo)
	log.Info("creating repo directory %s", repoPath)
	if _, err := createRepo(repoPath); err != nil {
		err = fmt.Errorf("Did not create new repo (%s)", err)

		return err
	}

	log.Info("writing pre-receive hook under %s", repoPath)
	if err := createPreReceiveHook(gitHome, repoPath); err != nil {
		err = fmt.Errorf("Did not write pre-receive hook (%s)", err)
		return err
	}

	cmd := exec.Command("git-shell", "-c", fmt.Sprintf("%s '%s'", operation, repo))
	log.Info(strings.Join(cmd.Args, " "))

	var errbuff bytes.Buffer

	cmd.Dir = gitHome
	cmd.Env = []string{
		fmt.Sprintf("RECEIVE_USER=%s", username),
		fmt.Sprintf("RECEIVE_REPO=%s", repo),
		fmt.Sprintf("RECEIVE_FINGERPRINT=%s", fingerprint),
		fmt.Sprintf("SSH_ORIGINAL_COMMAND=%s '%s'", operation, repo),
		fmt.Sprintf("SSH_CONNECTION=%s", conndata),
	}
	cmd.Env = append(cmd.Env, os.Environ()...)

	log.Debug("Working Dir: %s", cmd.Dir)
	log.Debug("Environment: %s", strings.Join(cmd.Env, ","))

	inpipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Stdout = channel
	cmd.Stderr = io.MultiWriter(channel.Stderr(), &errbuff)

	if err := cmd.Start(); err != nil {
		err = fmt.Errorf("Failed to start git pre-receive hook: %s (%s)", err, errbuff.Bytes())
		return err
	}

	if _, err := io.Copy(inpipe, channel); err != nil {
		err = fmt.Errorf("Failed to write git objects into the git pre-receive hook (%s)", err)
		return err
	}

	fmt.Println("Waiting for git-receive to run.")
	fmt.Println("Waiting for deploy.")
	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Failed to run git pre-receive hook: %s (%s)", errbuff.Bytes(), err)
		return err
	}
	if errbuff.Len() > 0 {
		log.Err("Unreported error: %s", errbuff.Bytes())
		return errors.New(errbuff.String())
	}
	log.Info("Deploy complete.")

	return nil
}

var createLock sync.Mutex

// createRepo creates a new Git repo if it is not present already.
//
// Largely inspired by gitreceived from Flynn.
//
// Returns a bool indicating whether a project was created (true) or already
// existed (false).
func createRepo(repoPath string) (bool, error) {
	createLock.Lock()
	defer createLock.Unlock()

	fi, err := os.Stat(repoPath)
	if err == nil && fi.IsDir() {
		// Nothing to do.
		log.Debug("Directory %s already exists.", repoPath)
		return false, nil
	} else if os.IsNotExist(err) {
		log.Debug("Creating new directory at %s", repoPath)
		// Create directory
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			log.Err("Failed to create repository: %s", err)
			return false, err
		}
		cmd := exec.Command("git", "init", "--bare")
		cmd.Dir = repoPath
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Info("git init output: %s", out)
			return false, err
		}

		return true, nil
	} else if err == nil {
		return false, errors.New("Expected directory, found file.")
	}
	return false, err
}

// createPreReceiveHook renders preReceiveHookTpl to repoPath/hooks/pre-receive
func createPreReceiveHook(gitHome, repoPath string) error {
	// parse & generate the template anew each receive for each new git home
	var hookByteBuf bytes.Buffer
	if err := preReceiveHookTpl.Execute(&hookByteBuf, map[string]string{"GitHome": gitHome}); err != nil {
		return err
	}

	writePath := filepath.Join(repoPath, "hooks", "pre-receive")
	log.Info("Writing pre-receive hook to %s", writePath)
	if err := ioutil.WriteFile(writePath, hookByteBuf.Bytes(), 0755); err != nil {
		return fmt.Errorf("Cannot write pre-receive hook to %s (%s)", writePath, err)
	}
	return nil
}
