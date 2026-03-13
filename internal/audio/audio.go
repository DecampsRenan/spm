package audio

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// VibesProcess represents a detached child process playing background music.
type VibesProcess struct {
	cmd *exec.Cmd
}

// StartVibes launches a detached child process that plays background music.
func StartVibes(fadeIn time.Duration) (*VibesProcess, error) {
	if os.Getenv("SPM_DISABLE_AUDIO") == "1" {
		return &VibesProcess{}, nil
	}

	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable: %w", err)
	}

	cmd := exec.Command(exe, "_play-music", fmt.Sprintf("%d", int(fadeIn.Seconds())))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start vibes process: %w", err)
	}

	return &VibesProcess{cmd: cmd}, nil
}

// StopImmediately kills the vibes process without fade-out.
func (v *VibesProcess) StopImmediately() {
	if v.cmd == nil || v.cmd.Process == nil {
		return
	}
	_ = v.cmd.Process.Kill()
	_ = v.cmd.Wait()
}

// FadeOutAndDetach sends SIGTERM to the vibes process (triggering fade-out)
// and detaches. The child process fades out and exits on its own.
func (v *VibesProcess) FadeOutAndDetach() {
	if v.cmd == nil || v.cmd.Process == nil {
		return
	}
	_ = v.cmd.Process.Signal(syscall.SIGTERM)
	go func() { _ = v.cmd.Wait() }()
}

// SoundName identifies a notification sound.
type SoundName string

const (
	SoundSuccess SoundName = "success"
	SoundError   SoundName = "error"
	SoundDing    SoundName = "ding"
)

// NotificationSound picks the right sound name for the given outcome.
func NotificationSound(success bool, vibes bool) SoundName {
	switch {
	case vibes && success:
		return SoundDing
	case success:
		return SoundSuccess
	default:
		return SoundError
	}
}

// PlayNotification launches a detached child process (re-execing the current
// binary with _play-sound) that plays the notification and exits. The caller
// returns immediately so the terminal is unblocked.
func PlayNotification(sound SoundName) error {
	if os.Getenv("SPM_DISABLE_AUDIO") == "1" {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	cmd := exec.Command(exe, "_play-sound", string(sound))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start notification process: %w", err)
	}

	// Detach — don't wait for the child.
	go func() { _ = cmd.Wait() }()
	return nil
}
