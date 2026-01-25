package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// Test icons
	if cfg.IconPaused != "🔕" {
		t.Errorf("IconPaused: got %q, want %q", cfg.IconPaused, "🔕")
	}
	if cfg.IconUnpaused != "🔔" {
		t.Errorf("IconUnpaused: got %q, want %q", cfg.IconUnpaused, "🔔")
	}
	if cfg.IconError != "🛑" {
		t.Errorf("IconError: got %q, want %q", cfg.IconError, "🛑")
	}

	// Test formats
	if cfg.FormatPaused != "{icon} {waiting_count}" {
		t.Errorf("FormatPaused: got %q, want %q", cfg.FormatPaused, "{icon} {waiting_count}")
	}
	if cfg.FormatUnpaused != "{icon}" {
		t.Errorf("FormatUnpaused: got %q, want %q", cfg.FormatUnpaused, "{icon}")
	}
	if cfg.FormatError != "{icon}" {
		t.Errorf("FormatError: got %q, want %q", cfg.FormatError, "{icon}")
	}

	// Test tooltips
	if cfg.TooltipPaused != "Notifications paused ({waiting_count} waiting)" {
		t.Errorf("TooltipPaused: got %q, want %q", cfg.TooltipPaused, "Notifications paused ({waiting_count} waiting)")
	}
	if cfg.TooltipUnpaused != "Notifications active" {
		t.Errorf("TooltipUnpaused: got %q, want %q", cfg.TooltipUnpaused, "Notifications active")
	}
	if cfg.TooltipError != "Dunst is not running" {
		t.Errorf("TooltipError: got %q, want %q", cfg.TooltipError, "Dunst is not running")
	}

	// Test settings
	if !cfg.ShowWaitingCount {
		t.Error("ShowWaitingCount: got false, want true")
	}
	if cfg.WaitingLengthMax != 9 {
		t.Errorf("WaitingLengthMax: got %d, want %d", cfg.WaitingLengthMax, 9)
	}

	// Test history defaults
	if cfg.History.Count != 0 {
		t.Errorf("History.Count: got %d, want %d", cfg.History.Count, 0)
	}
	if cfg.History.Format != "{time} {summary}" {
		t.Errorf("History.Format: got %q, want %q", cfg.History.Format, "{time} {summary}")
	}
	if cfg.History.MenuTool != "auto" {
		t.Errorf("History.MenuTool: got %q, want %q", cfg.History.MenuTool, "auto")
	}
	if cfg.History.TimeFormat != "relative" {
		t.Errorf("History.TimeFormat: got %q, want %q", cfg.History.TimeFormat, "relative")
	}
	if cfg.History.MaxLineLength != 100 {
		t.Errorf("History.MaxLineLength: got %d, want %d", cfg.History.MaxLineLength, 100)
	}
	if cfg.History.TruncateSuffix != "..." {
		t.Errorf("History.TruncateSuffix: got %q, want %q", cfg.History.TruncateSuffix, "...")
	}
	if cfg.History.FallbackIconTheme != "Adwaita" {
		t.Errorf("History.FallbackIconTheme: got %q, want %q", cfg.History.FallbackIconTheme, "Adwaita")
	}
}

func TestLoad_ValidFile(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "configs", "valid.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.IconPaused != "🔕" {
		t.Errorf("IconPaused: got %q, want %q", cfg.IconPaused, "🔕")
	}
	if cfg.WaitingLengthMax != 9 {
		t.Errorf("WaitingLengthMax: got %d, want %d", cfg.WaitingLengthMax, 9)
	}
}

func TestLoad_PartialFile(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "configs", "partial.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Custom values from file
	if cfg.IconPaused != "🚫" {
		t.Errorf("IconPaused: got %q, want %q", cfg.IconPaused, "🚫")
	}
	if cfg.FormatPaused != "{icon} PAUSED" {
		t.Errorf("FormatPaused: got %q, want %q", cfg.FormatPaused, "{icon} PAUSED")
	}

	// Default values (not in file)
	if cfg.IconUnpaused != "🔔" {
		t.Errorf("IconUnpaused: got %q, want %q (expected default)", cfg.IconUnpaused, "🔔")
	}
	if cfg.WaitingLengthMax != 9 {
		t.Errorf("WaitingLengthMax: got %d, want %d (expected default)", cfg.WaitingLengthMax, 9)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "configs", "nonexistent.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() should not error on missing file: %v", err)
	}

	// Should return defaults
	if cfg.IconPaused != "🔕" {
		t.Errorf("IconPaused: got %q, want %q (expected default)", cfg.IconPaused, "🔕")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "configs", "invalid.json")
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should return error for invalid JSON")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	// Empty path should use XDG config path
	// This won't fail even if file doesn't exist (returns defaults)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() with empty path failed: %v", err)
	}

	// Should return defaults (file likely doesn't exist)
	if cfg.IconPaused != "🔕" {
		t.Errorf("IconPaused: got %q, want %q", cfg.IconPaused, "🔕")
	}
}

func TestGetXDGConfigPath(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

		os.Setenv("XDG_CONFIG_HOME", "/tmp/test-config")
		path := getXDGConfigPath()

		expected := "/tmp/test-config/waybar/dunst-waybar.json"
		if path != expected {
			t.Errorf("got %q, want %q", path, expected)
		}
	})

	// Test without XDG_CONFIG_HOME (uses $HOME/.config)
	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() {
			if originalXDG != "" {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
			}
		}()

		os.Unsetenv("XDG_CONFIG_HOME")
		path := getXDGConfigPath()

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".config", "waybar", "dunst-waybar.json")
		if path != expected {
			t.Errorf("got %q, want %q", path, expected)
		}
	})
}

func TestLoad_WithWorkingDirectory(t *testing.T) {
	// Change to project root to test relative paths
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Go up directories until we find testdata
	for i := 0; i < 3; i++ {
		if _, err := os.Stat("testdata"); err == nil {
			break
		}
		if err := os.Chdir(".."); err != nil {
			t.Fatal(err)
		}
	}
	defer os.Chdir(wd)

	// Now test should work
	path := filepath.Join("testdata", "configs", "valid.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.IconPaused != "🔕" {
		t.Errorf("IconPaused: got %q, want %q", cfg.IconPaused, "🔕")
	}
}
