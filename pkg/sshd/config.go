package sshd

import (
	"time"
)

// Config represents the required SSH server configuration
type Config struct {
	SSHHostIP                  string        `envconfig:"SSH_HOST_IP" default:"0.0.0.0" required:"true"`
	SSHHostPort                int           `envconfig:"SSH_HOST_PORT" default:"2223" required:"true"`
	HealthSrvPort              int           `envconfig:"HEALTH_SERVER_PORT" default:"8092"`
	HealthSrvTestStorageRegion string        `envconfig:"HEALTH_SERVER_TEST_STORAGE_REGION" default:"us-east-1"`
	CleanerPollSleepDuration   time.Duration `envconfig:"CLEANER_POLL_SLEEP_DURATION" default:"1s"`
}
