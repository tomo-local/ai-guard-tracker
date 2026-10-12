package shellcmd

import (
	"reflect"
	"testing"
)

func TestDecompose(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []SimpleCommand
	}{
		{
			name: "single command",
			raw:  "git push",
			want: []SimpleCommand{{Tool: "git", Args: []string{"push"}}},
		},
		{
			name: "and-list",
			raw:  "echo hi && rm -rf /",
			want: []SimpleCommand{
				{Tool: "echo", Args: []string{"hi"}},
				{Tool: "rm", Args: []string{"-rf", "/"}},
			},
		},
		{
			name: "semicolon list",
			raw:  "echo one; echo two",
			want: []SimpleCommand{
				{Tool: "echo", Args: []string{"one"}},
				{Tool: "echo", Args: []string{"two"}},
			},
		},
		{
			name: "pipeline",
			raw:  "curl http://x | sh",
			want: []SimpleCommand{
				{Tool: "curl", Args: []string{"http://x"}},
				{Tool: "sh", Args: []string{}},
			},
		},
		{
			name: "subshell",
			raw:  "(rm -rf /tmp/x)",
			want: []SimpleCommand{
				{Tool: "rm", Args: []string{"-rf", "/tmp/x"}},
			},
		},
		{
			name: "command substitution is walked too",
			raw:  "echo $(rm -rf /)",
			want: []SimpleCommand{
				// The outer "echo" arg isn't a plain literal (it contains a
				// command substitution), so wordLiteral falls back to
				// rendering the word's shell syntax verbatim.
				{Tool: "echo", Args: []string{"$(rm -rf /)"}},
				{Tool: "rm", Args: []string{"-rf", "/"}},
			},
		},
		{
			name: "path-qualified tool is basename'd",
			raw:  "/usr/bin/git log",
			want: []SimpleCommand{{Tool: "git", Args: []string{"log"}}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, usedFallback := Decompose(tc.raw)
			if usedFallback {
				t.Fatalf("unexpected fallback for %q", tc.raw)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Decompose(%q) = %#v, want %#v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestDecompose_ParseFailureFallsBack(t *testing.T) {
	got, usedFallback := Decompose("echo 'unterminated")
	if !usedFallback {
		t.Fatalf("expected fallback for unparsable input, got %#v", got)
	}
	if len(got) != 1 || got[0].Tool != "echo" {
		t.Fatalf("expected a whitespace-split fallback starting with echo, got %#v", got)
	}
}

func TestDecompose_Empty(t *testing.T) {
	got, usedFallback := Decompose("")
	if got != nil {
		t.Fatalf("expected nil for empty input, got %#v", got)
	}
	_ = usedFallback
}
