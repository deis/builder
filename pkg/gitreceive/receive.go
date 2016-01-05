package gitreceive

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type errInvalidHTTPStatusCode struct {
	expected int
	actual   int
}

func (e errInvalidHTTPStatusCode) Error() string {
	return fmt.Sprintf("expected status code %d, got %d", e.expected, e.actual)
}

func receive(conf *Config, newRev string) error {
	urlStr := fmt.Sprintf("http://%s:%s/v2/hooks/push", conf.WorkflowHost, conf.WorkflowPort)
	bodyMap := map[string]string{
		"receive_user":         conf.Username,
		"receive_repo":         conf.App,
		"sha":                  conf.SHA,
		"fingerprint":          conf.Fingerprint,
		"ssh_connection":       conf.SSHConnection,
		"ssh_original_command": conf.SSHOriginalCommand,
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(bodyMap); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", urlStr, &body)
	if err != nil {
		return err
	}

	// TODO: use ctxhttp here (https://godoc.org/golang.org/x/net/context/ctxhttp)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return errInvalidHTTPStatusCode{expected: 201, actual: resp.StatusCode}
	}
	return nil
}
