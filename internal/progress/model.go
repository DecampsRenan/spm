package progress

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/decampsrenan/spm/internal/ui"
)

const maxLogLines = 5

// model is the bubbletea model for the install progress TUI.
type model struct {
	spinner   spinner.Model
	lines     []string // ring buffer of last N output lines
	startTime time.Time
	done      bool
	exitCode  int
	err       error
	width     int
	msgCh     <-chan tea.Msg
}

// Messages

type outputLineMsg string

type doneMsg struct {
	exitCode int
	err      error
}

type timerTickMsg time.Time

func newProgressModel(msgCh <-chan tea.Msg) model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(ui.ColorPrimary)

	return model{
		spinner:   s,
		startTime: time.Now(),
		width:     80,
		msgCh:     msgCh,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		listenCh(m.msgCh),
		tickTimer(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case outputLineMsg:
		line := string(msg)
		if line = strings.TrimSpace(line); line != "" {
			m.addLine(line)
		}
		return m, listenCh(m.msgCh)

	case doneMsg:
		m.done = true
		m.exitCode = msg.exitCode
		m.err = msg.err
		return m, tea.Quit

	case timerTickMsg:
		if m.done {
			return m, nil
		}
		return m, tickTimer()

	case spinner.TickMsg:
		if m.done {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() tea.View {
	var b strings.Builder
	elapsed := time.Since(m.startTime).Truncate(100 * time.Millisecond)

	if m.done {
		// Final summary
		if m.exitCode == 0 {
			status := ui.Success(fmt.Sprintf("Installed in %s", elapsed))
			b.WriteString("  " + status + "\n")
		} else {
			status := ui.Error(fmt.Sprintf("Failed in %s", elapsed))
			b.WriteString("  " + status + "\n")
		}
	} else {
		// Spinner + status
		status := fmt.Sprintf("  %s Installing...%s",
			m.spinner.View(),
			ui.Dim(fmt.Sprintf("%"+fmt.Sprintf("%d", max(1, m.width-24))+"s", elapsed)),
		)
		b.WriteString(status + "\n")
	}

	// Log lines — gradient: top lines are faded, bottom line is normal dim
	if len(m.lines) > 0 {
		b.WriteString("\n")
		total := len(m.lines)
		for i, line := range m.lines {
			// Truncate to terminal width
			display := line
			maxLen := m.width - 4
			if maxLen > 0 && len(display) > maxLen {
				display = display[:maxLen-1] + "…"
			}
			b.WriteString("    " + ui.DimGradient(display, i, total) + "\n")
		}
	}

	return tea.NewView(b.String())
}

// addLine appends a line to the ring buffer, keeping at most maxLogLines.
func (m *model) addLine(line string) {
	if len(m.lines) >= maxLogLines {
		m.lines = m.lines[1:]
	}
	m.lines = append(m.lines, line)
}

// listenCh returns a command that waits for the next message from the channel.
func listenCh(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// tickTimer schedules a timer update every 100ms.
func tickTimer() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return timerTickMsg(t)
	})
}
