package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/decampsrenan/spm/internal/ui"
)

const (
	maxResults    = 10
	debounceDelay = 300 * time.Millisecond
)

// model is the bubbletea model for interactive package search.
type model struct {
	textInput  textinput.Model
	initCmd    tea.Cmd
	results    []Result
	cursor     int
	loading    bool
	saveDev    bool
	selected   bool
	err        error
	lastQuery  string
	debounceID int
	width      int
}

// Selection holds the user's final choice from the search TUI.
type Selection struct {
	Package string
	SaveDev bool
}

// Messages

type searchResultMsg struct {
	query   string
	results []Result
}

type searchErrMsg struct {
	query string
	err   error
}

type debounceMsg struct {
	id    int
	query string
}

func newModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type to search npm packages..."
	ti.CharLimit = 100

	// Focus must be called here (pointer receiver) so the state
	// is captured in the struct before bubbletea copies it.
	blinkCmd := ti.Focus()

	return model{
		textInput: ti,
		initCmd:   blinkCmd,
		width:     80,
	}
}

func (m model) Init() tea.Cmd {
	return m.initCmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
			return m, nil

		case "tab":
			m.saveDev = !m.saveDev
			return m, nil

		case "enter":
			if len(m.results) > 0 && m.cursor < len(m.results) {
				m.selected = true
				return m, tea.Quit
			}
			return m, nil
		}
		// For all other keys (regular characters), fall through
		// to the textinput update below.

	case searchResultMsg:
		if msg.query == m.textInput.Value() {
			m.results = msg.results
			m.loading = false
			m.cursor = 0
		}
		return m, nil

	case searchErrMsg:
		if msg.query == m.textInput.Value() {
			m.err = msg.err
			m.loading = false
		}
		return m, nil

	case debounceMsg:
		if msg.id == m.debounceID && msg.query != "" && msg.query == m.textInput.Value() {
			m.loading = true
			m.err = nil
			return m, doSearch(msg.query)
		}
		return m, nil
	}

	// Update text input and check if value changed.
	prevValue := m.textInput.Value()
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	if m.textInput.Value() != prevValue {
		m.debounceID++
		id := m.debounceID
		query := m.textInput.Value()

		if query == "" {
			m.results = nil
			m.loading = false
			m.err = nil
			return m, cmd
		}

		debounceCmd := tea.Tick(debounceDelay, func(t time.Time) tea.Msg {
			return debounceMsg{id: id, query: query}
		})
		return m, tea.Batch(cmd, debounceCmd)
	}

	return m, cmd
}

func (m model) View() tea.View {
	var b strings.Builder

	// Search input
	prompt := ui.StyleBold.Render("Search packages: ")
	b.WriteString("  " + prompt + m.textInput.View() + "\n\n")

	// Results or status
	if m.textInput.Value() == "" {
		b.WriteString(ui.Dim("  Type to search npm packages...") + "\n")
	} else if m.loading {
		b.WriteString(ui.Dim("  Searching...") + "\n")
	} else if m.err != nil {
		b.WriteString("  " + ui.Warning(fmt.Sprintf("Search failed: %v", m.err)) + "\n")
	} else if len(m.results) == 0 && m.lastQuery != "" {
		b.WriteString(ui.Dim("  No packages found.") + "\n")
	} else {
		for i, r := range m.results {
			cursor := "  "
			nameStyle := lipgloss.NewStyle().Foreground(ui.ColorPrimary)
			if i == m.cursor {
				cursor = ui.StyleSuccess.Render("❯ ")
				nameStyle = nameStyle.Bold(true)
			}

			name := nameStyle.Render(r.Name)
			version := ui.Dim("v" + r.Version)

			desc := r.Description
			// Truncate description to fit
			maxDesc := m.width - len(r.Name) - len(r.Version) - 12
			if maxDesc < 0 {
				maxDesc = 20
			}
			if len(desc) > maxDesc {
				desc = desc[:maxDesc-1] + "…"
			}
			desc = ui.Dim(desc)

			b.WriteString(fmt.Sprintf("  %s%s  %s  %s\n", cursor, name, version, desc))
		}
	}

	// Footer
	b.WriteString("\n")
	devLabel := "no"
	if m.saveDev {
		devLabel = ui.StyleSuccess.Render("yes")
	}
	footer := ui.Dim("  tab") + " devDependency: " + devLabel +
		ui.Dim("  •  enter") + " add" +
		ui.Dim("  •  esc") + " cancel"
	b.WriteString(footer + "\n")

	return tea.NewView(b.String())
}

// doSearch fires an async npm registry search.
func doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := Query(context.Background(), query, maxResults)
		if err != nil {
			return searchErrMsg{query: query, err: err}
		}
		return searchResultMsg{query: query, results: results}
	}
}
