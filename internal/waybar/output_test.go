package waybar

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tcooper/dunst-waybar/internal/config"
)

func TestNewFormatter(t *testing.T) {
	cfg := config.Default()
	formatter := NewFormatter(cfg)

	if formatter == nil {
		t.Fatal("NewFormatter() returned nil")
	}

	if formatter.config != cfg {
		t.Error("Formatter config not set correctly")
	}
}

func TestFormat_AllStates(t *testing.T) {
	tests := []struct {
		name          string
		state         State
		waitingCount  int
		expectText    string
		expectTooltip string
		expectClass   string
		expectAlt     string
	}{
		{
			name:          "unpaused no waiting",
			state:         StateUnpaused,
			waitingCount:  0,
			expectText:    "🔔",
			expectTooltip: "Notifications active",
			expectClass:   "unpaused",
			expectAlt:     "unpaused",
		},
		{
			name:          "unpaused with waiting",
			state:         StateUnpaused,
			waitingCount:  3,
			expectText:    "🔔",
			expectTooltip: "Notifications active",
			expectClass:   "unpaused",
			expectAlt:     "unpaused",
		},
		{
			name:          "paused no waiting",
			state:         StatePaused,
			waitingCount:  0,
			expectText:    "🔕",
			expectTooltip: "Notifications paused",
			expectClass:   "paused",
			expectAlt:     "paused",
		},
		{
			name:          "paused with 1 waiting",
			state:         StatePaused,
			waitingCount:  1,
			expectText:    "🔕 1",
			expectTooltip: "Notifications paused (1 waiting)",
			expectClass:   "paused",
			expectAlt:     "paused",
		},
		{
			name:          "paused with 5 waiting",
			state:         StatePaused,
			waitingCount:  5,
			expectText:    "🔕 5",
			expectTooltip: "Notifications paused (5 waiting)",
			expectClass:   "paused",
			expectAlt:     "paused",
		},
		{
			name:          "paused with 9 waiting (at max)",
			state:         StatePaused,
			waitingCount:  9,
			expectText:    "🔕 9",
			expectTooltip: "Notifications paused (9 waiting)",
			expectClass:   "paused",
			expectAlt:     "paused",
		},
		{
			name:          "paused with 10 waiting (over max)",
			state:         StatePaused,
			waitingCount:  10,
			expectText:    "🔕 9+",
			expectTooltip: "Notifications paused (9+ waiting)",
			expectClass:   "paused",
			expectAlt:     "paused",
		},
		{
			name:          "paused with 100 waiting (way over max)",
			state:         StatePaused,
			waitingCount:  100,
			expectText:    "🔕 9+",
			expectTooltip: "Notifications paused (9+ waiting)",
			expectClass:   "paused",
			expectAlt:     "paused",
		},
		{
			name:          "error state",
			state:         StateError,
			waitingCount:  0,
			expectText:    "🛑",
			expectTooltip: "Dunst is not running",
			expectClass:   "error",
			expectAlt:     "error",
		},
	}

	cfg := config.Default()
	formatter := NewFormatter(cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatter.Format(tt.state, tt.waitingCount)

			if output.Text != tt.expectText {
				t.Errorf("text: got %q, want %q", output.Text, tt.expectText)
			}

			if output.Tooltip != tt.expectTooltip {
				t.Errorf("tooltip: got %q, want %q", output.Tooltip, tt.expectTooltip)
			}

			if output.Class != tt.expectClass {
				t.Errorf("class: got %q, want %q", output.Class, tt.expectClass)
			}

			if output.Alt != tt.expectAlt {
				t.Errorf("alt: got %q, want %q", output.Alt, tt.expectAlt)
			}
		})
	}
}

