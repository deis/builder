package sshd

// Config represents the required SSH server configuration
type Config struct {
	SSHHostIP     string `envconfig:"SSH_HOST_IP" default:"0.0.0.0" required:"true"`
	SSHHostPort   int    `envconfig:"SSH_HOST_PORT" default:"2223" required:"true"`
	HealthSrvPort int    `envconfig:"HEALTH_SERVER_PORT" default:"8092"`
}
