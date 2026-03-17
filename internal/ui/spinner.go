package ui

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-isatty"
)

type spinnerDoneMsg struct{}

type spinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case spinnerDoneMsg:
		m.done = true
		return m, tea.Quit
	case tea.KeyPressMsg:
		if msg.(tea.KeyPressMsg).String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m spinnerModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	return tea.NewView(m.spinner.View() + " " + StyleDim.Render(m.message))
}

// WithSpinner runs fn concurrently while showing a spinner.
// If stdout is not a TTY, it runs fn directly without the spinner.
func WithSpinner[T any](message string, fn func() (T, error)) (T, error) {
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return fn()
	}

	s := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(ColorPrimary)),
	)

	m := spinnerModel{
		spinner: s,
		message: message,
	}

	p := tea.NewProgram(m)

	var result T
	var fnErr error
	go func() {
		result, fnErr = fn()
		p.Send(spinnerDoneMsg{})
	}()

	if _, err := p.Run(); err != nil {
		var zero T
		return zero, fmt.Errorf("spinner error: %w", err)
	}

	return result, fnErr
}
