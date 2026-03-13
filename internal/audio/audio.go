package audio

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

//go:embed tashkent.mp3
var trackData []byte

//go:embed notification-pop.mp3
var successSoundData []byte

//go:embed error-001.mp3
var errorSoundData []byte

//go:embed ding.mp3
var dingSoundData []byte

// Player manages background audio playback with fade support.
type Player struct {
	volume  *effects.Volume
	done    chan struct{}
	stopped atomic.Bool
}

// NewPlayer creates a new audio player.
func NewPlayer() *Player {
	return &Player{
		done: make(chan struct{}),
	}
}

// Play starts looping the embedded track. If fadeIn > 0, the volume ramps up
// over that duration. Set SPM_DISABLE_AUDIO=1 to skip playback (useful in tests).
func (p *Player) Play(fadeIn time.Duration) error {
	if os.Getenv("SPM_DISABLE_AUDIO") == "1" {
		return nil
	}

	reader := bytes.NewReader(trackData)
	streamer, format, err := mp3.Decode(nopCloserReader{reader})
	if err != nil {
		return fmt.Errorf("decode mp3: %w", err)
	}

	var initErr error
	speakerOnce.Do(func() {
		initErr = speaker.Init(format.SampleRate, format.SampleRate.N(100*time.Millisecond))
	})
	if initErr != nil {
		return fmt.Errorf("init speaker: %w", initErr)
	}

	loop := beep.Loop(-1, streamer)

	startVolume := 0.0
	if fadeIn <= 0 {
		startVolume = 1.0
	}

	p.volume = &effects.Volume{
		Streamer: loop,
		Base:     2,
		Volume:   volumeToDb(startVolume),
		Silent:   startVolume == 0,
	}

	speaker.Play(beep.Seq(p.volume, beep.Callback(func() {
		close(p.done)
	})))

	if fadeIn > 0 {
		go p.fade(0, 1, fadeIn)
	}

	return nil
}

// Stop immediately silences playback. The underlying speaker is left open so
// that PlayNotification can reuse it afterwards.
func (p *Player) Stop() {
	if !p.stopped.CompareAndSwap(false, true) {
		return
	}
	if p.volume != nil {
		speaker.Lock()
		p.volume.Silent = true
		p.volume.Streamer = beep.Silence(-1)
		speaker.Unlock()
	}
}

// Close shuts down the shared speaker. Call once when all audio work is done.
func Close() {
	speaker.Close()
}

func (p *Player) fade(from, to float64, d time.Duration) {
	steps := 30
	stepDuration := d / time.Duration(steps)
	for i := 0; i <= steps; i++ {
		if p.stopped.Load() {
			return
		}
		level := from + (to-from)*float64(i)/float64(steps)
		speaker.Lock()
		if level <= 0.001 {
			p.volume.Silent = true
		} else {
			p.volume.Silent = false
			p.volume.Volume = volumeToDb(level)
		}
		speaker.Unlock()
		time.Sleep(stepDuration)
	}
}

// volumeToDb converts a linear volume (0.0–1.0) to a decibel scale for beep.
func volumeToDb(level float64) float64 {
	if level <= 0 {
		return -10
	}
	return math.Log2(level)
}

// nopCloserReader wraps a bytes.Reader to satisfy io.ReadCloser.
type nopCloserReader struct {
	*bytes.Reader
}

func (nopCloserReader) Close() error { return nil }

// PlayNotification decodes and plays a one-shot notification sound using beep,
// blocking until playback completes. When vibes is true and success is true,
// plays the elevator ding instead of the default pop.
func PlayNotification(success bool, vibes bool) error {
	if os.Getenv("SPM_DISABLE_AUDIO") == "1" {
		return nil
	}

	var data []byte
	switch {
	case vibes && success:
		data = dingSoundData
	case success:
		data = successSoundData
	default:
		data = errorSoundData
	}

	reader := bytes.NewReader(data)
	streamer, format, err := mp3.Decode(nopCloserReader{reader})
	if err != nil {
		return fmt.Errorf("decode notification mp3: %w", err)
	}

	// Speaker may already be initialised by the vibes Player; the once
	// guard inside Player won't help here so we use a package-level once.
	speakerOnce.Do(func() {
		err = speaker.Init(format.SampleRate, format.SampleRate.N(100*time.Millisecond))
	})
	if err != nil {
		return fmt.Errorf("init speaker: %w", err)
	}

	done := make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))
	<-done

	return nil
}

// speakerOnce ensures the speaker is initialised at most once across both
// Player (vibes) and PlayNotification.
var speakerOnce sync.Once
