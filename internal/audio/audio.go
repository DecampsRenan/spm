package audio

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
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

// SoundName identifies a notification sound.
type SoundName string

const (
	SoundSuccess SoundName = "success"
	SoundError   SoundName = "error"
	SoundDing    SoundName = "ding"
)

// NotificationSound picks the right sound name for the given outcome.
func NotificationSound(success bool, vibes bool) SoundName {
	switch {
	case vibes && success:
		return SoundDing
	case success:
		return SoundSuccess
	default:
		return SoundError
	}
}

// PlayNotification launches a detached child process (re-execing the current
// binary with _play-sound) that plays the notification and exits. The caller
// returns immediately so the terminal is unblocked.
func PlayNotification(sound SoundName) error {
	if os.Getenv("SPM_DISABLE_AUDIO") == "1" {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	cmd := exec.Command(exe, "_play-sound", string(sound))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start notification process: %w", err)
	}

	// Detach — don't wait for the child.
	go func() { _ = cmd.Wait() }()
	return nil
}

// PlaySound plays the named embedded sound using beep, blocking until done.
// This is intended to be called from the hidden _play-sound subcommand.
func PlaySound(name string) error {
	var data []byte
	switch SoundName(name) {
	case SoundSuccess:
		data = successSoundData
	case SoundError:
		data = errorSoundData
	case SoundDing:
		data = dingSoundData
	default:
		return fmt.Errorf("unknown sound: %s", name)
	}

	reader := bytes.NewReader(data)
	streamer, format, err := mp3.Decode(nopCloserReader{reader})
	if err != nil {
		return fmt.Errorf("decode mp3: %w", err)
	}

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(100*time.Millisecond)); err != nil {
		return fmt.Errorf("init speaker: %w", err)
	}

	done := make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))
	<-done

	speaker.Close()
	return nil
}

// speakerOnce ensures the speaker is initialised at most once.
var speakerOnce sync.Once
