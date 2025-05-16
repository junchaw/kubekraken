package utils

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Initialize color support
func init() {
	// Check if colors are supported and set the appropriate color profile
	if !termenv.EnvNoColor() {
		if termenv.ColorProfile() == termenv.Ascii {
			lipgloss.SetColorProfile(termenv.Ascii)
		}
	}
}

// Style contains all the styled output definitions
var Style = struct {
	Info    lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Text    lipgloss.Style
	Dim     lipgloss.Style
}{
	Info: lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")), // Blue

	Success: lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")), // Green

	Warning: lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")), // Orange

	Error: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")), // Red

	Text: lipgloss.NewStyle(),

	Dim: lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")), // Gray
}

func Exec(name string, arg ...string) ([]byte, []byte, error) {
	var stdout = bytes.Buffer{}
	var stderr = bytes.Buffer{}
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func ExecWithStdin(stdin string, name string, arg ...string) ([]byte, []byte, error) {
	var stdout = bytes.Buffer{}
	var stderr = bytes.Buffer{}
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(stdin)
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}
