package gitreceive

import (
	"error"
	"strings"
)

var (
	errControllerNotFound           = errors.New("Deis controller not found. Is it running?")
	errControllerServiceUnavailable = errors.New("Deis controller was unavailable. Is it healthy?")
)

type unexpectedControllerStatusCode struct {
	expected int
	actual   int
}

func (u unexpectedControllerStatusCode) Error() string {
	return fmt.Sprintf("Expected status code %d from Deis controller, got %d", u.expected, u.actual)
}

func controllerURLStr(conf *Config, additionalPath ...string) string {
	return fmt.Sprintf("http://%s:%s/%s", conf.WorkflowHost, conf.WorkflowPort, strings.Join(additionalPath, "/"))
}
