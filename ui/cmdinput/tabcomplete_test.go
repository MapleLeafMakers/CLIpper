package cmdinput

import (
	"reflect"
	"strings"
	"testing"
)

func TestTabComplete_Complete(t *testing.T) {
	tc := NewTabComplete()
	tc.completions = append(tc.completions, []Completer{NewStringCompleter("/set"), func(input string) (bool, []string) {
		matches := []string{}
		options := []string{"LogIncoming", "LogOutgoing", "MaxLines"}
		for _, o := range options {
			if strings.HasPrefix(strings.ToLower(o), strings.ToLower(input)) {
				matches = append(matches, o)
			}
		}
		return len(matches) > 0, matches
	}})

	opts := tc.Complete("/set", []string{})
	if len(opts) != 1 || opts[0] != "/set" {
		t.Errorf("Result was incorrect, got: %s, want: [/set].", opts)
	}

	expected := []string{"LogIncoming", "LogOutgoing", "MaxLines"}
	opts = tc.Complete("", []string{"/set"})
	if !reflect.DeepEqual(opts, expected) {
		t.Errorf("Result was incorrect, got: %+v, want: %+v.", opts, expected)
	}
}
