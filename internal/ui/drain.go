package ui

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// DrainTerminalResponses reads and discards any pending terminal response
// sequences from the TTY input buffer. This prevents escape sequences like
// the DECRPM response (e.g., ^[[?2026;2$y for synchronized output) from
// leaking into the shell after a bubbletea/huh program exits quickly.
func DrainTerminalResponses() {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return
	}
	defer tty.Close()

	fd := int(tty.Fd())
	if err := unix.SetNonblock(fd, true); err != nil {
		return
	}

	// Brief pause to let in-flight terminal responses arrive.
	time.Sleep(10 * time.Millisecond)

	// Drain any pending data.
	buf := make([]byte, 256)
	for {
		_, err := tty.Read(buf)
		if err != nil {
			break
		}
	}
}
