package sshd

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	"golang.org/x/crypto/ssh"

	"github.com/deis/builder/pkg/controller"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/log"
)

const (
	builderKeyLocation = "/var/run/secrets/api/auth/builder-key"
)

// ParseHostKeys parses the host key files.
//
// By default it looks in /etc/ssh for host keys of the patterh ssh_host_{{TYPE}}_key.
//
// Params:
// 	- keytypes ([]string): Key types to parse. Defaults to []string{rsa, dsa, ecdsa}
// 	- enableV1 (bool): Allow V1 keys. By default this is disabled.
// 	- path (string): Override the lookup pattern. If %s, it will be replaced with the keytype.
//
// Returns:
// 	[]ssh.Signer
func ParseHostKeys(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	log.Debugf(c, "Parsing ssh host keys")
	hostKeyTypes := p.Get("keytypes", []string{"rsa", "dsa", "ecdsa"}).([]string)
	pathTpl := p.Get("path", "/etc/ssh/ssh_host_%s_key").(string)
	hostKeys := make([]ssh.Signer, 0, len(hostKeyTypes))
	for _, t := range hostKeyTypes {
		path := fmt.Sprintf(pathTpl, t)

		if key, err := ioutil.ReadFile(path); err == nil {
			if hk, err := ssh.ParsePrivateKey(key); err == nil {
				log.Infof(c, "Parsed host key %s.", path)
				hostKeys = append(hostKeys, hk)
			} else {
				log.Errf(c, "Failed to parse host key %s (skipping): %s", path, err)
			}
		}
	}
	if c.Get("enableV1", false).(bool) {
		path := "/etc/ssh/ssh_host_key"
		if key, err := ioutil.ReadFile(path); err != nil {
			log.Errf(c, "Failed to read ssh_host_key")
		} else if hk, err := ssh.ParsePrivateKey(key); err == nil {
			log.Infof(c, "Parsed host key %s.", path)
			hostKeys = append(hostKeys, hk)
		} else {
			log.Errf(c, "Failed to parse host key %s: %s", path, err)
		}
	}
	return hostKeys, nil
}

// AuthKey authenticates based on a public key.
//
// Params:
// 	- metadata (ssh.ConnMetadata)
// 	- key (ssh.PublicKey)
//
// Returns:
// 	*ssh.Permissions
//
func AuthKey(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	log.Debugf(c, "Starting ssh authentication")
	key := p.Get("key", nil).(ssh.PublicKey)

	strKey := string(ssh.MarshalAuthorizedKey(key))
	log.Debugf(c, "Checking auth for user key %v", strKey)
	userInfo, err := controller.UserInfoFromKey(strKey)
	if err != nil {
		return nil, err
	}

	userInfo.Key = strKey
	c.Put("userinfo", userInfo)

	log.Infof(c, "Key accepted for user %s.", userInfo.Username)
	perm := &ssh.Permissions{
		Extensions: map[string]string{
			"user": userInfo.Username,
		},
	}
	return perm, nil
}

// Configure creates a new SSH configuration object.
//
// Config sets a PublicKeyCallback handler that forwards public key auth
// requests to the route named "pubkeyAuth".
//
// This assumes certain details about our environment, like the location of the
// host keys. It also provides only key-based authentication.
// ConfigureServerSshConfig
//
// Returns:
//  An *ssh.ServerConfig
func Configure(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	router := c.Get("cookoo.Router", nil).(*cookoo.Router)

	cfg := &ssh.ServerConfig{
		PublicKeyCallback: func(m ssh.ConnMetadata, k ssh.PublicKey) (*ssh.Permissions, error) {
			c.Put("metadata", m)
			c.Put("key", k)

			pubkeyAuth := c.Get("route.sshd.pubkeyAuth", "pubkeyAuth").(string)
			err := router.HandleRequest(pubkeyAuth, c, true)
			return c.Get("pubkeyAuth", &ssh.Permissions{}).(*ssh.Permissions), err
		},
	}

	return cfg, nil
}

// GenSSHKeys generates the default set of SSH host keys.
func GenSSHKeys(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	log.Debugf(c, "Generating ssh keys for sshd")
	// Generate a new key
	out, err := exec.Command("ssh-keygen", "-A").CombinedOutput()
	if err != nil {
		log.Infof(c, "ssh-keygen: %s", out)
		return nil, err
	}
	return nil, nil
}
