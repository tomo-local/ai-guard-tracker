package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"ai-guard-tracker/cmd"
)

func main() {
	args := os.Args[1:]

	var command string
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	c := cmd.New(os.Stdin, os.Stdout, os.Stderr)

	var err error
	switch command {
	case "check":
		err = c.RunCheck(args)
	case "lint":
		err = c.RunLint(args)
	case "", "-h", "--help", "help":
		showHelp()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\nRun 'guard --help' for usage.\n", command)
		os.Exit(1)
	}

	if err != nil {
		var exitErr *cmd.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		if errors.Is(err, flag.ErrHelp) {
			// Usage was already printed by the subcommand's FlagSet.
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Fprint(os.Stderr, `guard: AI agent command guardrails

Usage: guard <command> [options]

Commands:
  check   Evaluate a Bash command against .guard/policies (CLI arg or PreToolUse hook on stdin)
  lint    Validate .guard/policies manifests
  help    Show this help message
  -h, --help   Show this help message

Run 'guard <command> -h' for details and options of each command.

Examples:
  guard check "git push --force"
  echo '{"tool_name":"Bash","tool_input":{"command":"rm -rf /"}}' | guard check
  guard lint
`)
}
