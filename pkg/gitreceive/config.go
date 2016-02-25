package gitreceive

import (
	"strings"
	"time"
)

const (
	builderPodTick    = 100
	objectStorageTick = 500
)

type Config struct {
	// k8s service discovery env vars
	WorkflowHost string `envconfig:"DEIS_WORKFLOW_SERVICE_HOST" required:"true"`
	WorkflowPort string `envconfig:"DEIS_WORKFLOW_SERVICE_PORT" required:"true"`
	RegistryHost string `envconfig:"DEIS_REGISTRY_SERVICE_HOST" required:"true"`
	RegistryPort string `envconfig:"DEIS_REGISTRY_SERVICE_PORT" required:"true"`

	GitHome                       string `envconfig:"GIT_HOME" required:"true"`
	SSHConnection                 string `envconfig:"SSH_CONNECTION" required:"true"`
	SSHOriginalCommand            string `envconfig:"SSH_ORIGINAL_COMMAND" required:"true"`
	Repository                    string `envconfig:"REPOSITORY" required:"true"`
	Username                      string `envconfig:"USERNAME" required:"true"`
	Fingerprint                   string `envconfig:"FINGERPRINT" required:"true"`
	PodNamespace                  string `envconfig:"POD_NAMESPACE" required:"true"`
	StorageRegion                 string `envconfig:"STORAGE_REGION" default:"us-east-1"`
	Debug                         bool   `envconfig:"DEBUG" default:"false"`
	BuilderPodTickDurationMSec    int    `envconfig:"BUILDER_POD_TICK_DURATION" default:"50"`
	BuilderPodTickEndDurationMSec int    `envconfig:"BUILDER_POD_TICK_END_DURATION" default:"10"`
	BuilderPodWaitDurationMSec    int    `envconfig:"BUILDER_POD_WAIT_DURATION" default:"300000"` // 5 minutes
	ObjectStorageTickDurationMSec int    `envconfing:"OBJECT_STORAGE_TICK_DURATION" default:"500"`
	ObjectStorageWaitDurationMSec int    `envconfig:"OBJECT_STORAGE_WAIT_DURATION" default:"300000"` // 5 minutes
	Bucket                        string `envconfig:"BUCKET" default:"git"`
}

func (c Config) App() string {
	li := strings.LastIndex(c.Repository, ".")
	if li == -1 {
		return c.Repository
	}
	return c.Repository[0:li]
}

// BuilderPodTickDuration returns the size of the interval used to check for
// the end of the execution of a Pod building an application
func (c Config) BuilderPodTickDuration() time.Duration {
	return time.Duration(time.Duration(c.BuilderPodTickDurationMSec) * time.Millisecond)
}

// BuilderPodWaitDuration returns the maximum time to wait for the end
// of the execution of a Pod building an application
func (c Config) BuilderPodWaitDuration() time.Duration {
	return time.Duration(time.Duration(c.BuilderPodWaitDurationMSec) * time.Millisecond)
}

// BuilderPodTickEndDuration returns the size of the interval used to check for
// the succeeded status of a pod
func (c Config) BuilderPodTickEndDuration() time.Duration {
	return time.Duration(time.Duration(c.BuilderPodTickEndDurationMSec) * time.Millisecond)
}

// ObjectStorageTickDuration returns the size of the interval used to check for
// the end of an operation that involves the object storage
func (c Config) ObjectStorageTickDuration() time.Duration {
	return time.Duration(time.Duration(c.ObjectStorageTickDurationMSec) * time.Millisecond)
}

// ObjectStorageWaitDuration returns the maximum time to wait for the end of an
// operation that involves the object storage
func (c Config) ObjectStorageWaitDuration() time.Duration {
	return time.Duration(time.Duration(c.ObjectStorageWaitDurationMSec) * time.Millisecond)
}

// CheckDurations checks if ticks for builder and object storage are not bigger
// than the maximum duration. In case of this it will set the tick to the default
func (c *Config) CheckDurations() {
	if c.BuilderPodTickDurationMSec >= c.BuilderPodWaitDurationMSec {
		c.BuilderPodTickDurationMSec = builderPodTick
	}
	if c.BuilderPodTickDurationMSec < builderPodTick {
		c.BuilderPodTickDurationMSec = builderPodTick
	}

	if c.ObjectStorageTickDurationMSec >= c.ObjectStorageWaitDurationMSec {
		c.ObjectStorageTickDurationMSec = objectStorageTick
	}
	if c.ObjectStorageTickDurationMSec < objectStorageTick {
		c.ObjectStorageTickDurationMSec = objectStorageTick
	}
}
