package cmd

import (
	"strings"
	"testing"
)

func TestGetInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		start   string
		end     string
		wantErr bool
	}{
		{name: "comma pair", input: "15:16:15,15:17\n", start: "15:16:15", end: "15:17"},
		{name: "comma pair no newline", input: "15:16:15,15:17", start: "15:16:15", end: "15:17"},
		{name: "comma pair with spaces", input: "2024-01-01, 2024-01-02\n", start: "2024-01-01", end: "2024-01-02"},
		{name: "two lines", input: "15:16:15\n15:17:20", start: "15:16:15", end: "15:17:20"},
		{name: "two lines trailing newline", input: "15:16:15\n15:17:20\n", start: "15:16:15", end: "15:17:20"},
		{name: "two lines with comma dates", input: "Jan 2, 2024\nJan 5, 2024\n", start: "Jan 2, 2024", end: "Jan 5, 2024"},
		{name: "comma line then blank line", input: "15:16:15,15:17\n\n", start: "15:16:15", end: "15:17"},
		{name: "empty input", input: "", wantErr: true},
		{name: "blank line only", input: "\n", wantErr: true},
		{name: "one line no comma", input: "2024-01-01\n", wantErr: true},
		{name: "too many commas", input: "Jan 2, 2024,Jan 5, 2024\n", wantErr: true},
		{name: "empty start", input: ",15:17\n", wantErr: true},
		{name: "empty end", input: "15:16:15,\n", wantErr: true},
		{name: "blank first line with second line", input: "\n15:17:20\n", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := getInput(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("getInput(%q) expected error, got start=%q end=%q", tt.input, start, end)
				}
				return
			}
			if err != nil {
				t.Fatalf("getInput(%q) unexpected error: %v", tt.input, err)
			}
			if start != tt.start || end != tt.end {
				t.Errorf("getInput(%q) = (%q, %q), want (%q, %q)", tt.input, start, end, tt.start, tt.end)
			}
		})
	}
}
