package dunst

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"
)

// Notification represents a notification from Dunst history
type Notification struct {
	ID                uint32
	Summary           string
	Body              string
	AppName           string
	Urgency           string
	IconPath          string
	Timestamp         int64
	DefaultActionName string
}

// HistoryResponse represents the JSON structure from dunstctl history
type HistoryResponse struct {
	Type string            `json:"type"`
	Data []json.RawMessage `json:"data"`
}

// NotificationField represents a single field in the dunstctl output
type NotificationField struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// FetchHistory executes dunstctl history and parses the output
func FetchHistory() ([]Notification, error) {
	cmd := exec.Command("dunstctl", "history")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute dunstctl history: %w", err)
	}

	return ParseHistory(output)
}

// ParseHistory parses the JSON output from dunstctl history
func ParseHistory(data []byte) ([]Notification, error) {
	var response HistoryResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history response: %w", err)
	}

	// The data field contains an array, where the first element is the notification array
	if len(response.Data) == 0 {
		return []Notification{}, nil
	}

	var notificationFields []map[string]NotificationField
	if err := json.Unmarshal(response.Data[0], &notificationFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification fields: %w", err)
	}

	var notifications []Notification
	for _, fields := range notificationFields {
		notif, err := parseNotification(fields)
		if err != nil {
			// Skip malformed notifications
			continue
		}
		notifications = append(notifications, notif)
	}

	// Sort notifications by timestamp (most recent first)
	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].Timestamp > notifications[j].Timestamp
	})

	return notifications, nil
}

// parseNotification converts a map of fields to a Notification struct
func parseNotification(fields map[string]NotificationField) (Notification, error) {
	notif := Notification{}

	// Extract ID
	if field, ok := fields["id"]; ok {
		switch v := field.Data.(type) {
		case float64:
			notif.ID = uint32(v)
		}
	}

	// Extract Summary
	if field, ok := fields["summary"]; ok {
		if str, ok := field.Data.(string); ok {
			notif.Summary = str
		}
	}

	// Extract Body
	if field, ok := fields["body"]; ok {
		if str, ok := field.Data.(string); ok {
			notif.Body = str
		}
	}

	// Extract AppName
	if field, ok := fields["appname"]; ok {
		if str, ok := field.Data.(string); ok {
			notif.AppName = str
		}
	}

	// Extract Urgency
	if field, ok := fields["urgency"]; ok {
		if str, ok := field.Data.(string); ok {
			notif.Urgency = str
		}
	}

	// Extract IconPath
	if field, ok := fields["icon_path"]; ok {
		if str, ok := field.Data.(string); ok {
			notif.IconPath = str
		}
	}

	// Extract Timestamp (microseconds since some epoch)
	if field, ok := fields["timestamp"]; ok {
		switch v := field.Data.(type) {
		case float64:
			notif.Timestamp = int64(v)
		}
	}

	// Extract DefaultActionName
	if field, ok := fields["default_action_name"]; ok {
		if str, ok := field.Data.(string); ok {
			notif.DefaultActionName = str
		}
	}

	return notif, nil
}

// FormatTimestamp converts a timestamp to a human-readable format
func (n *Notification) FormatTimestamp(relative bool) string {
	if n.Timestamp == 0 {
		return ""
	}

	// dunstctl timestamps are microseconds since boot
	// Calculate elapsed time by reading system uptime
	elapsed, err := n.getElapsedTime()
	if err != nil || elapsed < 0 {
		return ""
	}

	if relative {
		return formatRelativeTime(elapsed)
	}
	
	// For absolute time, show when the notification was posted
	notificationTime := time.Now().Add(-elapsed)
	return notificationTime.Format("15:04")
}

// getElapsedTime calculates how long ago the notification was posted
func (n *Notification) getElapsedTime() (time.Duration, error) {
	// Read /proc/uptime to get system uptime in seconds
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	var uptimeSeconds float64
	_, err = fmt.Sscanf(string(data), "%f", &uptimeSeconds)
	if err != nil {
		return 0, err
	}

	// Convert uptime to microseconds
	uptimeMicroseconds := int64(uptimeSeconds * 1000000)
	
	// Calculate elapsed time
	elapsedMicroseconds := uptimeMicroseconds - n.Timestamp
	return time.Duration(elapsedMicroseconds) * time.Microsecond, nil
}

// formatRelativeTime formats a duration as relative time (e.g., "2m ago", "1h ago")
func formatRelativeTime(d time.Duration) string {
	seconds := int(d.Seconds())
	
	if seconds < 0 {
		return "just now"
	}
	
	if seconds < 60 {
		return fmt.Sprintf("%ds ago", seconds)
	}
	
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}
	
	days := hours / 24
	if days < 7 {
		return fmt.Sprintf("%dd ago", days)
	}
	
	weeks := days / 7
	return fmt.Sprintf("%dw ago", weeks)
}

// HasAction returns true if the notification has a default action
func (n *Notification) HasAction() bool {
	return n.DefaultActionName != ""
}
