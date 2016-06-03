package sshd

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/arschles/assert"
	"golang.org/x/crypto/ssh"
)

const (
	gitHome = "/git"
)

func TestGitPktLine(t *testing.T) {
	b := new(bytes.Buffer)
	str := "hello world"
	err := gitPktLine(b, str)
	assert.NoErr(t, err)

	outStr := string(b.Bytes())
	assert.True(t, len(outStr) > 4, "output string <= 4 chars")
	assert.Equal(t, outStr[:4], fmt.Sprintf("%04x", len(str)+4), "hex prefix")
	assert.Equal(t, outStr[4:], str, "remainder of string")
}

func serverConfigure() (*ssh.ServerConfig, error) {
	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return mockAuthKey()
		},
	}
	return cfg, nil
}

func clientConfig() *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: "username",
		Auth: []ssh.AuthMethod{
			ssh.Password("password"),
		},
	}
}

// TestServer tests the SSH server.
//
// This listens on the non-standard port 2244 of localhost. This will generate
// an entry in your known_hosts file, and will tie that to the testing key
// used here. It's not recommended that you try to start another SSH server on
// the same port (at a later time) or else you will have key issues that you
// must manually resolve.
func TestReceive(t *testing.T) {
	const testingServerAddr = "127.0.0.1:2244"
	key, err := sshTestingHostKey()
	assert.NoErr(t, err)

	cfg, err := serverConfigure()
	assert.NoErr(t, err)
	cfg.AddHostKey(key)

	c := NewCircuit()
	pushLock := NewInMemoryRepositoryLock(0)
	runServer(cfg, c, pushLock, testingServerAddr, time.Duration(0), t)

	// Give server time to initialize.
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, c.State(), ClosedState, "circuit state")

	// Connect to the server and issue env var set. This should return true.
	client, err := ssh.Dial("tcp", testingServerAddr, clientConfig())
	if err != nil {
		t.Fatalf("Failed to connect client to local server: %s", err)
	}
	sess, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create client session: %s", err)
	}
	defer sess.Close()

	if err := sess.Setenv("HELLO", "world"); err != nil {
		t.Fatal(err)
	}

	if out, err := sess.Output("ping"); err != nil {
		t.Errorf("Output '%s' Error %s", out, err)
	} else if string(out) != "pong" {
		t.Errorf("Expected 'pong', got '%s'", out)
	}

	// Create a new session because the success of the last one closed the
	// connection.
	sess, err = client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create client session: %s", err)
	}
	if err := sess.Run("illegal"); err == nil {
		t.Fatalf("expected a failed run with command 'illegal'")
	}
	if err := sess.Run("illegal command"); err == nil {
		t.Fatalf("expected a failed run with command 'illegal command'")
	}

}

// TestPushInvalidArgsLength tests trying to do a push with only the command, not the repo
func TestPushInvalidArgsLength(t *testing.T) {
	const testingServerAddr = "127.0.0.1:2252"
	key, err := sshTestingHostKey()
	assert.NoErr(t, err)

	cfg, err := serverConfigure()
	assert.NoErr(t, err)
	cfg.AddHostKey(key)

	c := NewCircuit()
	pushLock := NewInMemoryRepositoryLock(0)
	runServer(cfg, c, pushLock, testingServerAddr, 0*time.Second, t)

	// Give server time to initialize.
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, c.State(), ClosedState, "circuit state")

	// Connect to the server and issue env var set. This should return true.
	client, err := ssh.Dial("tcp", testingServerAddr, clientConfig())
	assert.NoErr(t, err)

	// check for invalid length of arguments
	sess, err := client.NewSession()
	assert.NoErr(t, err)
	defer sess.Close()
	if out, err := sess.Output("git-upload-pack"); err == nil {
		t.Errorf("Expected an error but '%s' was received", out)
	} else if string(out) != "" {
		t.Errorf("Expected , got '%s'", out)
	}
}

