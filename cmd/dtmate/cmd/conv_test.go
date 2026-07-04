package cmd

import (
	"slices"
	"testing"
)

func TestParseConvArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		positional []string
		brief      bool
		noNewline  bool
		help       bool
		wantErr    bool
	}{
		{name: "no flags", args: []string{"90m", "h"}, positional: []string{"90m", "h"}},
		{name: "brief before", args: []string{"-b", "90m", "h"}, positional: []string{"90m", "h"}, brief: true},
		{name: "brief after", args: []string{"90m", "h", "-b"}, positional: []string{"90m", "h"}, brief: true},
		{name: "brief long", args: []string{"--brief", "90m", "h"}, positional: []string{"90m", "h"}, brief: true},
		{name: "nonewline", args: []string{"-n", "90m", "h"}, positional: []string{"90m", "h"}, noNewline: true},
		{name: "nonewline long", args: []string{"--nonewline", "90m", "h"}, positional: []string{"90m", "h"}, noNewline: true},
		{name: "combined shorthand", args: []string{"-bn", "90m", "h"}, positional: []string{"90m", "h"}, brief: true, noNewline: true},
		{name: "separate shorthands", args: []string{"-n", "-b", "90m", "h"}, positional: []string{"90m", "h"}, brief: true, noNewline: true},
		{name: "help short", args: []string{"-h"}, help: true},
		{name: "help long", args: []string{"--help"}, help: true},
		{name: "negative brief duration", args: []string{"-90m", "h"}, positional: []string{"-90m", "h"}},
		{name: "negative verbose duration", args: []string{"-1 hour 30 minutes", "m"}, positional: []string{"-1 hour 30 minutes", "m"}},
		{name: "negative duration with flag", args: []string{"-b", "-90m", "h"}, positional: []string{"-90m", "h"}, brief: true},
		{name: "double dash terminator", args: []string{"-b", "--", "-90m", "h"}, positional: []string{"-90m", "h"}, brief: true},
		{name: "unknown shorthand", args: []string{"-x", "90m", "h"}, wantErr: true},
		{name: "unknown shorthand in cluster", args: []string{"-bx", "90m", "h"}, wantErr: true},
		{name: "unknown long flag", args: []string{"--bogus", "90m", "h"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			positional, brief, noNewline, help, err := parseConvArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(positional, tt.positional) {
				t.Errorf("positional: [computed: %v] != [correct: %v]", positional, tt.positional)
			}
			if brief != tt.brief {
				t.Errorf("brief: [computed: %v] != [correct: %v]", brief, tt.brief)
			}
			if noNewline != tt.noNewline {
				t.Errorf("noNewline: [computed: %v] != [correct: %v]", noNewline, tt.noNewline)
			}
			if help != tt.help {
				t.Errorf("help: [computed: %v] != [correct: %v]", help, tt.help)
			}
		})
	}
}
