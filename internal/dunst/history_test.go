package dunst

import (
	"os"
	"testing"
)

func TestParseHistory(t *testing.T) {
	// Read test fixture
	data, err := os.ReadFile("../../testdata/dunstctl-history.json")
	if err != nil {
		t.Skipf("Test fixture not found: %v", err)
	}

	notifications, err := ParseHistory(data)
	if err != nil {
		t.Fatalf("Failed to parse history: %v", err)
	}

	if len(notifications) == 0 {
		t.Skip("No notifications in test fixture")
	}

	// Validate first notification has expected fields
	first := notifications[0]
	if first.ID == 0 {
		t.Error("Expected ID to be set")
	}
	if first.Summary == "" {
		t.Error("Expected Summary to be set")
	}
	if first.AppName == "" {
		t.Error("Expected AppName to be set")
	}
	if first.Urgency == "" {
		t.Error("Expected Urgency to be set")
	}
}

func TestParseHistory_EmptyData(t *testing.T) {
	data := []byte(`{"type": "aa{sv}", "data": []}`)
	notifications, err := ParseHistory(data)
	if err != nil {
		t.Fatalf("Failed to parse empty history: %v", err)
	}
	if len(notifications) != 0 {
		t.Errorf("Expected 0 notifications, got %d", len(notifications))
	}
}

func TestParseHistory_InvalidJSON(t *testing.T) {
	data := []byte(`invalid json`)
	_, err := ParseHistory(data)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestNotification_HasAction(t *testing.T) {
	tests := []struct {
		name     string
		notif    Notification
		expected bool
	}{
		{
			name:     "with action",
			notif:    Notification{DefaultActionName: "default"},
			expected: true,
		},
		{
			name:     "without action",
			notif:    Notification{DefaultActionName: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.notif.HasAction(); got != tt.expected {
				t.Errorf("HasAction() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNotification_FormatTimestamp(t *testing.T) {
	notif := Notification{Timestamp: 34300328}

	// Test relative format
	relative := notif.FormatTimestamp(true)
	if relative == "" {
		t.Error("Expected non-empty relative timestamp")
	}

	// Test absolute format
	absolute := notif.FormatTimestamp(false)
	if absolute == "" {
		t.Error("Expected non-empty absolute timestamp")
	}

	// Test zero timestamp
	zeroNotif := Notification{Timestamp: 0}
	if got := zeroNotif.FormatTimestamp(true); got != "" {
		t.Errorf("Expected empty string for zero timestamp, got %q", got)
	}
}

func TestParseHistory_SortOrder(t *testing.T) {
	// Create test data with notifications in random order
	data := []byte(`{
		"type": "aa{sv}",
		"data": [
			[
				{
					"id": {"type": "i", "data": 1},
					"summary": {"type": "s", "data": "Old"},
					"appname": {"type": "s", "data": "test"},
					"urgency": {"type": "s", "data": "NORMAL"},
					"timestamp": {"type": "x", "data": 1000000}
				},
				{
					"id": {"type": "i", "data": 2},
					"summary": {"type": "s", "data": "Newest"},
					"appname": {"type": "s", "data": "test"},
					"urgency": {"type": "s", "data": "NORMAL"},
					"timestamp": {"type": "x", "data": 3000000}
				},
				{
					"id": {"type": "i", "data": 3},
					"summary": {"type": "s", "data": "Middle"},
					"appname": {"type": "s", "data": "test"},
					"urgency": {"type": "s", "data": "NORMAL"},
					"timestamp": {"type": "x", "data": 2000000}
				}
			]
		]
	}`)

	notifications, err := ParseHistory(data)
	if err != nil {
		t.Fatalf("Failed to parse history: %v", err)
	}

	if len(notifications) != 3 {
		t.Fatalf("Expected 3 notifications, got %d", len(notifications))
	}

	// Verify notifications are sorted by timestamp descending (newest first)
	if notifications[0].Timestamp != 3000000 {
		t.Errorf("First notification should be newest, got timestamp %d", notifications[0].Timestamp)
	}
	if notifications[1].Timestamp != 2000000 {
		t.Errorf("Second notification should be middle, got timestamp %d", notifications[1].Timestamp)
	}
	if notifications[2].Timestamp != 1000000 {
		t.Errorf("Third notification should be oldest, got timestamp %d", notifications[2].Timestamp)
	}

	// Verify summaries match expected order
	if notifications[0].Summary != "Newest" {
		t.Errorf("First notification summary should be 'Newest', got %q", notifications[0].Summary)
	}
	if notifications[1].Summary != "Middle" {
		t.Errorf("Second notification summary should be 'Middle', got %q", notifications[1].Summary)
	}
	if notifications[2].Summary != "Old" {
		t.Errorf("Third notification summary should be 'Old', got %q", notifications[2].Summary)
	}
}