// TestConcurrentPushSameRepo tests many concurrent pushes, each to the same repo
func TestConcurrentPushSameRepo(t *testing.T) {
	t.Skip("skipping because the global lock prevents testing the repository lock, for multiple concurrent pushes to the same repo")
	t.SkipNow()
	const testingServerAddr = "127.0.0.1:2245"
	key, err := sshTestingHostKey()
	assert.NoErr(t, err)

	cfg, err := serverConfigure()
	assert.NoErr(t, err)
	cfg.AddHostKey(key)

	c := NewCircuit()
	pushLock := NewInMemoryRepositoryLock(0)
	runServer(cfg, c, pushLock, testingServerAddr, 2*time.Second, t)

	// Give server time to initialize.
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, c.State(), ClosedState, "circuit state")

	// Connect to the server and issue env var set. This should return true.
	client, err := ssh.Dial("tcp", testingServerAddr, clientConfig())
	assert.NoErr(t, err)

	const numPushers = 10
	outCh := make(chan *sshSessionOutput, numPushers)
	for i := 0; i < numPushers; i++ {
		go func() {
			sess, err := client.NewSession()
			assert.NoErr(t, err)
			defer sess.Close()
			out, err := sess.Output("git-upload-pack /demo.git")
			outCh <- &sshSessionOutput{outStr: string(out), err: err}
		}()
	}

	foundOK := false
	to := 1 * time.Second
	multiPushLine, err := gitPktLineStr(multiplePush)
	assert.NoErr(t, err)
	for i := 0; i < numPushers; i++ {
		select {
		case sessOut := <-outCh:
			output := sessOut.outStr
			err := sessOut.err
			if output != multiPushLine && output != "OK" {
				t.Fatalf("expected 'OK' or '%s', but got '%s' (error '%s')", multiPushLine, output, err)
			}

			if sessOut.err != nil {
				t.Fatalf("found '%s' output with an error '%s'", output, err)
			}

			if !foundOK && output == "OK" {
				foundOK = true
			} else if output == "OK" {
				t.Fatalf("found second 'OK' when shouldn't have")
			}

		case <-time.After(to):
			t.Fatalf("didn't receive an output within %s", to)
		}
	}
}

// TestConcurrentPushDifferentRepo tests many concurrent pushes, each to a different repo
func TestConcurrentPushDifferentRepo(t *testing.T) {
	const testingServerAddr = "127.0.0.1:2247"
	key, err := sshTestingHostKey()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := serverConfigure()
	assert.NoErr(t, err)
	cfg.AddHostKey(key)
	c := NewCircuit()
	pushLock := NewInMemoryRepositoryLock(time.Duration(1 * time.Minute))
	runServer(cfg, c, pushLock, testingServerAddr, time.Duration(0), t)
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, c.State(), ClosedState, "circuit state")

	// Connect to the server and issue env var set. This should return true.
	client, err := ssh.Dial("tcp", testingServerAddr, clientConfig())
	assert.NoErr(t, err)

	const numRepos = 3
	repoNames := make([]string, numRepos)
	for i := 0; i < numRepos; i++ {
		repoNames[i] = fmt.Sprintf("repo%d", i)
	}
	var wg sync.WaitGroup
	for _, repoName := range repoNames {
		wg.Add(1)
		go func(repoName string) {
			defer wg.Done()
			sess, err := client.NewSession()
			assert.NoErr(t, err)
			out, err := sess.Output("git-upload-pack /" + repoName + ".git")
			assert.NoErr(t, err)
			assert.Equal(t, string(out), "OK", "output")
		}(repoName)
	}
	wg.Wait()
	assert.NoErr(t, waitWithTimeout(&wg, 1*time.Second))
}

// sshTestingHostKey loads the testing key.
func sshTestingHostKey() (ssh.Signer, error) {
	return ssh.ParsePrivateKey([]byte(testingHostKey))
}

func runServer(
	config *ssh.ServerConfig,
	c *Circuit,
	pushLock RepositoryLock,
	testAddr string,
	handlerSleepDur time.Duration,
	t *testing.T) {

	go func() {
		if err := Serve(config, c, gitHome, pushLock, testAddr, "mock"); err != nil {
			t.Fatalf("Failed serving with %s", err)
		}
	}()
}

func mockAuthKey() (*ssh.Permissions, error) {
	perm := &ssh.Permissions{
		Extensions: map[string]string{
			"user":        "deis",
			"fingerprint": "",
			"apps":        "demo,repo1,repo2,repo3,repo4,repo5,repo6,repo7,repo8,repo0,repo9",
		},
	}
	return perm, nil
}

func mockDummyReceive(sleepDur time.Duration) func(channel ssh.Channel, req *ssh.Request) error {
	return func(channel ssh.Channel, req *ssh.Request) error {

		time.Sleep(sleepDur)
		channel.Write([]byte("OK"))
		sendExitStatus(0, channel)
		req.Reply(true, nil)
		return nil
	}
}

func gitPktLineStr(str string) (string, error) {
	var buf bytes.Buffer
	if err := gitPktLine(&buf, str); err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}

// connMetadata mocks ssh.ConnMetadata for authentication.
type connMetadata struct{}

func (cm *connMetadata) User() string          { return "deis" }
func (cm *connMetadata) SessionID() []byte     { return []byte("1") }
func (cm *connMetadata) ClientVersion() []byte { return []byte("2.3.4") }
func (cm *connMetadata) ServerVersion() []byte { return []byte("2.3.4") }
func (cm *connMetadata) RemoteAddr() net.Addr  { return cm.localhost() }
func (cm *connMetadata) LocalAddr() net.Addr   { return cm.localhost() }
func (cm *connMetadata) localhost() net.Addr {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	return addrs[0]
}

