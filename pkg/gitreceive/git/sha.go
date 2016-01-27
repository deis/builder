package git

import (
	"fmt"
)

const (
	// this constant represents the length of a shortened git sha - 8 characters long
	shortShaIdx = 8
	fullShaLen  = 40
)

type ErrGitShaTooShort struct {
	sha string
}

func (e ErrGitShaTooShort) Error() string {
	return fmt.Sprintf("git sha %s was too short", e.sha)
}

type SHA struct {
	Full  string
	Short string
}

func NewSha(rawSha string) (*SHA, error) {
	if len(rawSha) < fullShaLen {
		return nil, ErrGitShaTooShort{sha: rawSha}
	}
	return &SHA{Full: rawSha, Short: rawSha[0:8]}, nil
}
