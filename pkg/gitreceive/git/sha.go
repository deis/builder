package git

import (
	"fmt"
	"regexp"
)

const (
	// this constant represents the length of a shortened git sha - 8 characters long
	shortShaIdx = 8
)

var shaRegex = regexp.MustCompile(`^[\da-f]{40}$`)

// ErrInvalidGitSha is returned by NewSha if the given raw sha is invalid for any reason
type ErrInvalidGitSha struct {
	sha string
}

// Error is the error interface implementation
func (e ErrInvalidGitSha) Error() string {
	return fmt.Sprintf("git sha %s was invalid", e.sha)
}

// SHA is the representaton of a git sha
type SHA struct {
	full  string
	short string
}

// NewSha creates a raw string to a SHA. Returns ErrInvalidGitSha if the sha was invalid
func NewSha(rawSha string) (*SHA, error) {
	if !shaRegex.MatchString(rawSha) {
		return nil, ErrInvalidGitSha{sha: rawSha}
	}
	return &SHA{full: rawSha, short: rawSha[0:8]}, nil
}

// Full returns the full git sha
func (s SHA) Full() string { return s.full }

// Short returns the first 8 characters of the sha
func (s SHA) Short() string { return s.short }
