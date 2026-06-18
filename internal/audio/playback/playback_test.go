//go:build darwin || cgo

package playback

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

func TestTrackDataEmbedded(t *testing.T) {
	if len(trackData) == 0 {
		t.Fatal("embedded track data is empty")
	}
}

func TestSuccessSoundDataEmbedded(t *testing.T) {
	if len(successSoundData) == 0 {
		t.Fatal("embedded success sound data is empty")
	}
}

func TestErrorSoundDataEmbedded(t *testing.T) {
	if len(errorSoundData) == 0 {
		t.Fatal("embedded error sound data is empty")
	}
}

func TestDingSoundDataEmbedded(t *testing.T) {
	if len(dingSoundData) == 0 {
		t.Fatal("embedded ding sound data is empty")
	}
}
