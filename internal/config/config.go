package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	IconPaused       string        `json:"icon-paused"`
	IconUnpaused     string        `json:"icon-unpaused"`
	IconError        string        `json:"icon-error"`
	FormatPaused     string        `json:"format-paused"`
	FormatUnpaused   string        `json:"format-unpaused"`
	FormatError      string        `json:"format-error"`
	TooltipPaused    string        `json:"tooltip-format-paused"`
	TooltipUnpaused  string        `json:"tooltip-format-unpaused"`
	TooltipError     string        `json:"tooltip-format-error"`
	ShowWaitingCount bool          `json:"show-waiting-count"`
	WaitingLengthMax int           `json:"waiting-length-max"`
	History          HistoryConfig `json:"history"`
}

// HistoryConfig holds configuration for notification history display
type HistoryConfig struct {
	Count             int    `json:"count"`
	Format            string `json:"format"`
	MenuTool          string `json:"menu-tool"`
	TimeFormat        string `json:"time-format"`
	MaxLineLength     int    `json:"max-line-length"`
	TruncateSuffix    string `json:"truncate-suffix"`
	FallbackIconTheme string `json:"fallback-icon-theme"`
}

// Default returns a Config with default values
func Default() *Config {
	return &Config{
		IconPaused:       "🔕",
		IconUnpaused:     "🔔",
		IconError:        "🛑",
		FormatPaused:     "{icon} {waiting_count}",
		FormatUnpaused:   "{icon}",
		FormatError:      "{icon}",
		TooltipPaused:    "Notifications paused ({waiting_count} waiting)",
		TooltipUnpaused:  "Notifications active",
		TooltipError:     "Dunst is not running",
		ShowWaitingCount: true,
		WaitingLengthMax: 9,
		History: HistoryConfig{
			Count:             0,
			Format:            "{time} {summary}",
			MenuTool:          "auto",
			TimeFormat:        "relative",
			MaxLineLength:     100,
			TruncateSuffix:    "...",
			FallbackIconTheme: "Adwaita",
		},
	}
}

// Load reads configuration from the specified path
// If path is empty, it uses the XDG config path
// If the file doesn't exist, it returns default config
func Load(path string) (*Config, error) {
	if path == "" {
		path = getXDGConfigPath()
	}

	// Start with defaults
	cfg := Default()

	// If file doesn't exist, return defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal directly into cfg (will overlay on defaults)
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// getXDGConfigPath returns the default config file path following XDG spec
func getXDGConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configHome = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configHome, "waybar", "dunst-waybar.json")
}
