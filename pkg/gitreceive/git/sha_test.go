package git

import (
	"testing"
)

func TestShaTooShort(t *testing.T) {
	shaStrs := []string{
		"abcdefghij",
		"abc",
		"",
	}
	for i, shaStr := range shaStrs {
		sha, err := NewSha(shaStr)
		if err == nil {
			t.Errorf("expected error for sha %s (#%d), got nothing", shaStr, i)
			continue
		}
		if sha != nil {
			t.Errorf("expected returned SHA struct to be nil, got %+v instead", *sha)
		}
	}
}

func TestNewSha(t *testing.T) {
	shaStrs := []string{
		"71a09fbed590558ff822536584fc77248f070384",
		"39d8ef3ea964c98f884a67aecef56b4c8abfc700",
		"1b48115b57304d04c8c2e72d62e0a9d1bcd0d48f",
	}
	for i, shaStr := range shaStrs {
		sha, err := NewSha(shaStr)
		if err != nil {
			t.Errorf("expected no error for sha %s (#%d), got %s", shaStr, i, err)
			continue
		}
		if sha.Full != shaStr {
			t.Errorf("expected full sha string to be %s, got %s", shaStr, sha.Full)
		}
		if sha.Short != shaStr[0:8] {
			t.Errorf("expected short sha to be first 8 characters of %s, got %s", shaStr, sha.Short)
		}
	}
}
