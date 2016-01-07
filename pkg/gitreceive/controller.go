package gitreceive

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errControllerNotFound           = errors.New("Deis controller not found. Is it running?")
	errControllerServiceUnavailable = errors.New("Deis controller was unavailable. Is it healthy?")
)

type unexpectedControllerStatusCode struct {
	endpoint string
	expected int
	actual   int
}

func newUnexpectedControllerStatusCode(endpoint string, expectedCode, actualCode int) unexpectedControllerStatusCode {
	return unexpectedControllerStatusCode{endpoint: endpoint, expected: expectedCode, actual: actualCode}
}

func (u unexpectedControllerStatusCode) Error() string {
	return fmt.Sprintf("Deis controller endpoint %s: expected status code %d, got %d", u.endpoint, u.expected, u.actual)
}

func controllerURLStr(conf *Config, additionalPath ...string) string {
	return fmt.Sprintf("http://%s:%s/%s", conf.WorkflowHost, conf.WorkflowPort, strings.Join(additionalPath, "/"))
}
