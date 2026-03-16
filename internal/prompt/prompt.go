package prompt

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"

	"github.com/decampsrenan/spm/internal/detector"
)

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
func SelectScript(scripts []string) (string, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return "", fmt.Errorf("no script specified and stdin is not a TTY — cannot prompt")
	}

	var choice string
	err := survey.AskOne(&survey.Select{
		Message: "Select a script to run:",
		Options: scripts,
	}, &choice)
	if err != nil {
		return "", err
	}

	return choice, nil
}
