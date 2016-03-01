package sshd

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/cookoo"
	"github.com/arschles/assert"
	"github.com/deis/builder/pkg/cleaner"
	"github.com/deis/builder/pkg/controller"
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

	cfg := ssh.ServerConfig{
		NoClientAuth: true,
	}
	cfg.AddHostKey(key)

	c := NewCircuit()
	pushLock := NewInMemoryRepositoryLock()
	cleanerRef := cleaner.NewRef()
	cxt := runServer(&cfg, c, pushLock, cleanerRef, testingServerAddr, t)

	// Give server time to initialize.
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, c.State(), ClosedState, "circuit state")

	// Connect to the server and issue env var set. This should return true.
	client, err := ssh.Dial("tcp", testingServerAddr, &ssh.ClientConfig{})
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

	closer := cxt.Get("sshd.Closer", nil).(chan interface{})
	closer <- true
}

func TestPush(t *testing.T) {
	const testingServerAddr = "127.0.0.1:2245"
	key, err := sshTestingHostKey()
	assert.NoErr(t, err)

	cfg := ssh.ServerConfig{
		NoClientAuth: true,
	}
	cfg.AddHostKey(key)

	c := NewCircuit()
	pushLock := NewInMemoryRepositoryLock()
	cleanerRef := cleaner.NewRef()
	runServer(&cfg, c, pushLock, cleanerRef, testingServerAddr, t)

	// Give server time to initialize.
	time.Sleep(200 * time.Millisecond)

	if c.State() != ClosedState {
		t.Fatalf("circuit was not in closed state")
	}

	// Connect to the server and issue env var set. This should return true.
	client, err := ssh.Dial("tcp", testingServerAddr, &ssh.ClientConfig{})
	if err != nil {
		t.Fatalf("Failed to connect client to local server: %s", err)
	}
	sess, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create client session: %s", err)
	}

	// check for invalid length of arguments
	if out, err := sess.Output("git-upload-pack"); err == nil {
		t.Errorf("Expected an error but '%s' was received", out)
	} else if string(out) != "" {
		t.Errorf("Expected , got '%s'", out)
	}
	sess.Close()

	go func() {
		sess, err = client.NewSession()
		if err != nil {
			t.Fatalf("Failed to create client session: %s", err)
		}
		if out, err := sess.Output("git-upload-pack /demo.git"); err != nil {
			t.Errorf("Unexpected error %s, Output '%s'", err, out)
		} else if string(out) != "OK" {
			t.Errorf("Expected 'OK' got '%s'", out)
		}
		sess.Close()
	}()

	time.Sleep(2 * time.Second)

	sess, err = client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create client session: %s", err)
	}
	if out, err := sess.Output("git-upload-pack /demo.git"); err == nil {
		t.Errorf("Expected an error but returned without errors '%s'", out)
	}
	sess.Close()
}

func TestManyConcurrentPushes(t *testing.T) {
	const testingServerAddr = "127.0.0.1:2247"
	key, err := sshTestingHostKey()
	if err != nil {
		t.Fatal(err)
	}
	cfg := ssh.ServerConfig{NoClientAuth: true}
	cfg.AddHostKey(key)
	c := NewCircuit()
	pushLock := NewInMemoryRepositoryLock()
	cleanerRef := cleaner.NewRef()
	runServer(&cfg, c, pushLock, cleanerRef, testingServerAddr, t)
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, c.State(), ClosedState, "circuit state")

	// Connect to the server and issue env var set. This should return true.
	client, err := ssh.Dial("tcp", testingServerAddr, &ssh.ClientConfig{})
	assert.NoErr(t, err)

	const numRepos = 20
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
	assert.NoErr(t, waitWithTimeout(&wg, 1*time.Second))
}

func TestWithCleanerLock(t *testing.T) {
	srv := &server{cleanerRef: cleaner.NewRef()}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			err := srv.withCleanerLock(func() error {
				wg.Done()
				return nil
			})
			assert.NoErr(t, err)
		}(i)
	}
	assert.NoErr(t, waitWithTimeout(&wg, 1*time.Second))
}

