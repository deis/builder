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

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/log"
	pkglog "github.com/deis/builder/pkg/log"
	"golang.org/x/crypto/ssh"
)

// prereceiveHookTplStr is the template for a pre-receive hook. The following template variables are passed into it:
//
// 	.GitHome: the path to Git's home directory.
const preReceiveHookTplStr = `#!/bin/bash
GIT_HOME={{.GitHome}} \
SSH_CONNECTION="$SSH_CONNECTION" \
SSH_ORIGINAL_COMMAND="$SSH_ORIGINAL_COMMAND" \
REPOSITORY="$RECEIVE_REPO" \
USERNAME="$RECEIVE_USER" \
FINGERPRINT="$RECEIVE_FINGERPRINT" \
POD_NAMESPACE="$POD_NAMESPACE" \
boot git-receive
`

var preReceiveHookTpl = template.Must(template.New("hooks").Parse(preReceiveHookTplStr))

// Receive receives a Git repo.
// This will only work for git-receive-pack.
//
// Params:
// 	- operation (string): e.g. git-receive-pack
// 	- repoName (string): The repository name, in the form '/REPO.git'.
// 	- channel (ssh.Channel): The channel.
// 	- request (*ssh.Request): The channel.
// 	- gitHome (string): Defaults to /home/git.
// 	- fingerprint (string): The fingerprint of the user's SSH key.
// 	- user (string): The name of the Deis user.
//
// Returns:
// 	- nothing
func Receive(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	if ok, z := p.Requires("channel", "request", "fingerprint", "permissions"); !ok {
		return nil, fmt.Errorf("Missing requirements %q", z)
	}
	repoName := p.Get("repoName", "").(string)
	operation := p.Get("operation", "").(string)
	channel := p.Get("channel", nil).(ssh.Channel)
	gitHome := p.Get("gitHome", "/home/git").(string)
	fingerprint := p.Get("fingerprint", nil).(string)
	user := p.Get("user", "").(string)

	pkglog.Debug("receiving git repo name: %s, operation: %s, fingerprint: %s, user: %s", repoName, operation, fingerprint, user)

	repo, err := cleanRepoName(repoName)
	if err != nil {
		log.Warnf(c, "Illegal repo name: %s.", err)
		channel.Stderr().Write([]byte("No repo given"))
		return nil, err
	}
	repo += ".git"

	if _, err := createRepo(c, filepath.Join(gitHome, repo), gitHome); err != nil {
		log.Infof(c, "Did not create new repo: %s", err)
	}
	cmd := exec.Command("git-shell", "-c", fmt.Sprintf("%s '%s'", operation, repo))
	log.Infof(c, strings.Join(cmd.Args, " "))

	var errbuff bytes.Buffer

	cmd.Dir = gitHome
	cmd.Env = []string{
		fmt.Sprintf("RECEIVE_USER=%s", user),
		fmt.Sprintf("RECEIVE_REPO=%s", repo),
		fmt.Sprintf("RECEIVE_FINGERPRINT=%s", fingerprint),
		fmt.Sprintf("SSH_ORIGINAL_COMMAND=%s '%s'", operation, repo),
		fmt.Sprintf("SSH_CONNECTION=%s", c.Get("SSH_CONNECTION", "0 0 0 0").(string)),
	}
	cmd.Env = append(cmd.Env, os.Environ()...)

	pkglog.Debug("Working Dir: %s", cmd.Dir)
	pkglog.Debug("Environment: %s", strings.Join(cmd.Env, ","))

	plumbCommand(cmd, channel, &errbuff)

	if err := cmd.Start(); err != nil {
		err = fmt.Errorf("Failed to start git pre-receive hook: %s (%s)", err, errbuff.Bytes())
		log.Warnf(c, err.Error())
		return nil, err
	}
	fmt.Println("Waiting for git-receive to run.")
	fmt.Println("Waiting for deploy.")
	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Failed to run git pre-receive hook: %s (%s)", errbuff.Bytes(), err)
		log.Errf(c, err.Error())
		return nil, err
	}
	if errbuff.Len() > 0 {
		log.Warnf(c, "Unreported error: %s", errbuff.Bytes())
	}
	log.Infof(c, "Deploy complete.\n")

	return nil, nil
}

// cleanRepoName cleans a repository name for a git-sh operation.
func cleanRepoName(name string) (string, error) {
	if len(name) == 0 {
		return name, errors.New("Empty repo name.")
	}
	if strings.Contains(name, "..") {
		return "", errors.New("Cannot change directory in file name.")
	}
	name = strings.Replace(name, "'", "", -1)
	return strings.TrimPrefix(strings.TrimSuffix(name, ".git"), "/"), nil
}

// plumbCommand connects the exec in/output and the channel in/output.
//
// The sidechannel is for sending errors to logs.
func plumbCommand(cmd *exec.Cmd, channel ssh.Channel, sidechannel io.Writer) {
	inpipe, _ := cmd.StdinPipe()
	go func() {
		io.Copy(inpipe, channel)
		inpipe.Close()
	}()

	cmd.Stdout = channel
	cmd.Stderr = io.MultiWriter(channel.Stderr(), sidechannel)
}

var createLock sync.Mutex

// createRepo creates a new Git repo if it is not present already.
//
// Largely inspired by gitreceived from Flynn.
//
// Returns a bool indicating whether a project was created (true) or already
// existed (false).
func createRepo(c cookoo.Context, repoPath, gitHome string) (bool, error) {
	createLock.Lock()
	defer createLock.Unlock()

	fi, err := os.Stat(repoPath)
	if err == nil && fi.IsDir() {
		// Nothing to do.
		log.Infof(c, "Directory %s already exists.", repoPath)
		return false, nil
	} else if os.IsNotExist(err) {
		log.Infof(c, "Creating new directory at %s", repoPath)
		// Create directory
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			log.Warnf(c, "Failed to create repository: %s", err)
			return false, err
		}
		cmd := exec.Command("git", "init", "--bare")
		cmd.Dir = repoPath
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Warnf(c, "git init output: %s", out)
			return false, err
		}

		// parse & generate the template anew each receive for each new git home
		var hookByteBuf bytes.Buffer
		if err := preReceiveHookTpl.Execute(&hookByteBuf, map[string]string{"GitHome": gitHome}); err != nil {
			return true, err
		}

		writePath := filepath.Join(repoPath, "hooks", "pre-receive")
		pkglog.Debug("Writing pre-receive hook to %s", writePath)
		if err := ioutil.WriteFile(writePath, hookByteBuf.Bytes(), 0755); err != nil {
			return false, fmt.Errorf("Cannot write pre-receive hook to %s (%s)", writePath, err)
		}

		return true, nil
	} else if err == nil {
		return false, errors.New("Expected directory, found file.")
	}
	return false, err
}
