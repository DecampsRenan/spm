package progress

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestViewShowsActionWhileRunning(t *testing.T) {
	tests := []struct {
		name   string
		action string
		done   string
	}{
		{"install", "Installing", "Installed"},
		{"add", "Adding", "Added"},
		{"remove", "Removing", "Removed"},
		{"init", "Initializing", "Initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan tea.Msg)
			m := newProgressModel(ch, tt.action, tt.done)
			view := m.View()
			if !strings.Contains(view.Content, tt.action+"...") {
				t.Errorf("running view should contain %q, got: %s", tt.action+"...", view.Content)
			}
		})
	}
}

func TestViewShowsDoneLabelOnSuccess(t *testing.T) {
	tests := []struct {
		name   string
		action string
		done   string
	}{
		{"install", "Installing", "Installed"},
		{"add", "Adding", "Added"},
		{"remove", "Removing", "Removed"},
		{"init", "Initializing", "Initialized"},
		{"deps", "Installing dependencies", "Dependencies installed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan tea.Msg)
			m := newProgressModel(ch, tt.action, tt.done)
			m.done = true
			m.exitCode = 0
			view := m.View()
			if !strings.Contains(view.Content, tt.done+" in ") {
				t.Errorf("success view should contain %q, got: %s", tt.done+" in ", view.Content)
			}
		})
	}
}

func TestViewShowsFailedOnError(t *testing.T) {
	ch := make(chan tea.Msg)
	m := newProgressModel(ch, "Installing", "Installed")
	m.done = true
	m.exitCode = 1
	view := m.View()
	if !strings.Contains(view.Content, "Failed in ") {
		t.Errorf("error view should contain 'Failed in', got: %s", view.Content)
	}
	if strings.Contains(view.Content, "Installed") {
		t.Errorf("error view should not contain 'Installed', got: %s", view.Content)
	}
}

func TestViewDoesNotShowDoneLabelOnError(t *testing.T) {
	tests := []struct {
		name   string
		action string
		done   string
	}{
		{"install", "Installing", "Installed"},
		{"remove", "Removing", "Removed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan tea.Msg)
			m := newProgressModel(ch, tt.action, tt.done)
			m.done = true
			m.exitCode = 1
			view := m.View()
			if strings.Contains(view.Content, tt.done) {
				t.Errorf("error view should not contain %q, got: %s", tt.done, view.Content)
			}
		})
	}
}

func TestDefaultLabels(t *testing.T) {
	cfg := Config{Args: []string{"echo"}}
	// Verify defaults are applied when labels are empty.
	action := cfg.Action
	if action == "" {
		action = "Installing"
	}
	done := cfg.Done
	if done == "" {
		done = "Installed"
	}

	ch := make(chan tea.Msg)
	m := newProgressModel(ch, action, done)

	view := m.View()
	if !strings.Contains(view.Content, "Installing...") {
		t.Errorf("default running view should contain 'Installing...', got: %s", view.Content)
	}

	m.done = true
	m.exitCode = 0
	view = m.View()
	if !strings.Contains(view.Content, "Installed in ") {
		t.Errorf("default success view should contain 'Installed in', got: %s", view.Content)
	}
}

func TestAddLine(t *testing.T) {
	ch := make(chan tea.Msg)
	m := newProgressModel(ch, "Installing", "Installed")

	for i := 0; i < maxLogLines+3; i++ {
		m.addLine("line")
	}

	if len(m.lines) != maxLogLines {
		t.Errorf("expected %d lines in ring buffer, got %d", maxLogLines, len(m.lines))
	}
}

func TestViewShowsLogLines(t *testing.T) {
	ch := make(chan tea.Msg)
	m := newProgressModel(ch, "Installing", "Installed")
	m.addLine("npm warn deprecated")
	m.addLine("added 42 packages")

	view := m.View()
	if !strings.Contains(view.Content, "npm warn deprecated") {
		t.Errorf("view should contain log line, got: %s", view.Content)
	}
	if !strings.Contains(view.Content, "added 42 packages") {
		t.Errorf("view should contain log line, got: %s", view.Content)
	}
}

func TestViewTruncatesLongLines(t *testing.T) {
	ch := make(chan tea.Msg)
	m := newProgressModel(ch, "Installing", "Installed")
	m.width = 20

	longLine := "this is a very long line that should be truncated to fit"
	m.addLine(longLine)

	view := m.View()
	// maxLen = width - 4 = 16, so truncated = 15 chars + "…"
	if strings.Contains(view.Content, longLine) {
		t.Error("view should truncate long lines")
	}
	if !strings.Contains(view.Content, "…") {
		t.Error("truncated line should end with '…'")
	}
}

func TestUpdateOutputLineMsg(t *testing.T) {
	ch := make(chan tea.Msg, 1)
	m := newProgressModel(ch, "Installing", "Installed")

	updated, _ := m.Update(outputLineMsg("hello world"))
	um := updated.(model)
	if len(um.lines) != 1 || um.lines[0] != "hello world" {
		t.Errorf("expected line 'hello world', got: %v", um.lines)
	}
}

func TestUpdateOutputLineMsgTrimsWhitespace(t *testing.T) {
	ch := make(chan tea.Msg, 1)
	m := newProgressModel(ch, "Installing", "Installed")

	updated, _ := m.Update(outputLineMsg("  hello  "))
	um := updated.(model)
	if len(um.lines) != 1 || um.lines[0] != "hello" {
		t.Errorf("expected trimmed line 'hello', got: %v", um.lines)
	}
}

func TestUpdateOutputLineMsgSkipsEmpty(t *testing.T) {
	ch := make(chan tea.Msg, 1)
	m := newProgressModel(ch, "Installing", "Installed")

	updated, _ := m.Update(outputLineMsg("   "))
	um := updated.(model)
	if len(um.lines) != 0 {
		t.Errorf("expected empty lines to be skipped, got: %v", um.lines)
	}
}

func TestUpdateDoneMsg(t *testing.T) {
	ch := make(chan tea.Msg)
	m := newProgressModel(ch, "Removing", "Removed")

	updated, _ := m.Update(doneMsg{exitCode: 0, err: nil})
	um := updated.(model)
	if !um.done {
		t.Error("model should be done after doneMsg")
	}
	if um.exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", um.exitCode)
	}
}

func TestUpdateDoneMsgWithError(t *testing.T) {
	ch := make(chan tea.Msg)
	m := newProgressModel(ch, "Removing", "Removed")

	updated, _ := m.Update(doneMsg{exitCode: 1, err: nil})
	um := updated.(model)
	if !um.done {
		t.Error("model should be done after doneMsg")
	}
	if um.exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", um.exitCode)
	}

	view := um.View()
	if !strings.Contains(view.Content, "Failed in ") {
		t.Errorf("error view should contain 'Failed in', got: %s", view.Content)
	}
}
