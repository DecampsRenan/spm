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

// Player manages background audio playback with fade support.
type Player struct {
	volume     *effects.Volume
	done       chan struct{}
	cancelFade chan struct{}
	initOnce   sync.Once
	stopped    atomic.Bool
	mu         sync.Mutex // guards cancelFade
}

// NewPlayer creates a new audio player.
func NewPlayer() *Player {
	return &Player{
		done:       make(chan struct{}),
		cancelFade: make(chan struct{}),
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
	p.initOnce.Do(func() {
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

// FadeOut cancels any in-progress fade, then gradually reduces the volume to
// zero over the given duration.
func (p *Player) FadeOut(d time.Duration) {
	if p.volume == nil || p.stopped.Load() {
		return
	}
	// Cancel any running fade-in before starting fade-out.
	p.mu.Lock()
	select {
	case <-p.cancelFade:
	default:
		close(p.cancelFade)
	}
	p.mu.Unlock()

	// Read current volume level instead of assuming 1.0.
	speaker.Lock()
	from := dbToVolume(p.volume.Volume)
	if p.volume.Silent {
		from = 0
	}
	speaker.Unlock()

	p.fade(from, 0, d)
}

// Stop immediately stops playback and releases resources.
func (p *Player) Stop() {
	if !p.stopped.CompareAndSwap(false, true) {
		return
	}
	speaker.Close()
}

func (p *Player) fade(from, to float64, d time.Duration) {
	steps := 30
	stepDuration := d / time.Duration(steps)
	for i := 0; i <= steps; i++ {
		if p.stopped.Load() {
			return
		}
		select {
		case <-p.cancelFade:
			return
		default:
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

// dbToVolume converts a decibel value back to linear volume (0.0–1.0).
func dbToVolume(db float64) float64 {
	v := math.Pow(2, db)
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// nopCloserReader wraps a bytes.Reader to satisfy io.ReadCloser.
type nopCloserReader struct {
	*bytes.Reader
}

func (nopCloserReader) Close() error { return nil }
