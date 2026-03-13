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
// execution. If notify is true, it plays a notification sound on completion.
// Otherwise it replaces the current process via syscall.Exec.
func Run(args []string, dryRun bool, vibes bool, notify bool) error {
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

	// Direct exec when no audio features needed.
	if !vibes && !notify {
		return syscall.Exec(bin, args, os.Environ())
	}

	// Child process mode: run command as subprocess so we can manage audio.
	var vibesProc *audio.VibesProcess
	if vibes {
		var err error
		vibesProc, err = audio.StartVibes(fadeDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not play music: %v\n", err)
		}
	}

	cmd := exec.Command(bin, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runErr := cmd.Run()

	// Signal vibes to fade out (detached — won't block us).
	if vibesProc != nil {
		if notify {
			// When notify is set, kill music immediately so the
			// notification sound is heard cleanly.
			vibesProc.StopImmediately()
		} else {
			vibesProc.FadeOutAndDetach()
		}
	}

	if notify {
		sound := audio.NotificationSound(runErr == nil, vibes)
		_ = audio.PlayNotification(sound)
	}

	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return runErr
	}

	return nil
}
