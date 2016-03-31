package k8s

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
)

var (
	emptyPullPolicy = api.PullPolicy("")
	// ValidPullPolicies is the set of pull policies that this package considers valid
	ValidPullPolicies = map[api.PullPolicy]struct{}{
		api.PullAlways:       struct{}{},
		api.PullIfNotPresent: struct{}{},
		api.PullNever:        struct{}{},
	}
)

// ErrInvalidPullPolicy is the error returned when trying to convert an unknown string to an api.PullPolicy
type ErrInvalidPullPolicy struct {
	str string
}

// Error is the error interface implementation
func (e ErrInvalidPullPolicy) Error() string {
	return fmt.Sprintf("%s is an invalid pull policy", e.str)
}

// PullPolicyFromString converts a string into an api.PullPolicy. returns an error if the string does not match a pull policy in ValidPullPolicies()
func PullPolicyFromString(ppStr string) (api.PullPolicy, error) {
	candidatePP := api.PullPolicy(ppStr)
	if _, ok := ValidPullPolicies[candidatePP]; !ok {
		return emptyPullPolicy, ErrInvalidPullPolicy{str: ppStr}
	}
	return candidatePP, nil
}
