package sshd

type Config struct {
	FetcherPort int    `envconfig:"fetcher_port" default:"3000"`
	SSHHostIP   string `envconfig:"ssh_host_ip" default:"0.0.0.0"`
	SSHHostPort int    `envconfig:"ssh_host_port" default:"2223"`
}
