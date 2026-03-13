//go:build !darwin && !cgo

package audio

import (
	"fmt"
	"time"
)

// Player is a no-op stub used when audio playback is not supported.
type Player struct{}

// NewPlayer creates a new no-op audio player.
func NewPlayer() *Player { return &Player{} }

// Play is a no-op on unsupported platforms.
func (p *Player) Play(fadeIn time.Duration) error { return nil }

// FadeOut is a no-op on unsupported platforms.
func (p *Player) FadeOut(d time.Duration) {}

// Stop is a no-op on unsupported platforms.
func (p *Player) Stop() {}

// PlayMusicAndWait is a no-op on unsupported platforms.
func PlayMusicAndWait(fadeIn time.Duration) error { return nil }

// PlaySound is a no-op on unsupported platforms.
func PlaySound(name string) error {
	switch SoundName(name) {
	case SoundSuccess, SoundError, SoundDing:
		return nil
	default:
		return fmt.Errorf("unknown sound: %s", name)
	}
}
