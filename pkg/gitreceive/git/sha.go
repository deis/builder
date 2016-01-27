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
	Full  string
	Short string
}

func NewSha(rawSha string) (*SHA, error) {
	if !shaRegex.Match([]byte(rawSha)) {
		return nil, ErrInvalidGitSha{sha: rawSha}
	}
	return &SHA{Full: rawSha, Short: rawSha[0:8]}, nil
}
