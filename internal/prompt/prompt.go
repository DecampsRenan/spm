package prompt

import (
	"fmt"
	"os"

	"charm.land/huh/v2"
	"github.com/mattn/go-isatty"

	"github.com/decampsrenan/spm/internal/detector"
	"github.com/decampsrenan/spm/internal/ui"
)

func theme() huh.Theme {
	return huh.ThemeFunc(huh.ThemeCharm)
}

func runField(field huh.Field) error {
	return huh.NewForm(huh.NewGroup(field)).
		WithTheme(theme()).
		Run()
}

// Confirm asks the user a yes/no question. Returns true if they confirmed.
func Confirm(message string) (bool, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return false, fmt.Errorf("confirmation required but stdin is not a TTY — use --yes to skip")
	}

	var confirmed bool
	err := runField(
		huh.NewConfirm().
			Title(message).
			Value(&confirmed),
	)
	if err != nil {
		return false, err
	}
	return confirmed, nil
}

// Select asks the user to pick a package manager when multiple lock files are detected.
func Select(detections []detector.Detection) (detector.Detection, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return detector.Detection{}, fmt.Errorf("multiple lock files found but stdin is not a TTY — cannot prompt")
	}

	options := make([]huh.Option[string], len(detections))
	for i, d := range detections {
		options[i] = huh.NewOption(string(d.PM), string(d.PM))
	}

	var choice string
	err := runField(
		huh.NewSelect[string]().
			Title(ui.Info("Multiple lock files detected.") + " Which package manager?").
			Options(options...).
			Value(&choice),
	)
	if err != nil {
		return detector.Detection{}, err
	}

	for _, d := range detections {
		if string(d.PM) == choice {
			return d, nil
		}
	}

	return detector.Detection{}, fmt.Errorf("unexpected selection: %s", choice)
}

// SelectScript asks the user to pick a script from the available list.
// Each option is displayed as "name — command" with the command truncated if needed.
func SelectScript(scriptNames []string, scriptCmds []string) (string, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return "", fmt.Errorf("no script specified and stdin is not a TTY — cannot prompt")
	}

	const maxCmdLen = 40

	options := make([]huh.Option[int], len(scriptNames))
	for i, name := range scriptNames {
		cmd := scriptCmds[i]
		if len(cmd) > maxCmdLen {
			cmd = cmd[:maxCmdLen-1] + "…"
		}
		label := fmt.Sprintf("%s — %s", name, ui.Dim(cmd))
		options[i] = huh.NewOption(label, i)
	}

	var choice int
	err := runField(
		huh.NewSelect[int]().
			Title("Select a script to run:").
			Options(options...).
			Value(&choice),
	)
	if err != nil {
		return "", err
	}

	return scriptNames[choice], nil
}

// SelectPM asks the user to pick a package manager (used by spm init).
func SelectPM() (detector.PackageManager, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return "", fmt.Errorf("no package manager specified and stdin is not a TTY — pass it as argument: spm init <pm>")
	}

	options := []huh.Option[string]{
		huh.NewOption(string(detector.NPM), string(detector.NPM)),
		huh.NewOption(string(detector.Yarn), string(detector.Yarn)),
		huh.NewOption(string(detector.Pnpm), string(detector.Pnpm)),
		huh.NewOption(string(detector.Bun), string(detector.Bun)),
	}

	var choice string
	err := runField(
		huh.NewSelect[string]().
			Title("Which package manager?").
			Options(options...).
			Value(&choice),
	)
	if err != nil {
		return "", err
	}

	return detector.PackageManager(choice), nil
}

// SelectFromAll asks the user to pick a package manager when no lock file is found.
func SelectFromAll(projectDir string) (detector.Detection, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return detector.Detection{}, fmt.Errorf("no lock file found but stdin is not a TTY — cannot prompt")
	}

	options := []huh.Option[string]{
		huh.NewOption(string(detector.NPM), string(detector.NPM)),
		huh.NewOption(string(detector.Yarn), string(detector.Yarn)),
		huh.NewOption(string(detector.Pnpm), string(detector.Pnpm)),
		huh.NewOption(string(detector.Bun), string(detector.Bun)),
		huh.NewOption(string(detector.Deno), string(detector.Deno)),
	}

	var choice string
	err := runField(
		huh.NewSelect[string]().
			Title(ui.Warning("No lock file found.") + " Which package manager do you want to use?").
			Options(options...).
			Value(&choice),
	)
	if err != nil {
		return detector.Detection{}, err
	}

	return detector.Detection{PM: detector.PackageManager(choice), Dir: projectDir}, nil
}
