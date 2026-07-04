// dur_test.go verifies the CLI-layer helpers for the 'dur' sub-command.
// It covers the negativeDurationHint factory as registered for 'dur', which
// turns the cryptic pflag "unknown shorthand flag" error (produced when a user
// passes a negative-looking duration such as -1h) into a clear message
// directing them to -s/--sub, while leaving every other flag error untouched.
package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestNegativeDurationHint(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		rewrite  bool   // true when the error should be replaced with the hint
		contains string // substring the returned error must contain
	}{
		{name: "brief negative duration", err: errors.New("unknown shorthand flag: '1' in -1h"), rewrite: true, contains: "-s/--sub"},
		{name: "fractional negative duration", err: errors.New("unknown shorthand flag: '1' in -1.5h"), rewrite: true, contains: "-s/--sub"},
		{name: "verbose negative duration", err: errors.New("unknown shorthand flag: '9' in -90m"), rewrite: true, contains: "-s/--sub"},
		{name: "unknown letter flag", err: errors.New("unknown shorthand flag: 'x' in -x"), rewrite: false},
		{name: "unrelated error", err: errors.New("some other error"), rewrite: false},
	}

	hint := negativeDurationHint("dur", "Use -s/--sub to subtract, e.g.:\n  dtmate dur now 1h -s")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hint(nil, tt.err)
			if !tt.rewrite {
				if got != tt.err {
					t.Fatalf("expected error to pass through unchanged, got: %v", got)
				}
				return
			}
			if got == tt.err {
				t.Fatal("expected error to be rewritten, got the original")
			}
			if !strings.Contains(got.Error(), tt.contains) {
				t.Errorf("rewritten error %q does not contain %q", got.Error(), tt.contains)
			}
		})
	}
}
