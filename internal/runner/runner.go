package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Run executes the given command. If dryRun is true, it prints what would be
// run and returns nil. Otherwise it replaces the current process via syscall.Exec.
func Run(args []string, dryRun bool) error {
	if len(args) == 0 {
		return fmt.Errorf("no command to run")
	}

	if dryRun {
		fmt.Printf("Would run: %s\n", strings.Join(args, " "))
		return nil
	}

	bin, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", args[0], err)
	}

	return syscall.Exec(bin, args, os.Environ())
}
