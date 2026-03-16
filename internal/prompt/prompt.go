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
