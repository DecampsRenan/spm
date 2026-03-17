package ui

import (
	"image/color"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

// Color palette — adaptive to dark/light terminal backgrounds.

var hasDarkBG = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)

func lightDark(light, dark string) color.Color {
	if hasDarkBG {
		return lipgloss.Color(dark)
	}
	return lipgloss.Color(light)
}

var (
	ColorPrimary = lightDark("#6D28D9", "#7C3AED")
	ColorSuccess = lightDark("#059669", "#10B981")
	ColorError   = lightDark("#DC2626", "#EF4444")
	ColorWarning = lightDark("#D97706", "#F59E0B")
	ColorInfo    = lightDark("#2563EB", "#3B82F6")
	ColorDim     = lightDark("#4B5563", "#6B7280")
	ColorCommand = lightDark("#7C3AED", "#A78BFA")
)

var (
	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	StyleError   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	StyleInfo    = lipgloss.NewStyle().Foreground(ColorInfo)
	StyleDim     = lipgloss.NewStyle().Foreground(ColorDim)
	StyleCommand = lipgloss.NewStyle().Foreground(ColorCommand)
	StyleBold    = lipgloss.NewStyle().Bold(true)
	StylePath    = lipgloss.NewStyle().Foreground(ColorPrimary)
)

// Success returns a success-styled message with a checkmark prefix.
func Success(msg string) string {
	return StyleSuccess.Render("✓ " + msg)
}

// Error returns an error-styled message with a cross prefix.
func Error(msg string) string {
	return StyleError.Render("✗ " + msg)
}

// Warning returns a warning-styled message with a triangle prefix.
func Warning(msg string) string {
	return StyleWarning.Render("▲ " + msg)
}

// Info returns an info-styled message.
func Info(msg string) string {
	return StyleInfo.Render(msg)
}

// Dim returns a dimmed message for secondary text.
func Dim(msg string) string {
	return StyleDim.Render(msg)
}

// Command returns a styled command preview for dry-run output.
func Command(args []string) string {
	label := StyleDim.Render("Would run: ")
	cmd := StyleCommand.Render(strings.Join(args, " "))
	return label + cmd
}

// Header returns a bold header.
func Header(msg string) string {
	return StyleBold.Render(msg)
}

// Path returns a styled file path.
func Path(path string) string {
	return StylePath.Render(path)
}

// Println prints styled text to stdout, downsampling colors for the terminal.
func Println(a ...any) {
	lipgloss.Fprintln(os.Stdout, a...)
}

// Printf prints formatted styled text to stdout, downsampling colors.
func Printf(format string, a ...any) {
	lipgloss.Fprintf(os.Stdout, format, a...)
}

// Eprintln prints styled text to stderr, downsampling colors for the terminal.
func Eprintln(a ...any) {
	lipgloss.Fprintln(os.Stderr, a...)
}
