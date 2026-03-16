package runner

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/decampsrenan/spm/internal/audio"
	"github.com/mattn/go-isatty"
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

	cmd := exec.Command(bin, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Place the child in its own process group so the terminal SIGINT does
	// not reach it directly — we forward the signal ourselves.
	if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Foreground: true,
			Ctty:       int(os.Stdin.Fd()),
		}
	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
	}

	// Intercept SIGINT/SIGTERM so we can forward them to the child's
	// process group and stop background tasks.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		sig := <-sigCh
		if vibesProc != nil {
			vibesProc.StopImmediately()
		}
		if cmd.Process != nil {
			// Forward to the child's entire process group.
			_ = syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
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
