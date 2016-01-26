package gitreceive

import (
	"strings"
)

type Config struct {
	// k8s service discovery env vars
	WorkflowHost string `envconfig:"DEIS_WORKFLOW_SERVICE_HOST" required:"true"`
	WorkflowPort string `envconfig:"DEIS_WORKFLOW_SERVICE_PORT" required:"true"`
	RegistryHost string `envconfig:"DEIS_REGISTRY_SERVICE_HOST" required:"true"`
	RegistryPort string `envconfig:"DEIS_REGISTRY_SERVICE_PORT" required:"true"`

	GitHome            string `envconfig:"GIT_HOME" required:"true"`
	SSHConnection      string `envconfig:"SSH_CONNECTION" required:"true"`
	SSHOriginalCommand string `envconfig:"SSH_ORIGINAL_COMMAND" required:"true"`
	Repository         string `envconfig:"REPOSITORY" required:"true"`
	Username           string `envconfig:"USERNAME" required:"true"`
	Fingerprint        string `envconfig:"FINGERPRINT" required:"true"`
	PodNamespace       string `envconfig:"POD_NAMESPACE" required:"true"`
	Debug              bool   `envconfig:"DEBUG" default:"false"`
}

func (c Config) App() string {
	li := strings.LastIndex(c.Repository, ".")
	if li == -1 {
		return c.Repository
	}
	return c.Repository[0:li]
}
