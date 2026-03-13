package main

import (
	"github.com/decampsrenan/spm/cmd"
	"github.com/decampsrenan/spm/internal/audio"
)

var version = "dev"

func main() {
	// When launched as a sound-playing subprocess, play and exit.
	if audio.RunPlaybackSubprocess() {
		return
	}

	cmd.SetVersion(version)
	cmd.Execute()
}
