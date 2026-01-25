package dunst

import (
	"testing"
)

func TestExtractNotificationID(t *testing.T) {
	tests := []struct {
		name      string
		formatted string
		wantID    uint32
		wantFound bool
	}{
		{
			name:      "with hash prefix",
			formatted: "#42 Test notification",
			wantID:    42,
			wantFound: true,
		},
		{
			name:      "with ID prefix",
			formatted: "ID:123 Another test",
			wantID:    123,
			wantFound: true,
		},
		{
			name:      "plain number",
			formatted: "5 Test",
			wantID:    5,
			wantFound: true,
		},
		{
			name:      "no ID",
			formatted: "Test notification without ID",
			wantID:    0,
			wantFound: false,
		},
		{
			name:      "zero ID",
			formatted: "#0 Test",
			wantID:    0,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotFound := ExtractNotificationID(tt.formatted)
			if gotID != tt.wantID {
				t.Errorf("ExtractNotificationID() ID = %v, want %v", gotID, tt.wantID)
			}
			if gotFound != tt.wantFound {
				t.Errorf("ExtractNotificationID() found = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func TestHandleSelection_Empty(t *testing.T) {
	err := HandleSelection("", []Notification{})
	if err != nil {
		t.Errorf("Expected nil error for empty selection, got %v", err)
	}
}

func TestHandleSelection_NoNotifications(t *testing.T) {
	err := HandleSelection("No notifications", []Notification{})
	if err != nil {
		t.Errorf("Expected nil error for 'No notifications', got %v", err)
	}
}

// Note: Testing PopNotification and InvokeAction requires dunstctl to be running
// These would be better suited for integration tests
