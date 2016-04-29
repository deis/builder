package sshd

import (
	"time"
)

// Config represents the required SSH server configuration.
type Config struct {
	SSHHostIP                    string `envconfig:"SSH_HOST_IP" default:"0.0.0.0" required:"true"`
	SSHHostPort                  int    `envconfig:"SSH_HOST_PORT" default:"2223" required:"true"`
	HealthSrvPort                int    `envconfig:"HEALTH_SERVER_PORT" default:"8092"`
	HealthSrvTestStorageRegion   string `envconfig:"STORAGE_REGION" default:"us-east-1"`
	CleanerPollSleepDurationSec  int    `envconfig:"CLEANER_POLL_SLEEP_DURATION_SEC" default:"5"`
	StorageType                  string `envconfig:"BUILDER_STORAGE" default:"minio"`
	SlugBuilderImagePullPolicy   string `envconfig:"SLUG_BUILDER_IMAGE_PULL_POLICY" default:"Always"`
	DockerBuilderImagePullPolicy string `envconfig:"DOCKER_BUILDER_IMAGE_PULL_POLICY" default:"Always"`
}

// CleanerPollSleepDuration returns c.CleanerPollSleepDurationSec as a time.Duration.
func (c Config) CleanerPollSleepDuration() time.Duration {
	return time.Duration(c.CleanerPollSleepDurationSec) * time.Second
}
