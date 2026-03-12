package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/decampsrenan/spm/internal/audio"
)

const fadeDuration = 3 * time.Second

// Run executes the given command. If dryRun is true, it prints what would be
// run and returns nil. If vibes is true, it plays background music during
// execution. Otherwise it replaces the current process via syscall.Exec.
func Run(args []string, dryRun bool, vibes bool) error {
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

	if !vibes {
		return syscall.Exec(bin, args, os.Environ())
	}

	// Vibes mode: run as child process so we can play music concurrently.
	player := audio.NewPlayer()
	if err := player.Play(fadeDuration); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not play music: %v\n", err)
	}
	defer player.Stop()

	cmd := exec.Command(bin, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		player.FadeOut(fadeDuration)
		player.Stop()
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	player.FadeOut(fadeDuration)
	return nil
}
