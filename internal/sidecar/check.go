// Package sidecar appends structured audit records to .guard/logs/ as a
// side effect of guard check, independent of the decision logic itself.
package sidecar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CheckLogEntry is one JSONL record in the check log: what was evaluated
// and what guard check decided.
type CheckLogEntry struct {
	Timestamp         time.Time       `json:"timestamp"`
	Mode              string          `json:"mode"` // "hook" | "cli"
	SessionID         string          `json:"session_id,omitempty"`
	RawCommand        string          `json:"raw_command"`
	Action            string          `json:"action"`
	Reason            string          `json:"reason,omitempty"`
	Next              string          `json:"next,omitempty"`
	MatchedPolicies   []string        `json:"matched_policies,omitempty"`
	Default           bool            `json:"default"`
	UsedFallbackParse bool            `json:"used_fallback_parse"`
	CustomCheck       *CustomCheckLog `json:"custom_check,omitempty"`
	Commands          []CommandLog    `json:"commands,omitempty"`
}

// CommandLog is the verdict for one simple command decomposed out of
// RawCommand (e.g. the two sides of "a && b").
type CommandLog struct {
	Tool   string `json:"tool"`
	Action string `json:"action"`
}

type CustomCheckLog struct {
	ExitCode int    `json:"exit_code"`
	TimedOut bool   `json:"timed_out"`
	Override string `json:"override,omitempty"`
}

// AppendCheckLog appends entry as one JSON line to path, creating the
// parent directory and file as needed. JSON Lines keeps this a single
// atomic O_APPEND write regardless of how many `guard check` processes
// run concurrently (e.g. parallel Bash tool calls) -- no read-modify-write
// cycle, no locking needed, and the file stays streamable/grep-able.
func AppendCheckLog(path string, entry CheckLogEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file %q: %w", path, err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}
	data = append(data, '\n')

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write log entry: %w", err)
	}
	return nil
}
