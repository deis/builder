package git

import (
	"fmt"
	"regexp"
)

const (
	// this constant represents the length of a shortened git sha - 8 characters long
	shortShaIdx = 8
	fullShaLen  = 40
)

var shaRegex = regexp.MustCompile(`^[\da-f]{40}$`)

type ErrInvalidGitSha struct {
	sha string
}

func (e ErrInvalidGitSha) Error() string {
	return fmt.Sprintf("git sha %s was invalid", e.sha)
}

type SHA struct {
	full  string
	short string
}

func NewSha(rawSha string) (*SHA, error) {
	if !shaRegex.MatchString(rawSha) {
		return nil, ErrInvalidGitSha{sha: rawSha}
	}
	return &SHA{full: rawSha, short: rawSha[0:8]}, nil
}

func (s SHA) Full() string  { return s.full }
func (s SHA) Short() string { return s.short }
