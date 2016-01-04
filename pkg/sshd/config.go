package sshd

type Config struct {
	FetcherPort int    `envconfig:"fetcher_port"`
	SSHHostIP   string `envconfig:"ssh_host_ip"`
	SSHHostPort int    `envconfig:"ssh_host_port"`
}
