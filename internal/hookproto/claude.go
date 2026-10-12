// Package hookproto defines the JSON shapes Claude Code sends to and
// expects from a PreToolUse hook command.
package hookproto

import "encoding/json"

// PreToolUseInput is the payload Claude Code pipes to a PreToolUse hook's
// stdin. ToolInput is kept raw because its shape depends on ToolName.
type PreToolUseInput struct {
	SessionID     string          `json:"session_id"`
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name"`
	ToolInput     json.RawMessage `json:"tool_input"`
}

// BashToolInput is the shape of ToolInput when ToolName == "Bash".
type BashToolInput struct {
	Command string `json:"command"`
}

// PermissionDecision values accepted by Claude Code's hookSpecificOutput.
const (
	PermissionAllow = "allow"
	PermissionDeny  = "deny"
	PermissionAsk   = "ask"
)

// PreToolUseOutput is written to stdout by a PreToolUse hook. Leaving
// HookSpecificOutput nil (encoded as `{}`) tells Claude Code the hook has
// no opinion and its normal permission flow should decide.
type PreToolUseOutput struct {
	HookSpecificOutput *PreToolUseHookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

type PreToolUseHookSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
}