var (
	testingHostKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0xOK/wubqj+e4HNp+yAdK4WJnLZCvcjS2DwaxwF+E968kSeU
27SOqiol7Y0UwLGLpB6rpIBnSqXo70xiMUSrnteKmMejddzfbGkvnyvo0dwE4nDd
vnbz64I25xfjTldb4RtNvpk6ymr0soq0EEYssLmdnt7pIgHT71n9RNtu+RPpRe5n
B2ImVeeEsQBhxFsIkkT21JqBhZQRVpeAAOHwainWpkP2MF2ajYUoirs5qOkPxxaw
Mc4i5CSvmFDkWjqkNt84QH9M9M/ws8qX76nImYOPHiF0KRbxamWsYjvdHJCSckdC
mOM7UtsQs8wC3E0xpuPEI0pNRTHCsgH7+KGxmwIDAQABAoIBAAOQufFS7d8zUeiy
qmCeiz+X8todzgTMppsWcNFZuhp10bOV+pK3ew1uxtM7ZdVXamdsSTPvI0+Ee+nG
3YW9hjSZqXKpNJ6iC3gWUsKaiEU7NS3qACTed4JL4ceHhMRm/1tPDcIhbnfK1LVL
WH1J4ileCUaMt11msIDDgV6vYjF81733O+8kPnh5BaFLIOuPdmAPfsZC2WQfBTka
6F5bhe9mcraQohWOGC/NKBbV9o6Ua2GT5ZJILtyPwfx8ctnQHLfmlTOI7qpRyMCU
1hGwlWxyvZRyY4loZehy0c7DaEWJqWS1AST9AbUcNXciYSt/5pUP76W0L6NzwJdh
C1jIY2ECgYEA+JwlIzhsZRsN0jA3A2qWRt3WGdliujAqDvVj4e8E+QnlTh/MDVKF
x3F+w58DHRKJrH7d1nD1fq2id6vh3Sl7xGHZiztOpolY0xlOt71X+2anX+QTEX5Q
d1jB/zQliUsxzIjqn31dKUlAfoI5XiWrxuP1Py8gZSTnnBl8bkdKZysCgYEA2VnG
+bhBdw/0RJVsleyHBrq0+MnQ80dxj6XatKvniVDqjHQefq088W2ULeI5wVjdMy59
CVnDVS6759pLkWu5br7Agb+NGyVKd3o0CT0Jn6JJj9kq1Wq7iOedJF+GtabVp4gk
efIYECkS7BKe1GFH5vRM8FbyyepRFBCgrH1ep1ECgYEAiRojaO7+6CspThcE379y
LJa+MfcueRuCtkkh0kFsbqLEcHccouQ1nq26iMsyfl/wyM4WLOKSoE/FX1XM85ij
BsQnop8MWs83ywMT5ERpNt1/xGQVF/qfCZJLOiBZ6wMq7W88ZMRQEiqxhJLwbDk+
KCsi3rtwlBbsG6v6cR6jq40CgYAzH4nMvQkw7yC+bQMgdIUCETJ1/kpWnqxYZGN/
8ZtBUjYJGVr+4tKd2u9qp3Z8QuGsozen1mQ6igaKr27s4pC4Osfe/OY8x1Wvqp/I
uIGl+a8h1avcjQFVX1036/wsh/RjNoOV51q/mlmoC20ueT9HVJkwQtNSqPmvJYYV
bFuyMQKBgQCsRVEJ6eqai+Pz4bY2UfBnkU6ZHdySI+fQB/T770p0/SbrYMBxNrPQ
v3+ZZfZMlci4pxBtXqrnoyj4uUoqZtR3ENLz53SN1i0vpT7DtC6gMnEF1UWiaoJ6
6mGH5/bxCg9wpV7qpqR0EbFM/dhQFZmmnirOS8x+00hJvc1HFiuN/A==
-----END RSA PRIVATE KEY-----
`

	testingClientPubKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC/OImeiJppXJQY+fKpULj1cvM1FL5M9brc3Diqi8IbyVVvEoYMgcLri0msIOJl3SmkSFj5FAMZo/CswicedXwjB1LXBfbZRNG5cD+heYdwjE7bOZSeuMUOWkqbaj7Zd3XruJ91X0CKo0G2q47QzzzZFobL30ts09yX26ACfGjkNUjWMRKXm9iq2I4CdFK+YmfZz6GQl8pevIfuFTjL5uUMrlXPjh5KwLtuAbdlsp8oZH2aV/ajNWXMw2LYAJnny8MHGflZUtvVs9XUsemJwnTR9TdMNGcrcyTC+8Ceqnvxs3OL6i5ggDBhJnjWIc13n3otAlyGvW+zcWjypuBhotjz donotuse`
)
