package audio

import (
	"testing"
)

func TestNewPlayer(t *testing.T) {
	p := NewPlayer()
	if p == nil {
		t.Fatal("NewPlayer returned nil")
	}
}

func TestStopWithoutPlay(t *testing.T) {
	p := NewPlayer()
	// Stop on a player that was never started should not panic.
	p.Stop()
}

func TestDoubleStop(t *testing.T) {
	p := NewPlayer()
	// Double stop should not panic.
	p.Stop()
	p.Stop()
}
