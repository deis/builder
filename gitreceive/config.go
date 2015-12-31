package main

import (
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	WorkflowHost       string `envconfig:"deis_workflow_service_host"`
	WorkflowPort       string `envconfig:"deis_workflow_service_port"`
	GitHome            string `envconfig:"git_home"`
	SSHConnection      string `envconfig:"ssh_connection"`
	SSHOriginalCommand string `envconfig:"ssh_original_command"`
	Repository         string `envconfig:"repository"`
	SHA                string `envconfing:"sha"`
	Username           string `envconfig:"username"`
	App                string `envconfing:"app"`
	Fingerprint        string `envconfing:"fingerprint"`
}

func getConfig(appName string) (*config, error) {
	conf := &config{}
	if err := envconfig(appName, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
