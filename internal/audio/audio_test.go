package audio

import (
	"testing"
)

func TestTrackDataEmbedded(t *testing.T) {
	if len(trackData) == 0 {
		t.Fatal("embedded track data is empty")
	}
}

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

func TestFadeOutWithoutPlay(t *testing.T) {
	p := NewPlayer()
	// FadeOut on a player that was never started should not panic.
	p.FadeOut(0)
}

func TestDoubleStop(t *testing.T) {
	p := NewPlayer()
	// Double stop should not panic.
	p.Stop()
	p.Stop()
}

func TestVolumeToDb(t *testing.T) {
	tests := []struct {
		name  string
		level float64
	}{
		{"zero", 0},
		{"half", 0.5},
		{"full", 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := volumeToDb(tt.level)
			if tt.level == 1.0 && db != 0 {
				t.Fatalf("volumeToDb(1.0) = %f, want 0", db)
			}
			if tt.level == 0 && db >= 0 {
				t.Fatalf("volumeToDb(0) = %f, want negative", db)
			}
		})
	}
}
