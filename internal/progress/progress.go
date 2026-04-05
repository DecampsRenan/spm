package progress

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/decampsrenan/spm/internal/audio"
	"github.com/decampsrenan/spm/internal/ui"
)

const fadeDuration = 3 * time.Second

// Config holds options for the progress TUI.
type Config struct {
	Args   []string
	DryRun bool
	Vibes  bool
	Notify bool
	Action string // shown during progress, e.g. "Installing"
	Done   string // shown on success, e.g. "Installed"
}

// Run executes the given command with a progress TUI.
// It pipes stdout/stderr through a bubbletea model that shows a spinner,
// scrolling log lines, and elapsed time.
func Run(cfg Config) error {
	args := cfg.Args
	dryRun := cfg.DryRun
	vibes := cfg.Vibes
	notify := cfg.Notify

	action := cfg.Action
	if action == "" {
		action = "Installing"
	}
	done := cfg.Done
	if done == "" {
		done = "Installed"
	}
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

	// Start background music if requested.
	var vibesProc *audio.VibesProcess
	if vibes {
		vibesProc, err = audio.StartVibes(fadeDuration)
		if err != nil {
			ui.Eprintln(ui.Warning(fmt.Sprintf("could not play music: %v", err)))
		}
	}

	// Set up subprocess with piped output.
	cmd := exec.Command(bin, args[1:]...)
	cmd.Stdin = os.Stdin

	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("create pipe: %w", err)
	}
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Start(); err != nil {
		w.Close()
		r.Close()
		return fmt.Errorf("start command: %w", err)
	}
	w.Close() // close write end in parent — child has its own fd

	// Single channel for ordered messages to the TUI.
	msgCh := make(chan tea.Msg, 100)

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			msgCh <- outputLineMsg(scanner.Text())
		}
		r.Close()

		// Wait for process after all output has been read.
		waitErr := cmd.Wait()
		exitCode := 0
		if waitErr != nil {
			if exitErr, ok := waitErr.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		msgCh <- doneMsg{exitCode: exitCode, err: waitErr}
		close(msgCh)
	}()

	// Run TUI.
	m := newProgressModel(msgCh, action, done)
	p := tea.NewProgram(m)

	// Handle signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		s := <-sigCh
		if vibesProc != nil {
			vibesProc.StopImmediately()
		}
		_ = cmd.Process.Kill()
		p.Quit()
		ui.DrainTerminalResponses()
		if s == syscall.SIGTERM {
			os.Exit(143)
		}
		os.Exit(130)
	}()

	finalModel, err := p.Run()
	ui.DrainTerminalResponses()
	if err != nil {
		return fmt.Errorf("progress TUI error: %w", err)
	}

	// Extract exit code from final model.
	fm, ok := finalModel.(model)
	exitCode := 0
	if ok {
		exitCode = fm.exitCode
	}

	// Stop music.
	if vibesProc != nil {
		if notify {
			vibesProc.StopImmediately()
		} else {
			vibesProc.FadeOutAndDetach()
		}
	}

	// Play notification sound.
	if notify {
		sound := audio.NotificationSound(exitCode == 0, vibes)
		_ = audio.PlayNotification(sound)
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}
