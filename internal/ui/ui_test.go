package ui

import (
	"strings"
	"testing"
)

func TestSuccessContainsMessage(t *testing.T) {
	result := Success("file removed")
	if !strings.Contains(result, "file removed") {
		t.Errorf("Success() should contain the message, got: %s", result)
	}
}

func TestErrorContainsMessage(t *testing.T) {
	result := Error("something broke")
	if !strings.Contains(result, "something broke") {
		t.Errorf("Error() should contain the message, got: %s", result)
	}
}

func TestWarningContainsMessage(t *testing.T) {
	result := Warning("watch out")
	if !strings.Contains(result, "watch out") {
		t.Errorf("Warning() should contain the message, got: %s", result)
	}
}

func TestInfoContainsMessage(t *testing.T) {
	result := Info("some info")
	if !strings.Contains(result, "some info") {
		t.Errorf("Info() should contain the message, got: %s", result)
	}
}

func TestDimContainsMessage(t *testing.T) {
	result := Dim("secondary text")
	if !strings.Contains(result, "secondary text") {
		t.Errorf("Dim() should contain the message, got: %s", result)
	}
}

func TestCommandContainsArgs(t *testing.T) {
	result := Command([]string{"npm", "install", "react"})
	if !strings.Contains(result, "npm install react") {
		t.Errorf("Command() should contain joined args, got: %s", result)
	}
	if !strings.Contains(result, "Would run") {
		t.Errorf("Command() should contain 'Would run' label, got: %s", result)
	}
}

func TestHeaderContainsMessage(t *testing.T) {
	result := Header("title")
	if !strings.Contains(result, "title") {
		t.Errorf("Header() should contain the message, got: %s", result)
	}
}

func TestPathContainsPath(t *testing.T) {
	result := Path("/foo/bar/node_modules")
	if !strings.Contains(result, "/foo/bar/node_modules") {
		t.Errorf("Path() should contain the path, got: %s", result)
	}
}
