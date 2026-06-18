package progress

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestAddLine_AppendsBelowLimit(t *testing.T) {
	m := model{}
	m.addLine("one")
	m.addLine("two")
	if len(m.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(m.lines))
	}
	if m.lines[0] != "one" || m.lines[1] != "two" {
		t.Errorf("unexpected lines: %v", m.lines)
	}
}

func TestAddLine_RingBufferDropsOldest(t *testing.T) {
	m := model{}
	for i := 0; i < maxLogLines+3; i++ {
		m.addLine("line-" + string(rune('A'+i)))
	}
	if len(m.lines) != maxLogLines {
		t.Fatalf("expected %d lines (ring buffer), got %d", maxLogLines, len(m.lines))
	}
	// First line should be the (maxLogLines+3)-th added, i.e. index 3.
	want := "line-" + string(rune('A'+3))
	if m.lines[0] != want {
		t.Errorf("expected oldest line %q, got %q", want, m.lines[0])
	}
}

func TestUpdate_OutputLine_AppendsAndReturnsCmd(t *testing.T) {
	ch := make(chan tea.Msg, 1)
	m := newProgressModel(ch)
	next, cmd := m.Update(outputLineMsg("hello world"))
	nm := next.(model)
	if len(nm.lines) != 1 || nm.lines[0] != "hello world" {
		t.Errorf("expected line appended, got %v", nm.lines)
	}
	if cmd == nil {
		t.Error("expected a follow-up Cmd to continue listening")
	}
	close(ch)
}

func TestUpdate_OutputLine_IgnoresEmpty(t *testing.T) {
	ch := make(chan tea.Msg, 1)
	m := newProgressModel(ch)
	next, _ := m.Update(outputLineMsg("   "))
	nm := next.(model)
	if len(nm.lines) != 0 {
		t.Errorf("expected empty line to be dropped, got %v", nm.lines)
	}
	close(ch)
}

func TestUpdate_DoneMsg_SetsFieldsAndQuits(t *testing.T) {
	ch := make(chan tea.Msg)
	m := newProgressModel(ch)
	next, cmd := m.Update(doneMsg{exitCode: 42, err: nil})
	nm := next.(model)
	if !nm.done {
		t.Error("expected done=true")
	}
	if nm.exitCode != 42 {
		t.Errorf("expected exitCode=42, got %d", nm.exitCode)
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestUpdate_WindowSize_UpdatesWidth(t *testing.T) {
	m := newProgressModel(make(chan tea.Msg))
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if next.(model).width != 120 {
		t.Errorf("expected width=120, got %d", next.(model).width)
	}
}

func TestUpdate_TimerTick_WhileRunning_ReturnsCmd(t *testing.T) {
	m := newProgressModel(make(chan tea.Msg))
	_, cmd := m.Update(timerTickMsg(time.Now()))
	if cmd == nil {
		t.Error("expected timer tick to schedule next tick while running")
	}
}

func TestUpdate_TimerTick_WhenDone_NoCmd(t *testing.T) {
	m := newProgressModel(make(chan tea.Msg))
	m.done = true
	_, cmd := m.Update(timerTickMsg(time.Now()))
	if cmd != nil {
		t.Error("expected no follow-up tick when done")
	}
}

func TestView_SuccessShowsInstalled(t *testing.T) {
	m := newProgressModel(make(chan tea.Msg))
	m.done = true
	m.exitCode = 0
	out := m.View().Content
	if !strings.Contains(out, "Installed") {
		t.Errorf("expected 'Installed' in success view, got %q", out)
	}
}

func TestView_FailureShowsFailed(t *testing.T) {
	m := newProgressModel(make(chan tea.Msg))
	m.done = true
	m.exitCode = 1
	out := m.View().Content
	if !strings.Contains(out, "Failed") {
		t.Errorf("expected 'Failed' in failure view, got %q", out)
	}
}

func TestView_RunningShowsInstalling(t *testing.T) {
	m := newProgressModel(make(chan tea.Msg))
	out := m.View().Content
	if !strings.Contains(out, "Installing") {
		t.Errorf("expected 'Installing' while running, got %q", out)
	}
}

func TestView_TruncatesLongLines(t *testing.T) {
	m := newProgressModel(make(chan tea.Msg))
	m.width = 20
	m.addLine(strings.Repeat("x", 100))
	out := m.View().Content
	if !strings.Contains(out, "…") {
		t.Errorf("expected truncation marker in output, got %q", out)
	}
}

func TestInit_ReturnsBatch(t *testing.T) {
	m := newProgressModel(make(chan tea.Msg))
	if cmd := m.Init(); cmd == nil {
		t.Error("Init returned nil cmd")
	}
}

func TestListenCh_ReturnsMsg(t *testing.T) {
	ch := make(chan tea.Msg, 1)
	ch <- outputLineMsg("x")
	cmd := listenCh(ch)
	msg := cmd()
	if _, ok := msg.(outputLineMsg); !ok {
		t.Errorf("expected outputLineMsg, got %T", msg)
	}
}

func TestListenCh_ClosedChannelReturnsNil(t *testing.T) {
	ch := make(chan tea.Msg)
	close(ch)
	cmd := listenCh(ch)
	if msg := cmd(); msg != nil {
		t.Errorf("expected nil from closed channel, got %v", msg)
	}
}
