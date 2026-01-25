package menu

import (
	"testing"
	
	"github.com/tomncooper/dunst-waybar/internal/history"
)

func TestDetectTool(t *testing.T) {
	tool, err := detectTool()
	if err != nil {
		t.Skipf("No menu tools available: %v", err)
	}

	// Should return one of the valid tools
	validTools := []Tool{ToolRofi, ToolWofi, ToolDmenu}
	found := false
	for _, valid := range validTools {
		if tool == valid {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("detectTool returned invalid tool: %s", tool)
	}
}

func TestIsToolAvailable(t *testing.T) {
	tests := []struct {
		name string
		tool Tool
	}{
		{"rofi", ToolRofi},
		{"wofi", ToolWofi},
		{"dmenu", ToolDmenu},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available := isToolAvailable(tt.tool)
			// Just verify it returns a boolean, don't test availability
			// as it depends on the system
			_ = available
		})
	}
}

func TestIsToolAvailable_Invalid(t *testing.T) {
	tool := Tool("nonexistent-tool-xyz")
	if isToolAvailable(tool) {
		t.Error("Expected nonexistent tool to be unavailable")
	}
}

func TestNewLauncher_Auto(t *testing.T) {
	launcher, err := NewLauncher("auto")
	if err != nil {
		t.Skipf("No menu tools available: %v", err)
	}

	if launcher.tool == "" || launcher.tool == ToolAuto {
		t.Error("Expected launcher to have a specific tool set")
	}
}

func TestNewLauncher_Specific(t *testing.T) {
	// Try to create launcher with rofi
	launcher, err := NewLauncher("rofi")
	if err != nil {
		t.Skipf("rofi not available: %v", err)
	}

	if launcher.tool != ToolRofi {
		t.Errorf("Expected tool to be rofi, got %s", launcher.tool)
	}
}

func TestNewLauncher_Invalid(t *testing.T) {
	_, err := NewLauncher("nonexistent-tool-xyz")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}

func TestLauncher_Show_EmptyItems(t *testing.T) {
	launcher, err := NewLauncher("auto")
	if err != nil {
		t.Skipf("No menu tools available: %v", err)
	}

	_, err = launcher.Show([]history.MenuItem{}, "Test")
	if err == nil {
		t.Error("Expected error for empty items")
	}
}

func TestLauncher_FormatForRofi(t *testing.T) {
	launcher := &Launcher{tool: ToolRofi}
	
	items := []history.MenuItem{
		{Text: "Item 1", IconPath: "/path/to/icon1.png"},
		{Text: "Item 2", IconPath: ""},
		{Text: "Item 3", IconPath: "/path/to/icon3.svg"},
	}
	
	result := launcher.formatForRofi(items)
	
	// Check that icon paths are included with proper formatting
	if !contains(result, "Item 1\x00icon\x1f/path/to/icon1.png") {
		t.Error("Expected Item 1 with icon path")
	}
	if !contains(result, "Item 2\n") && !contains(result, "\nItem 3") {
		t.Error("Expected Item 2 without icon metadata")
	}
	if !contains(result, "Item 3\x00icon\x1f/path/to/icon3.svg") {
		t.Error("Expected Item 3 with icon path")
	}
}

func TestLauncher_FormatPlainText(t *testing.T) {
	launcher := &Launcher{tool: ToolDmenu}
	
	items := []history.MenuItem{
		{Text: "Item 1", IconPath: "/path/to/icon1.png"},
		{Text: "Item 2", IconPath: ""},
	}
	
	result := launcher.formatPlainText(items)
	
	expected := "Item 1\nItem 2"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Note: We can't easily test Show() with actual menu interaction
// as it requires a display and user interaction
// Integration tests would be needed for full coverage
