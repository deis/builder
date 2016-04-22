package pkg

import (
	"github.com/Masterminds/cookoo"
	"github.com/deis/builder/pkg/env"
	"github.com/deis/builder/pkg/git"
	"github.com/deis/builder/pkg/sshd"
)

// routes builds the Cookoo registry.
//
// Esssentially this is a list of all of the things that Builder can do, broken
// down into a step-by-step list.
func routes(reg *cookoo.Registry) {

	// The "boot" route starts up the builder as a daemon process. Along the
	// way, it starts and configures multiple services, including sshd.
	reg.AddRoute(cookoo.Route{
		Name: "boot",
		Help: "Boot the builder",
		Does: []cookoo.Task{

			// SSHD: Configure host keys.
			cookoo.Cmd{
				Name: sshd.HostKeys,
				Fn:   sshd.ParseHostKeys,
			},
			cookoo.Cmd{
				Name: sshd.ServerConfig,
				Fn:   sshd.Configure,
			},

			// If there's an EXTERNAL_PORT, we publish info to etcd.
			cookoo.Cmd{
				Name: "externalport",
				Fn:   env.Get,
				Using: []cookoo.Param{
					{Name: "EXTERNAL_PORT", DefaultValue: ""},
				},
			},

			// DAEMON: Finally, we wait around for a signal, and then cleanup.
			cookoo.Cmd{
				Name: "listen",
				Fn:   KillOnExit,
				Using: []cookoo.Param{
					{Name: "sshd", From: "cxt:sshdstart"},
				},
			},
		},
	})

	// This provides a very basic SSH ping.
	// Called by the sshd.Server
	reg.AddRoute(cookoo.Route{
		Name: "sshPing",
		Help: "Handles an ssh exec ping.",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "ping",
				Fn:   sshd.Ping,
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
			// Auth against the keys
			cookoo.Cmd{
				Name: "authN",
				Fn:   sshd.AuthKey,
				Using: []cookoo.Param{
					{Name: "metadata", From: "cxt:metadata"},
					{Name: "key", From: "cxt:key"},
					{Name: "repoName", From: "cxt:repository"},
				},
			},
		},
	})

	// This proxies a client session into a git receive.
	//
	// Called by the sshd.Server
	reg.AddRoute(cookoo.Route{
		Name: "sshGitReceive",
		Help: "Handle a git receive over an SSH connection.",
		Does: []cookoo.Task{
			cookoo.Cmd{
				Name: "receive",
				Fn:   git.Receive,
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
}