func TestFormatString_IconReplacement(t *testing.T) {
	cfg := config.Default()
	formatter := NewFormatter(cfg)

	result := formatter.formatString("{icon} test", "🔔", 0)
	expected := "🔔 test"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestFormatString_WaitingCountReplacement(t *testing.T) {
	cfg := config.Default()
	formatter := NewFormatter(cfg)

	tests := []struct {
		name     string
		format   string
		count    int
		expected string
	}{
		{
			name:     "count 0 removes placeholder",
			format:   "{icon} {waiting_count}",
			count:    0,
			expected: "{icon}",
		},
		{
			name:     "count 3 shows number",
			format:   "{icon} {waiting_count}",
			count:    3,
			expected: "{icon} 3",
		},
		{
			name:     "count 10 shows 9+",
			format:   "{icon} {waiting_count}",
			count:    10,
			expected: "{icon} 9+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatString(tt.format, "{icon}", tt.count)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatString_EmptyPatternCleanup(t *testing.T) {
	cfg := config.Default()
	formatter := NewFormatter(cfg)

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "removes ({waiting_count} waiting)",
			format:   "Paused ({waiting_count} waiting)",
			expected: "Paused",
		},
		{
			name:     "removes leading space + pattern",
			format:   "Paused ({waiting_count} waiting)",
			expected: "Paused",
		},
		{
			name:     "removes {waiting_count} in middle",
			format:   "Icon {waiting_count} here",
			expected: "Icon here", // Spaces will be collapsed
		},
		{
			name:     "removes trailing {waiting_count}",
			format:   "Count: {waiting_count}",
			expected: "Count:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatString(tt.format, "icon", 0)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatString_ShowWaitingCountFalse(t *testing.T) {
	cfg := config.Default()
	cfg.ShowWaitingCount = false
	formatter := NewFormatter(cfg)

	// Even with count > 0, should not show if ShowWaitingCount is false
	result := formatter.formatString("{icon} {waiting_count}", "🔔", 5)
	expected := "🔔"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestOutput_Write(t *testing.T) {
	output := &Output{
		Text:    "🔔",
		Tooltip: "Active",
		Class:   "unpaused",
		Alt:     "unpaused",
	}

	// We can't easily test actual stdout, but we can marshal to JSON
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// Verify it's valid JSON
	var decoded Output
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if decoded.Text != output.Text {
		t.Errorf("text: got %q, want %q", decoded.Text, output.Text)
	}
	if decoded.Tooltip != output.Tooltip {
		t.Errorf("tooltip: got %q, want %q", decoded.Tooltip, output.Tooltip)
	}
	if decoded.Class != output.Class {
		t.Errorf("class: got %q, want %q", decoded.Class, output.Class)
	}
	if decoded.Alt != output.Alt {
		t.Errorf("alt: got %q, want %q", decoded.Alt, output.Alt)
	}
}

func TestFormat_WithCustomConfig(t *testing.T) {
	cfg := &config.Config{
		IconPaused:       "⏸️",
		IconUnpaused:     "▶️",
		IconError:        "❌",
		FormatPaused:     "{icon} ({waiting_count})",
		FormatUnpaused:   "{icon}",
		FormatError:      "{icon} ERROR",
		TooltipPaused:    "Paused: {waiting_count}",
		TooltipUnpaused:  "Active",
		TooltipError:     "Error!",
		ShowWaitingCount: true,
		WaitingLengthMax: 5,
	}

	formatter := NewFormatter(cfg)

	t.Run("custom icons", func(t *testing.T) {
		output := formatter.Format(StatePaused, 3)
		if !strings.Contains(output.Text, "⏸️") {
			t.Errorf("expected custom icon ⏸️ in text: %q", output.Text)
		}
	})

	t.Run("custom max", func(t *testing.T) {
		output := formatter.Format(StatePaused, 6)
		if !strings.Contains(output.Text, "5+") {
			t.Errorf("expected 5+ in text (custom max): %q", output.Text)
		}
	})

	t.Run("custom format", func(t *testing.T) {
		output := formatter.Format(StateError, 0)
		expected := "❌ ERROR"
		if output.Text != expected {
			t.Errorf("text: got %q, want %q", output.Text, expected)
		}
	})
}

func TestFormat_ComplexPatterns(t *testing.T) {
	cfg := config.Default()
	formatter := NewFormatter(cfg)

	// Test that our empty cleanup works with the default tooltip pattern
	output := formatter.Format(StatePaused, 0)

	// Should be clean without "(0 waiting)" or "( waiting)"
	if strings.Contains(output.Tooltip, "()") {
		t.Errorf("tooltip contains empty parens: %q", output.Tooltip)
	}
	if strings.Contains(output.Tooltip, "( waiting)") {
		t.Errorf("tooltip contains malformed pattern: %q", output.Tooltip)
	}

	// Should just be "Notifications paused"
	expected := "Notifications paused"
	if output.Tooltip != expected {
		t.Errorf("tooltip: got %q, want %q", output.Tooltip, expected)
	}
}
