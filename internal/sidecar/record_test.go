package sidecar

import "testing"

func TestBuildRecordLogEntry_Bash(t *testing.T) {
	raw := []byte(`{
		"session_id": "sess-1",
		"hook_event_name": "PostToolUse",
		"tool_name": "Bash",
		"tool_use_id": "toolu_01",
		"duration_ms": 42,
		"tool_input": {"command": "echo hi"},
		"tool_response": {"stdout": "hi\n", "stderr": "", "interrupted": false}
	}`)

	entry, ok := BuildRecordLogEntry(raw)
	if !ok {
		t.Fatalf("expected ok=true for a Bash payload")
	}
	if entry.Command != "echo hi" {
		t.Errorf("Command = %q, want %q", entry.Command, "echo hi")
	}
	if entry.Stdout != "hi\n" {
		t.Errorf("Stdout = %q, want %q", entry.Stdout, "hi\n")
	}
	if entry.Interrupted {
		t.Errorf("Interrupted = true, want false")
	}
	if entry.SessionID != "sess-1" || entry.ToolUseID != "toolu_01" {
		t.Errorf("SessionID/ToolUseID = %q/%q, want sess-1/toolu_01", entry.SessionID, entry.ToolUseID)
	}
	if entry.DurationMs == nil || *entry.DurationMs != 42 {
		t.Errorf("DurationMs = %v, want 42", entry.DurationMs)
	}
}

func TestBuildRecordLogEntry_StdoutIsVerbatimNotSummarized(t *testing.T) {
	long := ""
	for i := 0; i < 10000; i++ {
		long += "x"
	}
	raw := []byte(`{"tool_name":"Bash","tool_input":{"command":"yes x | head -c 10000"},"tool_response":{"stdout":"` + long + `"}}`)

	entry, ok := BuildRecordLogEntry(raw)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if len(entry.Stdout) != 10000 {
		t.Fatalf("expected Stdout to be kept verbatim (10000 chars), got %d chars", len(entry.Stdout))
	}
}

func TestBuildRecordLogEntry_NonBashIsIgnored(t *testing.T) {
	raw := []byte(`{"tool_name":"Read","tool_input":{"file_path":"/x"},"tool_response":{"content":"..."}}`)

	_, ok := BuildRecordLogEntry(raw)
	if ok {
		t.Fatalf("expected ok=false for a non-Bash tool")
	}
}

func TestBuildRecordLogEntry_MalformedPayloadIsIgnored(t *testing.T) {
	for _, raw := range [][]byte{
		[]byte(`not json`),
		[]byte(`{}`),
		[]byte(`{"tool_input":{"command":"x"}}`), // no tool_name
	} {
		if _, ok := BuildRecordLogEntry(raw); ok {
			t.Fatalf("expected ok=false for malformed payload %q", raw)
		}
	}
}

func TestBuildRecordLogEntry_InterruptedIsPropagated(t *testing.T) {
	raw := []byte(`{"tool_name":"Bash","tool_input":{"command":"sleep 100"},"tool_response":{"interrupted":true}}`)

	entry, ok := BuildRecordLogEntry(raw)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if !entry.Interrupted {
		t.Fatalf("expected Interrupted=true")
	}
}
