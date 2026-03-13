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
const errorFadeDuration = 500 * time.Millisecond

// Run executes the given command. If dryRun is true, it prints what would be
// run and returns nil. If vibes is true, it plays background music during
// execution. If notify is true, it plays a short sound when the command
// finishes (different for success and error). Otherwise it replaces the
// current process via syscall.Exec.
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

	if !vibes && !notify {
		return syscall.Exec(bin, args, os.Environ())
	}

	// Child process mode: needed for vibes (concurrent music) and/or notify
	// (post-command sound). Audio device is closed once at the end.
	var player *audio.Player
	if vibes {
		player = audio.NewPlayer()
		if err := player.Play(fadeDuration); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not play music: %v\n", err)
		}
	}

	cmd := exec.Command(bin, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdErr := cmd.Run()

	// Stop vibes: immediate stop when notify follows, fade-out otherwise.
	if player != nil {
		if notify {
			player.Stop()
		} else if cmdErr != nil {
			player.FadeOut(errorFadeDuration)
			player.Stop()
		} else {
			player.FadeOut(fadeDuration)
			player.Stop()
		}
		audio.CloseAudio()
	}

	// Fire-and-forget: launches a background system process to play the sound.
	if notify {
		if err := audio.PlayNotification(cmdErr == nil, vibes); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not play notification: %v\n", err)
		}
	}

	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return cmdErr
	}

	return nil
}
