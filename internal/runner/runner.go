package runner

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/decampsrenan/spm/internal/audio"
	"github.com/decampsrenan/spm/internal/ui"
)

const fadeDuration = 3 * time.Second

// RunSubprocess executes the given command as a subprocess (not process replacement).
// If dryRun is true, it prints what would be run and returns nil.
func RunSubprocess(args []string, dryRun bool) error {
	if len(args) == 0 {
		return fmt.Errorf("no command to run")
	}

	if dryRun {
		ui.Println(ui.Command(args))
		return nil
	}

	bin, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", args[0], err)
	}

	cmd := exec.Command(bin, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes the given command as a subprocess. If dryRun is true, it prints
// what would be run and returns nil. If vibes is true, it plays background
// music during execution. If notify is true, it plays a notification sound on
// completion.
//
// The command runs as a child (rather than via syscall.Exec) so we can drain
// terminal response sequences after it exits. Long-running TUI-style commands
// (e.g. turbo, vite) frequently query the terminal for capabilities; if they
// are killed mid-query the response leaks into the shell's stdin and surfaces
// as random characters at the next prompt.
func Run(args []string, dryRun bool, vibes bool, notify bool) error {
	if len(args) == 0 {
		return fmt.Errorf("no command to run")
	}

	if dryRun {
		ui.Println(ui.Command(args))
		return nil
	}

	bin, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", args[0], err)
	}

	var vibesProc *audio.VibesProcess
	if vibes {
		var err error
		vibesProc, err = audio.StartVibes(fadeDuration)
		if err != nil {
			ui.Eprintln(ui.Warning(fmt.Sprintf("could not play music: %v", err)))
		}
	}

	cmd := exec.Command(bin, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		if vibesProc != nil {
			vibesProc.StopImmediately()
		}
		return fmt.Errorf("start command: %w", err)
	}

	// Catch SIGINT/SIGTERM so the parent stays alive long enough to drain the
	// terminal after the child exits. Forward the signal to the child so it
	// terminates whether the signal arrived via the foreground process group
	// (Ctrl+C) or was addressed to spm directly.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var caughtSig atomic.Int32
	sigDone := make(chan struct{})
	go func() {
		defer close(sigDone)
		s, ok := <-sigCh
		if !ok {
			return
		}
		caughtSig.Store(int32(s.(syscall.Signal)))
		if cmd.Process != nil {
			_ = cmd.Process.Signal(s)
		}
	}()

	runErr := cmd.Wait()
	signal.Stop(sigCh)
	close(sigCh)
	<-sigDone

	// Drain any pending terminal response sequences (DECRPM, cursor pos, …)
	// the child may have left in stdin after being killed mid-query.
	ui.DrainTerminalResponses()

	if vibesProc != nil {
		if notify {
			vibesProc.StopImmediately()
		} else {
			vibesProc.FadeOutAndDetach()
		}
	}

	if notify {
		sound := audio.NotificationSound(runErr == nil, vibes)
		_ = audio.PlayNotification(sound)
	}

	switch syscall.Signal(caughtSig.Load()) {
	case syscall.SIGTERM:
		os.Exit(143)
	case syscall.SIGINT:
		os.Exit(130)
	}

	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return runErr
	}

	return nil
}
