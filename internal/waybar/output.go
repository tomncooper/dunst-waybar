package waybar

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tcooper/dunst-waybar/internal/config"
)

// State represents the current dunst state
type State int

const (
	StateUnpaused State = iota
	StatePaused
	StateError
)

// Output represents the JSON output for waybar
type Output struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
	Alt     string `json:"alt"`
}

// Formatter formats dunst state into waybar JSON output
type Formatter struct {
	config *config.Config
}

// NewFormatter creates a new formatter with the given config
func NewFormatter(cfg *config.Config) *Formatter {
	return &Formatter{config: cfg}
}

// Format generates waybar output for the given state
func (f *Formatter) Format(state State, waitingCount int) *Output {
	var text, tooltip, class, alt, icon string

	switch state {
	case StateUnpaused:
		icon = f.config.IconUnpaused
		text = f.formatString(f.config.FormatUnpaused, icon, waitingCount)
		tooltip = f.formatString(f.config.TooltipUnpaused, icon, waitingCount)
		class = "unpaused"
		alt = "unpaused"

	case StatePaused:
		icon = f.config.IconPaused
		text = f.formatString(f.config.FormatPaused, icon, waitingCount)
		tooltip = f.formatString(f.config.TooltipPaused, icon, waitingCount)
		class = "paused"
		alt = "paused"

	case StateError:
		icon = f.config.IconError
		text = f.formatString(f.config.FormatError, icon, waitingCount)
		tooltip = f.formatString(f.config.TooltipError, icon, waitingCount)
		class = "error"
		alt = "error"
	}

	return &Output{
		Text:    text,
		Tooltip: tooltip,
		Class:   class,
		Alt:     alt,
	}
}

// formatString replaces template variables in the format string
func (f *Formatter) formatString(format, icon string, waitingCount int) string {
	result := format

	result = strings.ReplaceAll(result, "{icon}", icon)

	// Format waiting count
	waitingStr := ""
	if f.config.ShowWaitingCount && waitingCount > 0 {
		if waitingCount > f.config.WaitingLengthMax {
			waitingStr = fmt.Sprintf("%d+", f.config.WaitingLengthMax)
		} else {
			waitingStr = fmt.Sprintf("%d", waitingCount)
		}
	}

	// Replace {waiting_count} and clean up surrounding spaces/punctuation if empty
	if waitingStr == "" {
		// Remove common patterns around empty waiting_count
		// Order matters - try most specific patterns first
		result = strings.ReplaceAll(result, " ({waiting_count} waiting)", "")
		result = strings.ReplaceAll(result, "({waiting_count} waiting)", "")
		result = strings.ReplaceAll(result, " {waiting_count}", "")
		result = strings.ReplaceAll(result, "{waiting_count} ", "")
		result = strings.ReplaceAll(result, "{waiting_count}", "")

		// Clean up any remaining double spaces
		result = strings.ReplaceAll(result, "  ", " ")
		result = strings.TrimSpace(result)
	} else {
		result = strings.ReplaceAll(result, "{waiting_count}", waitingStr)
	}

	return result
}

// Write writes the output to stdout as JSON
func (o *Output) Write() error {
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}

	// Print directly to stdout (println auto-flushes)
	fmt.Println(string(data))

	return nil
}
