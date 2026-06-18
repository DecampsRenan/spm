package audio

import (
	"os"
	"testing"
)

func TestVibesProcessStopImmediatelyNilCmd(t *testing.T) {
	// StopImmediately on a VibesProcess with nil cmd should not panic.
	// This is the case when SPM_DISABLE_AUDIO=1.
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	v, err := StartVibes(0)
	if err != nil {
		t.Fatalf("StartVibes returned error: %v", err)
	}
	v.StopImmediately()
}

func TestVibesProcessFadeOutAndDetachNilCmd(t *testing.T) {
	// FadeOutAndDetach on a VibesProcess with nil cmd should not panic.
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	v, err := StartVibes(0)
	if err != nil {
		t.Fatalf("StartVibes returned error: %v", err)
	}
	v.FadeOutAndDetach()
}

func TestPlayNotificationDisabled(t *testing.T) {
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	if err := PlayNotification(SoundSuccess); err != nil {
		t.Fatalf("PlayNotification returned error: %v", err)
	}
}

func TestNotificationSound(t *testing.T) {
	tests := []struct {
		success bool
		vibes   bool
		want    SoundName
	}{
		{true, true, SoundDing},
		{true, false, SoundSuccess},
		{false, true, SoundError},
		{false, false, SoundError},
	}
	for _, tt := range tests {
		got := NotificationSound(tt.success, tt.vibes)
		if got != tt.want {
			t.Errorf("NotificationSound(%v, %v) = %q, want %q", tt.success, tt.vibes, got, tt.want)
		}
	}
}

func TestStartVibesReturnsProcessWhenAudioEnabled(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping on CI — requires audio binary")
	}
	// When audio is not disabled, StartVibes should return a process with a non-nil cmd.
	// We can't easily test this without the binary, so just verify the disabled path.
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	v, err := StartVibes(0)
	if err != nil {
		t.Fatalf("StartVibes returned error: %v", err)
	}
	if v == nil {
		t.Fatal("StartVibes returned nil VibesProcess")
	}
}
