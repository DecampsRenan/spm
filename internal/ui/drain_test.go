package ui

import (
	"testing"
)

func TestDrainTerminalResponsesDoesNotPanic(t *testing.T) {
	// DrainTerminalResponses opens /dev/tty which may not be available in CI.
	// This test verifies the function handles errors gracefully and never panics.
	DrainTerminalResponses()
}
