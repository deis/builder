package controller

import (
	"fmt"

	"github.com/deis/builder/pkg/conf"
	deis "github.com/deis/controller-sdk-go"
	"github.com/deis/pkg/log"
)

// New creates a new SDK client configured as the builder.
func New(host, port, builderKeyPath string) (*deis.Client, error) {

	client, err := deis.New(true, fmt.Sprintf("http://%s:%s/", host, port), "")
	if err != nil {
		return client, err
	}
	client.UserAgent = "deis-builder"

	builderKey, err := conf.GetBuilderKey(builderKeyPath)
	if err != nil {
		return client, err
	}
	client.HooksToken = builderKey

	return client, nil
}

// CheckAPICompat checks for API compatibility errors and warns about them.
func CheckAPICompat(c *deis.Client, err error) error {
	if err == deis.ErrAPIMismatch {
		log.Info("WARNING: SDK and Controller API versions do not match. SDK: %s Controller: %s",
			deis.APIVersion, c.ControllerAPIVersion)

		// API mismatch isn't fatal, so after warning continue on.
		return nil
	}

	return err
}
