package markovianomatic_test

import (
	"testing"

	"github.com/zeroed/markovianomatic"
)

var aPrefix = markovianomatic.Prefix{"foo", "bar"}

func TestPrefixString(t *testing.T) {
	desired := "foo bar"

	if tos := aPrefix.String(); tos != desired {
		t.Errorf("Prefix string is expected to be %s but we have %s", desired, tos)
	}
}

func TestPrefixShift(t *testing.T) {
	desired := "bar foo"
	aPrefix.Shift("foo")

	if tos := aPrefix.String(); tos != desired {
		t.Errorf("Prefix string is expected to be %s but we have %s", desired, tos)
	}
}
