package runner

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/decampsrenan/spm/internal/audio"
)

const (
	fadeDuration    = 3 * time.Second
	killGracePeriod = 5 * time.Second
)

// isYarn reports whether the command binary is yarn.
func isYarn(binName string) bool {
	return filepath.Base(binName) == "yarn"
}

// Run executes the given command. If dryRun is true, it prints what would be
// run and returns nil. If vibes is true, it plays background music during
// execution. If notify is true, it plays a notification sound on completion.
// When the package manager is yarn, or audio features are enabled, the command
// runs as a subprocess with proper signal forwarding. Otherwise it replaces
// the current process via syscall.Exec.
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

	// Yarn mishandles SIGINT by sending SIGKILL to children instead of
	// forwarding the signal, so we run it as a subprocess with process-group
	// isolation and manual signal forwarding. Audio features also require
	// subprocess mode.
	needsSubprocess := vibes || notify || isYarn(args[0])
	if !needsSubprocess {
		return syscall.Exec(bin, args, os.Environ())
	}

	// Child process mode: run command as subprocess.
	var vibesProc *audio.VibesProcess
	if vibes {
		var err error
		vibesProc, err = audio.StartVibes(fadeDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not play music: %v\n", err)
		}
	}

	yarn := isYarn(args[0])

	cmd := exec.Command(bin, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if yarn {
		// Yarn must NOT be in the foreground process group: when the
		// terminal delivers SIGINT to the foreground group, yarn's
		// broken handler sends SIGKILL to its children (e.g. Cypress)
		// causing crash dialogs. Setpgid isolates yarn so only spm
		// receives the terminal SIGINT; we then forward SIGTERM.
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	// Intercept SIGINT so we can stop background tasks / forward to yarn.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	var interrupted atomic.Bool

	go func() {
		<-sigCh
		interrupted.Store(true)
		if vibesProc != nil {
			vibesProc.StopImmediately()
		}
		if yarn && cmd.Process != nil {
			// Send SIGTERM (not SIGINT) to yarn's process group so
			// children can shut down gracefully without yarn
			// SIGKILL-ing them.
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
			// Safety net: force-kill after grace period if still alive.
			go func() {
				time.Sleep(killGracePeriod)
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}()
		} else {
			os.Exit(130)
		}
	}()

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
			// If the user pressed Ctrl+C, always report 130 regardless
			// of what signal we used to stop the child.
			if interrupted.Load() {
				os.Exit(130)
			}
			code := exitErr.ExitCode()
			// When a process is killed by a signal, Go returns -1.
			// Compute the conventional 128+signal exit code instead.
			if code < 0 {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
					code = 128 + int(status.Signal())
				}
			}
			os.Exit(code)
		}
		return runErr
	}

	return nil
}
