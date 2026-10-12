package cmd

import (
	"flag"
	"fmt"

	"ai-guard-tracker/internal/setup"
)

const initUsage = `Usage:
  guard init [-rules <dir>] [-log <path>]

Options:
  -rules string   policy directory to set up (default ".guard/policies", env GUARD_POLICIES_DIR)
  -log string     check decision log path, used only to derive the logs
                   directory to create (default ".guard/logs/check.jsonl", env GUARD_LOG_PATH)

Precedence for both: -flag > env var > built-in default.

Creates the logs directory, and seeds the policies directory with this
CLI's own starter policies -- but only if it has no policy manifests
yet; an existing policies directory is left untouched. Never modifies
.claude/settings.json -- prints the hook snippet to add there instead.
`

// RunInit sets up .guard/policies (seeded with starter policies if empty)
// and the logs directory, then prints the Claude Code hook snippet to
// add manually. The actual bootstrap logic lives in internal/setup;
// this is just flag parsing and reporting the result.
func (c *Cmd) RunInit(args []string) error {
	flagSet := flag.NewFlagSet("init", flag.ContinueOnError)
	flagSet.SetOutput(c.Stderr)
	rulesDir := flagSet.String("rules", defaultPoliciesDir(), "policy directory to set up (env GUARD_POLICIES_DIR)")
	logPath := flagSet.String("log", defaultLogPath(), "check decision log path, used to derive the logs directory (env GUARD_LOG_PATH)")
	flagSet.Usage = func() { fmt.Fprint(c.Stderr, initUsage) }
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	res, err := setup.Run(*rulesDir, *logPath, c.StarterPolicies)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.Stdout, "logs directory ready: %s\n", res.LogsDir)
	if res.ExistingPolicyCount > 0 {
		fmt.Fprintf(c.Stdout, "policies directory already has %d polic%s, left untouched: %s\n", res.ExistingPolicyCount, plural(res.ExistingPolicyCount), res.PoliciesDir)
	} else {
		fmt.Fprintf(c.Stdout, "policies directory seeded with %d starter polic%s: %s\n", res.SeededPolicyCount, plural(res.SeededPolicyCount), res.PoliciesDir)
	}

	fmt.Fprint(c.Stdout, setup.HookSettingsSnippet)
	return nil
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
