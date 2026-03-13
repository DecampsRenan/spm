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

// speakerOnce ensures the speaker is initialised at most once. The beep library
// does not allow re-initialisation after Close, so we keep the speaker alive
// and only close it when the process is done with all audio.
var speakerOnce sync.Once
var speakerInitErr error

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

	if err := initSpeaker(format.SampleRate); err != nil {
		return fmt.Errorf("init speaker: %w", err)
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

// FadeOut gradually reduces the volume to zero over the given duration.
func (p *Player) FadeOut(d time.Duration) {
	if p.volume == nil || p.stopped.Load() {
		return
	}
	p.fade(1, 0, d)
}

// Stop immediately mutes playback and marks the player as stopped.
// It does NOT close the speaker — call CloseAudio() when all audio is done.
func (p *Player) Stop() {
	if !p.stopped.CompareAndSwap(false, true) {
		return
	}
	if p.volume != nil {
		speaker.Lock()
		p.volume.Silent = true
		speaker.Unlock()
	}
}

// CloseAudio tears down the audio device. Call this once when the process is
// finished with all audio playback (vibes + notification).
func CloseAudio() {
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

// initSpeaker initialises the speaker exactly once for the process lifetime.
func initSpeaker(rate beep.SampleRate) error {
	speakerOnce.Do(func() {
		speakerInitErr = speaker.Init(rate, rate.N(100*time.Millisecond))
	})
	return speakerInitErr
}

const playSoundEnv = "_SPM_PLAY_SOUND"

// PlayNotification spawns the current binary as a detached subprocess that
// plays the notification sound via beep. The caller returns immediately.
// When vibes is true and success is true, the elevator "ding" is played.
// Set SPM_DISABLE_AUDIO=1 to skip playback.
func PlayNotification(success bool, vibes bool) error {
	if os.Getenv("SPM_DISABLE_AUDIO") == "1" {
		return nil
	}

	data := errorSoundData
	if success && vibes {
		data = dingSoundData
	} else if success {
		data = successSoundData
	}

	// Write sound to a temp file so the subprocess can read it.
	f, err := os.CreateTemp("", "spm-notify-*.mp3")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	f.Close()

	// Re-launch ourselves with the special env var.
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	cmd := exec.Command(self)
	cmd.Env = append(os.Environ(), playSoundEnv+"="+f.Name())
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start playback subprocess: %w", err)
	}

	return nil
}

// RunPlaybackSubprocess checks if the process was launched as a sound-playing
// subprocess. If so, it plays the sound, cleans up, and returns true.
// Call this early in main() before any CLI parsing.
func RunPlaybackSubprocess() bool {
	soundFile := os.Getenv(playSoundEnv)
	if soundFile == "" {
		return false
	}
	defer os.Remove(soundFile)

	data, err := os.ReadFile(soundFile)
	if err != nil {
		return true
	}

	reader := bytes.NewReader(data)
	streamer, format, err := mp3.Decode(nopCloserReader{reader})
	if err != nil {
		return true
	}

	if err := initSpeaker(format.SampleRate); err != nil {
		return true
	}

	done := make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))
	<-done
	speaker.Close()
	return true
}

// nopCloserReader wraps a bytes.Reader to satisfy io.ReadCloser.
type nopCloserReader struct {
	*bytes.Reader
}

func (nopCloserReader) Close() error { return nil }
