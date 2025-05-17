package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"gopkg.in/yaml.v2"
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

// StyleText is a struct that contains a text and a style, so that we can get both styled or unstyled text
type StyleText struct {
	Text  string
	Style *lipgloss.Style
}

func (s StyleText) GetText() string {
	return s.Text
}

func (s StyleText) Render() string {
	if s.Style != nil {
		return s.Style.Render(s.Text)
	}
	return s.Text
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

func FileExt(format string) string {
	if format == "json" {
		return ".json"
	}
	if format == "yaml" || format == "yml" {
		return ".yaml"
	}
	return ".txt"
}

func PutFileWithFormat(path string, data any, format string, textFunc func() string) error {
	if format == "json" {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal data to json: %v", err)
		}
		return os.WriteFile(path, jsonData, 0600)
	}
	if format == "yaml" || format == "yml" {
		yamlData, err := yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data to yaml: %v", err)
		}
		return os.WriteFile(path, yamlData, 0600)
	}
	return os.WriteFile(path, []byte(textFunc()), 0600)
}
