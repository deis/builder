package gitreceive

import (
	"strings"
	"time"
)

const (
	builderPodTick    = 100
	objectStorageTick = 500
)

// Config is the envconfig (http://github.com/kelseyhightower/envconfig) compatible struct for the
// builder's git-receive hook.
type Config struct {
	// k8s service discovery env vars
	ControllerHost string `envconfig:"DEIS_CONTROLLER_SERVICE_HOST" required:"true"`
	ControllerPort string `envconfig:"DEIS_CONTROLLER_SERVICE_PORT" required:"true"`
	RegistryHost   string `envconfig:"DEIS_REGISTRY_SERVICE_HOST" required:"true"`
	RegistryPort   string `envconfig:"DEIS_REGISTRY_SERVICE_PORT" required:"true"`

	GitHome                       string `envconfig:"GIT_HOME" required:"true"`
	SSHConnection                 string `envconfig:"SSH_CONNECTION" required:"true"`
	SSHOriginalCommand            string `envconfig:"SSH_ORIGINAL_COMMAND" required:"true"`
	Repository                    string `envconfig:"REPOSITORY" required:"true"`
	Username                      string `envconfig:"USERNAME" required:"true"`
	Fingerprint                   string `envconfig:"FINGERPRINT" required:"true"`
	PodNamespace                  string `envconfig:"POD_NAMESPACE" required:"true"`
	StorageRegion                 string `envconfig:"STORAGE_REGION" default:"us-east-1"`
	Debug                         bool   `envconfig:"DEBUG" default:"false"`
	BuilderPodTickDurationMSec    int    `envconfig:"BUILDER_POD_TICK_DURATION" default:"100"`
	BuilderPodWaitDurationMSec    int    `envconfig:"BUILDER_POD_WAIT_DURATION" default:"900000"` // 15 minutes
	ObjectStorageTickDurationMSec int    `envconfing:"OBJECT_STORAGE_TICK_DURATION" default:"500"`
	ObjectStorageWaitDurationMSec int    `envconfig:"OBJECT_STORAGE_WAIT_DURATION" default:"300000"` // 5 minutes
	SessionIdleIntervalMsec       int    `envconfig:"SESSION_IDLE_INTERVAL" default:"10000"`         // 10 seconds
	SlugBuilderImage              string `envconfig:"SLUGBUILDER_IMAGE_NAME" default:"quay.io/deisci/slugbuilder:v2-beta"`
	DockerBuilderImage            string `envconfig:"DOCKERBUILDER_IMAGE_NAME" default:"quay.io/deisci/dockerbuilder:v2-beta"`
	SlugBuilderImagePullPolicy    string `envconfig:"SLUG_BUILDER_IMAGE_PULL_POLICY" default:"Always"`
	DockerBuilderImagePullPolicy  string `envconfig:"DOCKER_BUILDER_IMAGE_PULL_POLICY" default:"Always"`
	StorageType                   string `envconfig:"BUILDER_STORAGE" default:"minio"`
}

// App returns the application name represented by c. The app name is the same as c.Repository
// with the last '.' and beyond stripped off.
func (c Config) App() string {
	li := strings.LastIndex(c.Repository, ".")
	if li == -1 {
		return c.Repository
	}
	return c.Repository[0:li]
}

// BuilderPodTickDuration returns the size of the interval used to check for
// the end of the execution of a Pod building an application.
func (c Config) BuilderPodTickDuration() time.Duration {
	return time.Duration(time.Duration(c.BuilderPodTickDurationMSec) * time.Millisecond)
}

// BuilderPodWaitDuration returns the maximum time to wait for the end
// of the execution of a Pod building an application.
func (c Config) BuilderPodWaitDuration() time.Duration {
	return time.Duration(time.Duration(c.BuilderPodWaitDurationMSec) * time.Millisecond)
}

// ObjectStorageTickDuration returns the size of the interval used to check for
// the end of an operation that involves the object storage.
func (c Config) ObjectStorageTickDuration() time.Duration {
	return time.Duration(time.Duration(c.ObjectStorageTickDurationMSec) * time.Millisecond)
}

// ObjectStorageWaitDuration returns the maximum time to wait for the end of an
// operation that involves the object storage.
func (c Config) ObjectStorageWaitDuration() time.Duration {
	return time.Duration(time.Duration(c.ObjectStorageWaitDurationMSec) * time.Millisecond)
}

// SessionIdleInterval returns the ticker interval to wait for status
func (c Config) SessionIdleInterval() time.Duration {
	return time.Duration(time.Duration(c.SessionIdleIntervalMsec) * time.Millisecond)
}

// CheckDurations checks if ticks for builder and object storage are not bigger
// than the maximum duration. In case of this it will set the tick to the default.
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
