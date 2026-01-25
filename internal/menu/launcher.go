package menu

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/tomncooper/dunst-waybar/internal/history"
)

// Tool represents a menu tool (rofi, wofi, or dmenu)
type Tool string

const (
	ToolRofi  Tool = "rofi"
	ToolWofi  Tool = "wofi"
	ToolDmenu Tool = "dmenu"
	ToolAuto  Tool = "auto"
)

// Launcher handles menu tool execution
type Launcher struct {
	tool Tool
}

// NewLauncher creates a new menu launcher with the specified tool
// If tool is "auto", it will auto-detect the best available tool
func NewLauncher(tool string) (*Launcher, error) {
	t := Tool(tool)
	
	if t == ToolAuto || t == "" {
		detected, err := detectTool()
		if err != nil {
			return nil, err
		}
		t = detected
	}

	// Validate tool is available
	if !isToolAvailable(t) {
		return nil, fmt.Errorf("menu tool %q not found in PATH", t)
	}

	return &Launcher{tool: t}, nil
}

// Show displays the menu with the given items and returns the selected item
// Returns empty string and nil error if user cancelled
func (l *Launcher) Show(items []history.MenuItem, prompt string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to display")
	}

	// Build command
	var cmd *exec.Cmd
	var input string
	
	switch l.tool {
	case ToolRofi:
		cmd = exec.Command("rofi", "-dmenu", "-show-icons", "-p", prompt)
		input = l.formatForRofi(items)
	case ToolWofi:
		cmd = exec.Command("wofi", "--dmenu", "-p", prompt)
		input = l.formatPlainText(items)
	case ToolDmenu:
		cmd = exec.Command("dmenu", "-p", prompt)
		input = l.formatPlainText(items)
	default:
		return "", fmt.Errorf("unsupported menu tool: %s", l.tool)
	}

	// Prepare input
	cmd.Stdin = strings.NewReader(input)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	err := cmd.Run()
	if err != nil {
		// Exit code 1 typically means user cancelled
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil // User cancelled
		}
		return "", fmt.Errorf("menu tool failed: %w (stderr: %s)", err, stderr.String())
	}

	// Return selected item (trimmed)
	selected := strings.TrimSpace(stdout.String())
	return selected, nil
}

// formatForRofi formats menu items for rofi with icon support
// Format: text\0icon\x1f/path/to/icon\n
func (l *Launcher) formatForRofi(items []history.MenuItem) string {
	var buf strings.Builder
	for _, item := range items {
		buf.WriteString(item.Text)
		if item.IconPath != "" {
			buf.WriteString("\x00icon\x1f")
			buf.WriteString(item.IconPath)
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// formatPlainText formats menu items as plain text (for wofi/dmenu)
func (l *Launcher) formatPlainText(items []history.MenuItem) string {
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = item.Text
	}
	return strings.Join(lines, "\n")
}

// detectTool auto-detects the best available menu tool
// Priority: rofi > wofi > dmenu
func detectTool() (Tool, error) {
	tools := []Tool{ToolRofi, ToolWofi, ToolDmenu}
	for _, tool := range tools {
		if isToolAvailable(tool) {
			return tool, nil
		}
	}
	return "", fmt.Errorf("no menu tool found (tried: rofi, wofi, dmenu)")
}

// isToolAvailable checks if a tool is available in PATH
func isToolAvailable(tool Tool) bool {
	_, err := exec.LookPath(string(tool))
	return err == nil
}
