package main

import (
	"bytes"
	"net/http"
)

// #!/usr/bin/env bash
// set -eo pipefail
//
// repository=$1
// app=${1%.git}
// sha=$2
// username=$3
// fingerprint=$4
//
// curl \
//   -X 'POST' --fail \
//   -H 'Content-Type: application/json' \
//   -H "X-Deis-Builder-Auth: {{ getv "/deis/controller/builderKey" }}" \
//   -d "{\"receive_user\": \"$username\", \"receive_repo\": \"$app\", \"sha\": \"$sha\", \"fingerprint\": \"$fingerprint\", \"ssh_connection\": \"$SSH_CONNECTION\", \"ssh_original_command\": \"$SSH_ORIGINAL_COMMAND\"}" \
//   --silent "http://$DEIS_WORKFLOW_SERVICE_HOST:$DEIS_WORKFLOW_SERVICE_PORT/v2/hooks/push" >/dev/null

func receive(conf *config, newRev string) error {
	urlStr := fmt.Sprintf("http://%s:%s/v2/hooks/push", conf.WorkflowHost, conf.WorkflowPort)
	bodyMap := map[string]string{
		"receive_user":         conf.User,
		"receive_repo":         conf.App,
		"sha":                  conf.SHA,
		"fingerprint":          conf.Fingerprint,
		"ssh_connection":       conf.SSHConnection,
		"ssh_original_command": conf.SSHOriginalCommand,
	}
	var body bytes.Buffer
	if err := json.NewEncoder().Encode(&body, bodyMap); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", urlStr, &body)
	if err != nil {
		return err
	}
	resp, err := http.Do(req)
	if err != nil {
		return err
	}
}