func TestDelete(t *testing.T) {
	const testingServerAddr = "127.0.0.1:2246"
	key, err := sshTestingHostKey()
	assert.NoErr(t, err)

	cfg := ssh.ServerConfig{NoClientAuth: true}
	cfg.AddHostKey(key)

	c := NewCircuit()
	pushLock := NewInMemoryRepositoryLock()
	cleanerRef := cleaner.NewRef()
	runServer(&cfg, c, pushLock, cleanerRef, testingServerAddr, t)

	// Give server time to initialize.
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, c.State(), ClosedState, "circuit state")

	// Connect to the server and issue env var set. This should return true.
	client, err := ssh.Dial("tcp", testingServerAddr, &ssh.ClientConfig{})
	assert.NoErr(t, err)
	sess, err := client.NewSession()
	assert.NoErr(t, err)

	repoName := "demo"
	cleanerRef.Lock()

	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		sess, err = client.NewSession()
		assert.NoErr(t, err)
		// this is expected to hang because cleanerRef is locked
		out, err := sess.Output("git-upload-pack /" + repoName + ".git")
		assert.NoErr(t, err)
		assert.Equal(t, string(out), "OK", "output")
		sess.Close()
	}()

	select {
	case <-doneCh:
		t.Fatalf("push succeeded while cleaner was locked")
	case <-time.After(100 * time.Millisecond):
	}
}

// sshTestingHostKey loads the testing key.
func sshTestingHostKey() (ssh.Signer, error) {
	return ssh.ParsePrivateKey([]byte(testingHostKey))
}

func runServer(config *ssh.ServerConfig, c *Circuit, pushLock RepositoryLock, cleanerRef cleaner.Ref, testAddr string, t *testing.T) cookoo.Context {
	reg, router, cxt := cookoo.Cookoo()
	cxt.Put(ServerConfig, config)
	cxt.Put(Address, testAddr)
	cxt.Put("cookoo.Router", router)

	reg.AddRoute(cookoo.Route{
		Name: "sshPing",
		Help: "Handles an ssh exec ping.",
		Does: cookoo.Tasks{
			cookoo.Cmd{
				Name: "ping",
				Fn:   Ping,
				Using: []cookoo.Param{
					{Name: "request", From: "cxt:request"},
					{Name: "channel", From: "cxt:channel"},
				},
			},
		},
	})

	reg.AddRoute(cookoo.Route{
		Name: "pubkeyAuth",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "authN",
				Fn:   mockAuthKey,
				Using: []cookoo.Param{
					{Name: "metadata", From: "cxt:metadata"},
					{Name: "key", From: "cxt:key"},
					{Name: "repoName", From: "cxt:repository"},
				},
			},
		},
	})

	reg.AddRoute(cookoo.Route{
		Name: "sshGitReceive",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "receive",
				Fn:   mockDummyReceive,
				Using: []cookoo.Param{
					{Name: "request", From: "cxt:request"},
					{Name: "channel", From: "cxt:channel"},
					{Name: "operation", From: "cxt:operation"},
					{Name: "repoName", From: "cxt:repository"},
					{Name: "permissions", From: "cxt:authN"},
					{Name: "userinfo", From: "cxt:userinfo"},
				},
			},
		},
	})

	go func() {
		if err := Serve(reg, router, c, gitHome, pushLock, cleanerRef, cxt); err != nil {
			t.Fatalf("Failed serving with %s", err)
		}
	}()

	return cxt
}

func mockAuthKey(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	c.Put("userinfo", &controller.UserInfo{
		Username:    "deis",
		Key:         testingClientPubKey,
		Fingerprint: "",
		Apps:        []string{"demo"},
	})

	perm := &ssh.Permissions{
		Extensions: map[string]string{
			"user": "deis",
		},
	}
	return perm, nil
}

func mockDummyReceive(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	channel := p.Get("channel", nil).(ssh.Channel)
	req := p.Get("request", nil).(*ssh.Request)
	time.Sleep(5 * time.Second)
	channel.Write([]byte("OK"))
	sendExitStatus(0, channel)
	req.Reply(true, nil)
	return nil, nil
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
