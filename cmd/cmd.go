package cmd

import (
	"fmt"
	"io"
)

// Cmd holds the dependencies every subcommand needs (standard streams),
// mirroring the dependency-injected struct+method pattern used across
// this repo's other made/ CLIs.
type Cmd struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func New(stdin io.Reader, stdout, stderr io.Writer) *Cmd {
	return &Cmd{Stdin: stdin, Stdout: stdout, Stderr: stderr}
}

// ExitError carries a specific process exit code for a legitimate,
// non-error outcome (e.g. "guard check" deciding DENY/CONFIRM), as
// opposed to a genuine failure. main.go checks for this type before
// falling back to the generic "error: ..." + exit 1 path, so a DENY
// verdict is never reported to the user as if guard itself had crashed.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}
