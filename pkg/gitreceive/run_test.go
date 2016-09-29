package gitreceive

import (
	"testing"
)

func TestReadLine(t *testing.T) {
	//readLine() expects a line with two spaces.
	good := "foo bar car"
	foo, bar, car, err := readLine(good)
	if err != nil {
		t.Errorf("Expected err to be nil, got '%v'", err)
	}
	if foo != "foo" {
		t.Errorf("Expected 'foo', got '%s'", foo)
	}
	if bar != "bar" {
		t.Errorf("Expected 'bar', got '%s'", bar)
	}
	if car != "car" {
		t.Errorf("Expected 'car', got '%s'", car)
	}
	bad := "foo bar"
	if _, _, _, err = readLine(bad); err == nil {
		t.Error("Expected err to be not nil, got nil")
	}
}

func TestRun(t *testing.T) {
	// NOTE(bacongobbler): not much we can test at this time other than it fails based on bad setup
	if err := Run(nil, nil, nil, nil); err == nil {
		t.Errorf("expected error to be non-nil, got nil")
	}
}
