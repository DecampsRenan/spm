package audio

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
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
	volume   *effects.Volume
	done     chan struct{}
	initOnce sync.Once
	stopped  atomic.Bool
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

// PlayNotification plays a one-shot notification sound (fire-and-forget)
// via an OS-native audio command so the main process can exit immediately.
// When vibes is true and success is true, plays the elevator ding instead.
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

	tmpFile, err := writeTempMP3(data)
	if err != nil {
		return fmt.Errorf("write notification mp3: %w", err)
	}

	cmd := osPlayerCmd(tmpFile)
	if err := cmd.Start(); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("play notification: %w", err)
	}

	// Clean up temp file after playback completes in background.
	go func() {
		_ = cmd.Wait()
		_ = os.Remove(tmpFile)
	}()

	return nil
}

func writeTempMP3(data []byte) (string, error) {
	f, err := os.CreateTemp("", "spm-*.mp3")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func osPlayerCmd(file string) *exec.Cmd {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("afplay", file)
	default:
		// Linux: try paplay first, fall back to aplay.
		if path, err := exec.LookPath("paplay"); err == nil {
			return exec.Command(path, file)
		}
		if path, err := exec.LookPath("aplay"); err == nil {
			return exec.Command(path, file)
		}
		return exec.Command("ffplay", "-nodisp", "-autoexit", file)
	}
}
