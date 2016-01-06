package gitreceive

type Config struct {
	// k8s service discovery env vars
	WorkflowHost string `envconfig:"deis_workflow_service_host"`
	WorkflowPort string `envconfig:"deis_workflow_service_port"`
	RegistryHost string `envconfig:"deis_registry_service_host"`
	RegistryPort string `envconfig:"deis_registry_service_port"`

	GitHome            string `envconfig:"git_home"`
	SSHConnection      string `envconfig:"ssh_connection"`
	SSHOriginalCommand string `envconfig:"ssh_original_command"`
	Repository         string `envconfig:"repository"`
	SHA                string `envconfing:"sha"`
	Username           string `envconfig:"username"`
	App                string `envconfing:"app"`
	Fingerprint        string `envconfing:"fingerprint"`
	PodNamespace       string `envconfig:"pod_namespace"`
}
