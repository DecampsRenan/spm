//go:build !cgo

package audio

import "time"

// Player is a no-op stub used when CGO is disabled.
type Player struct{}

// NewPlayer creates a new no-op audio player.
func NewPlayer() *Player { return &Player{} }

// Play is a no-op when CGO is disabled.
func (p *Player) Play(fadeIn time.Duration) error { return nil }

// FadeOut is a no-op when CGO is disabled.
func (p *Player) FadeOut(d time.Duration) {}

// Stop is a no-op when CGO is disabled.
func (p *Player) Stop() {}

// CloseAudio is a no-op when CGO is disabled.
func CloseAudio() {}

// PlayNotification is a no-op when CGO is disabled.
func PlayNotification(success bool, vibes bool) error { return nil }

// RunPlaybackSubprocess is a no-op when CGO is disabled.
func RunPlaybackSubprocess() bool { return false }
