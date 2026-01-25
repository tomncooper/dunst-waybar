package history

import (
	"testing"

	"github.com/tomncooper/dunst-waybar/internal/config"
	"github.com/tomncooper/dunst-waybar/internal/dunst"
)

func TestFormatter_Format_Empty(t *testing.T) {
	cfg := config.Default()
	formatter := NewFormatter(cfg)

	result := formatter.Format([]dunst.Notification{})
	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}
	if result[0].Text != "No notifications" {
		t.Errorf("Expected 'No notifications', got %q", result[0].Text)
	}
}

func TestFormatter_Format_Single(t *testing.T) {
	cfg := config.Default()
	formatter := NewFormatter(cfg)

	notifications := []dunst.Notification{
		{
			ID:       1,
			Summary:  "Test notification",
			Body:     "Test body",
			AppName:  "test-app",
			Urgency:  "NORMAL",
			IconPath: "/usr/share/icons/test.png",
		},
	}

	result := formatter.Format(notifications)
	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}

	// Check text contains summary
	if !contains(result[0].Text, "Test notification") {
		t.Errorf("Expected summary in output, got %q", result[0].Text)
	}
	
	// Check icon path is preserved
	if result[0].IconPath != "/usr/share/icons/test.png" {
		t.Errorf("Expected icon path to be preserved, got %q", result[0].IconPath)
	}
}

func TestFormatter_FormatNotification_Variables(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		notif    dunst.Notification
		contains []string
	}{
		{
			name:   "summary only",
			format: "{summary}",
			notif: dunst.Notification{
				Summary: "My Summary",
			},
			contains: []string{"My Summary"},
		},
		{
			name:   "icon placeholder removed",
			format: "{icon} {summary}",
			notif: dunst.Notification{
				Summary:  "Test",
				Urgency:  "NORMAL",
				IconPath: "/path/to/icon.png",
			},
			contains: []string{"Test"}, // icon removed from text
		},
		{
			name:   "appname and summary",
			format: "[{appname}] {summary}",
			notif: dunst.Notification{
				AppName: "firefox",
				Summary: "Page loaded",
			},
			contains: []string{"[firefox]", "Page loaded"},
		},
		{
			name:   "id in format",
			format: "#{id} {summary}",
			notif: dunst.Notification{
				ID:      42,
				Summary: "Test",
			},
			contains: []string{"#42", "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.History.Format = tt.format
			formatter := NewFormatter(cfg)

			result := formatter.formatNotification(tt.notif)
			for _, substr := range tt.contains {
				if !contains(result.Text, substr) {
					t.Errorf("Expected %q to contain %q", result.Text, substr)
				}
			}
		})
	}
}

func TestFormatter_IconPath(t *testing.T) {
	cfg := config.Default()
	cfg.History.Format = "{summary}"
	formatter := NewFormatter(cfg)

	tests := []struct {
		name             string
		iconPath         string
		urgency          string
		wantIconContains string
	}{
		{"no icon path - gets fallback", "", "NORMAL", "notification"},
		{"with icon path", "/usr/share/icons/breeze/status/64/dialog-information.svg", "NORMAL", "/usr/share/icons/breeze/status/64/dialog-information.svg"},
		{"different icon", "/usr/share/pixmaps/firefox.png", "LOW", "/usr/share/pixmaps/firefox.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notif := dunst.Notification{
				Summary:  "Test",
				IconPath: tt.iconPath,
				Urgency:  tt.urgency,
			}
			result := formatter.formatNotification(notif)
			
			if !contains(result.IconPath, tt.wantIconContains) {
				t.Errorf("Expected icon path to contain %q, got %q", tt.wantIconContains, result.IconPath)
			}
		})
	}
}

func TestFormatter_EmojiFallback(t *testing.T) {
	cfg := config.Default()
	cfg.History.Format = "{icon} {summary}"
	formatter := NewFormatter(cfg)

	tests := []struct {
		name         string
		urgency      string
		iconPath     string
		wantContains string
	}{
		{"low urgency no icon", "LOW", "", "dialog-information"},
		{"normal urgency no icon", "NORMAL", "", "notification"},
		{"critical urgency no icon", "CRITICAL", "", "dialog-error"},
		{"unknown urgency no icon", "UNKNOWN", "", "dialog-information"},
		{"with icon path - uses original", "NORMAL", "/path/to/icon.png", "/path/to/icon.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notif := dunst.Notification{
				Summary:  "Test",
				Urgency:  tt.urgency,
				IconPath: tt.iconPath,
			}
			result := formatter.formatNotification(notif)
			
			if !contains(result.IconPath, tt.wantContains) {
				t.Errorf("Expected icon path to contain %q, got %q", tt.wantContains, result.IconPath)
			}
			
			// Text should not contain emoji or icon placeholders
			if contains(result.Text, "{icon}") {
				t.Errorf("Text should not contain icon placeholder: %q", result.Text)
			}
		})
	}
}

func TestFormatter_FallbackIconTheme(t *testing.T) {
	tests := []struct {
		name         string
		theme        string
		urgency      string
		wantContains string
	}{
		{"Adwaita theme", "Adwaita", "NORMAL", "Adwaita"},
		{"breeze theme", "breeze", "NORMAL", "breeze"},
		{"breeze-dark theme", "breeze-dark", "CRITICAL", "breeze-dark"},
		{"unknown theme defaults to Adwaita", "UnknownTheme", "LOW", "Adwaita"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.History.FallbackIconTheme = tt.theme
			cfg.History.Format = "{icon} {summary}"
			formatter := NewFormatter(cfg)

			notif := dunst.Notification{
				Summary:  "Test",
				Urgency:  tt.urgency,
				IconPath: "", // No icon - should use fallback
			}
			result := formatter.formatNotification(notif)

			if !contains(result.IconPath, tt.wantContains) {
				t.Errorf("Expected icon path to contain %q, got %q", tt.wantContains, result.IconPath)
			}
		})
	}
}

func TestFormatter_Truncation(t *testing.T) {
	cfg := config.Default()
	cfg.History.Format = "{summary}"
	cfg.History.MaxLineLength = 20
	cfg.History.TruncateSuffix = "..."

	formatter := NewFormatter(cfg)

	notif := dunst.Notification{
		Summary: "This is a very long notification summary that should be truncated",
	}

	result := formatter.formatNotification(notif)
	if len(result.Text) > cfg.History.MaxLineLength {
		t.Errorf("Expected length <= %d, got %d: %q", cfg.History.MaxLineLength, len(result.Text), result.Text)
	}
	if !contains(result.Text, "...") {
		t.Errorf("Expected truncation suffix in result: %q", result.Text)
	}
}

func TestFormatter_NoTruncation(t *testing.T) {
	cfg := config.Default()
	cfg.History.Format = "{summary}"
	cfg.History.MaxLineLength = 0 // Disable truncation

	formatter := NewFormatter(cfg)

	longSummary := "This is a very long notification summary that should NOT be truncated because truncation is disabled"
	notif := dunst.Notification{Summary: longSummary}

	result := formatter.formatNotification(notif)
	if result.Text != longSummary {
		t.Errorf("Expected full summary, got truncated: %q", result.Text)
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
