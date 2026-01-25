package history

import (
	"fmt"
	"strings"

	"github.com/tomncooper/dunst-waybar/internal/config"
	"github.com/tomncooper/dunst-waybar/internal/dunst"
)

// Formatter formats notification history for display
type Formatter struct {
	cfg *config.Config
}

// NewFormatter creates a new history formatter
func NewFormatter(cfg *config.Config) *Formatter {
	return &Formatter{cfg: cfg}
}

// MenuItem represents a formatted menu item with optional icon
type MenuItem struct {
	Text     string
	IconPath string
}

// Format formats a list of notifications for menu display
func (f *Formatter) Format(notifications []dunst.Notification) []MenuItem {
	if len(notifications) == 0 {
		return []MenuItem{{Text: "No notifications"}}
	}

	formatted := make([]MenuItem, 0, len(notifications))
	for _, notif := range notifications {
		item := f.formatNotification(notif)
		formatted = append(formatted, item)
	}

	return formatted
}

// formatNotification formats a single notification according to the template
func (f *Formatter) formatNotification(notif dunst.Notification) MenuItem {
	format := f.cfg.History.Format

	// Determine icon to use (path or fallback icon name)
	iconPath := notif.IconPath
	if iconPath == "" {
		// Use standard icon theme name as fallback
		iconPath = f.getUrgencyIconName(notif.Urgency)
	}

	// Replace template variables
	result := format
	result = strings.ReplaceAll(result, "{icon}", "") // Icons handled via icon path
	result = strings.ReplaceAll(result, "{summary}", notif.Summary)
	result = strings.ReplaceAll(result, "{body}", notif.Body)
	result = strings.ReplaceAll(result, "{appname}", notif.AppName)
	result = strings.ReplaceAll(result, "{urgency}", notif.Urgency)
	result = strings.ReplaceAll(result, "{time}", notif.FormatTimestamp(f.cfg.History.TimeFormat == "relative"))
	result = strings.ReplaceAll(result, "{id}", fmt.Sprintf("%d", notif.ID))

	// Clean up any double spaces
	result = strings.Join(strings.Fields(result), " ")

	// Truncate if necessary
	if f.cfg.History.MaxLineLength > 0 && len(result) > f.cfg.History.MaxLineLength {
		truncLen := f.cfg.History.MaxLineLength - len(f.cfg.History.TruncateSuffix)
		if truncLen > 0 {
			result = result[:truncLen] + f.cfg.History.TruncateSuffix
		}
	}

	return MenuItem{
		Text:     result,
		IconPath: iconPath,
	}
}

// getUrgencyIconName returns an icon path or name based on urgency
// Uses the configured fallback icon theme
func (f *Formatter) getUrgencyIconName(urgency string) string {
	theme := f.cfg.History.FallbackIconTheme
	if theme == "" {
		theme = "Adwaita"
	}

	// Map of urgency to icon names for different themes
	iconMaps := map[string]map[string][]string{
		"Adwaita": {
			"LOW":      {"/usr/share/icons/Adwaita/symbolic/status/dialog-information-symbolic.svg", "dialog-information-symbolic"},
			"NORMAL":   {"/usr/share/icons/Adwaita/symbolic/legacy/preferences-system-notifications-symbolic.svg", "preferences-system-notifications-symbolic"},
			"CRITICAL": {"/usr/share/icons/Adwaita/symbolic/status/dialog-error-symbolic.svg", "dialog-error-symbolic"},
		},
		"breeze": {
			"LOW":      {"/usr/share/icons/breeze/status/16/dialog-information-symbolic.svg", "dialog-information-symbolic"},
			"NORMAL":   {"/usr/share/icons/breeze/status/16/notifications-symbolic.svg", "notifications-symbolic"},
			"CRITICAL": {"/usr/share/icons/breeze/status/16/dialog-error-symbolic.svg", "dialog-error-symbolic"},
		},
		"breeze-dark": {
			"LOW":      {"/usr/share/icons/breeze-dark/status/16/dialog-information-symbolic.svg", "dialog-information-symbolic"},
			"NORMAL":   {"/usr/share/icons/breeze-dark/status/16/notifications-symbolic.svg", "notifications-symbolic"},
			"CRITICAL": {"/usr/share/icons/breeze-dark/status/16/dialog-error-symbolic.svg", "dialog-error-symbolic"},
		},
	}

	// Get icon paths for the configured theme, fallback to Adwaita
	themeName := strings.ToLower(theme)
	themeMap, ok := iconMaps[themeName]
	if !ok {
		themeMap = iconMaps["Adwaita"]
	}

	// Get icon paths for urgency, fallback to LOW
	paths, ok := themeMap[urgency]
	if !ok {
		paths = themeMap["LOW"]
	}

	// Return the first path (absolute path)
	// If it doesn't exist, rofi will fall back to the theme name (second element)
	return paths[0]
}


