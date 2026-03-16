package prompt

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"

	"github.com/decampsrenan/spm/internal/detector"
)

// Confirm asks the user a yes/no question. Returns true if they confirmed.
func Confirm(message string) (bool, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return false, fmt.Errorf("confirmation required but stdin is not a TTY — use --yes to skip")
	}

	var confirmed bool
	err := survey.AskOne(&survey.Confirm{
		Message: message,
	}, &confirmed)
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

	options := make([]string, len(detections))
	for i, d := range detections {
		options[i] = string(d.PM)
	}

	var choice string
	err := survey.AskOne(&survey.Select{
		Message: "Multiple lock files detected. Which package manager?",
		Options: options,
	}, &choice)
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

	options := make([]string, len(scriptNames))
	for i, name := range scriptNames {
		cmd := scriptCmds[i]
		if len(cmd) > maxCmdLen {
			cmd = cmd[:maxCmdLen-1] + "…"
		}
		options[i] = fmt.Sprintf("%s — %s", name, cmd)
	}

	var choice string
	err := survey.AskOne(&survey.Select{
		Message: "Select a script to run:",
		Options: options,
	}, &choice)
	if err != nil {
		return "", err
	}

	// Extract the script name (everything before " — ")
	for i, opt := range options {
		if opt == choice {
			return scriptNames[i], nil
		}
	}

	return "", fmt.Errorf("unexpected selection: %s", choice)
}

// SelectFromAll asks the user to pick a package manager when no lock file is found.
func SelectFromAll(projectDir string) (detector.Detection, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return detector.Detection{}, fmt.Errorf("no lock file found but stdin is not a TTY — cannot prompt")
	}

	options := []string{string(detector.NPM), string(detector.Yarn), string(detector.Pnpm)}

	var choice string
	err := survey.AskOne(&survey.Select{
		Message: "No lock file found. Which package manager do you want to use?",
		Options: options,
	}, &choice)
	if err != nil {
		return detector.Detection{}, err
	}

	return detector.Detection{PM: detector.PackageManager(choice), Dir: projectDir}, nil
}
