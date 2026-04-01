package search

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/mattn/go-isatty"

	"github.com/decampsrenan/spm/internal/ui"
)

// RunInteractive launches the interactive package search TUI.
// Returns the selected package name and whether --save-dev was toggled.
// Returns an error if the user cancelled or stdin is not a TTY.
func RunInteractive() (*Selection, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return nil, fmt.Errorf("no package specified and stdin is not a TTY — cannot prompt")
	}

	m := newModel()
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	ui.DrainTerminalResponses()
	if err != nil {
		return nil, fmt.Errorf("search TUI error: %w", err)
	}

	fm, ok := finalModel.(model)
	if !ok || !fm.selected {
		return nil, fmt.Errorf("no package selected")
	}

	pkg := fm.results[fm.cursor].Name
	if fm.selectedVersion != "" {
		pkg += "@" + fm.selectedVersion
	}
	return &Selection{
		Package: pkg,
		SaveDev: fm.saveDev,
	}, nil
}
