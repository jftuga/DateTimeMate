// durmath_test.go verifies the CLI-layer helpers for the 'durmath'
// sub-command. It covers the negativeDurationHint factory as registered for
// 'durmath', ensuring pflag digit-shorthand errors (produced when a user
// passes a negative-looking duration such as -30m) are rewritten with the
// durmath wording, while every other flag error passes through untouched.
package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestDurMathNegativeDurationHint(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		rewrite  bool   // true when the error should be replaced with the hint
		contains string // substring the returned error must contain
	}{
		{name: "brief negative duration", err: errors.New("unknown shorthand flag: '3' in -30m"), rewrite: true, contains: "-a/--add or -s/--sub"},
		{name: "fractional negative duration", err: errors.New("unknown shorthand flag: '1' in -1.5h"), rewrite: true, contains: "'durmath'"},
		{name: "unknown letter flag", err: errors.New("unknown shorthand flag: 'x' in -x"), rewrite: false},
		{name: "unrelated error", err: errors.New("some other error"), rewrite: false},
	}

	hint := negativeDurationHint("durmath", "Use -a/--add or -s/--sub to control the operation, e.g.:\n  dtmate durmath 2h 30m -s")
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
