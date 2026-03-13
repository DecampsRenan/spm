//go:build darwin || cgo

package audio

import (
	"testing"
)

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
