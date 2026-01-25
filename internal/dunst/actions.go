package dunst

import (
	"fmt"
	"os/exec"
	"strings"
)

// PopNotification pops a notification from history and displays it using dunstctl
func PopNotification(id uint32) error {
	cmd := exec.Command("dunstctl", "history-pop", fmt.Sprintf("%d", id))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pop notification %d: %w (output: %s)", id, err, string(output))
	}
	return nil
}

// InvokeAction invokes the default action for a notification
// For now, this just pops the notification as we can't easily invoke actions via dunstctl
func InvokeAction(notif Notification) error {
	// Currently dunstctl doesn't provide a direct way to invoke notification actions
	// The best we can do is pop the notification to re-display it
	// Users can then interact with it normally
	return PopNotification(notif.ID)
}

// HandleSelection handles the user's selection from the history menu
// If the notification has an action, it invokes it; otherwise it pops the notification
func HandleSelection(selected string, notifications []Notification) error {
	if selected == "" || selected == "No notifications" {
		// User cancelled or no notifications
		return nil
	}

	// Find the notification that matches the selected line
	// We need to match against the formatted output
	// For now, let's just pop the first notification that has the summary in the selection
	for _, notif := range notifications {
		if strings.Contains(selected, notif.Summary) {
			// Check if notification has an action
			if notif.HasAction() {
				return InvokeAction(notif)
			}
			return PopNotification(notif.ID)
		}
	}

	return fmt.Errorf("no matching notification found for selection: %q", selected)
}

// ExtractNotificationID attempts to extract a notification ID from a formatted line
// This is a helper for more robust selection matching
func ExtractNotificationID(formatted string) (uint32, bool) {
	// Look for patterns like "#42" or "ID:42" in the string
	// This assumes the format includes {id} somewhere
	parts := strings.Fields(formatted)
	for _, part := range parts {
		part = strings.TrimPrefix(part, "#")
		part = strings.TrimPrefix(part, "ID:")
		
		var id uint32
		if _, err := fmt.Sscanf(part, "%d", &id); err == nil && id > 0 {
			return id, true
		}
	}
	return 0, false
}
