package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tomncooper/dunst-waybar/internal/config"
	"github.com/tomncooper/dunst-waybar/internal/dunst"
	"github.com/tomncooper/dunst-waybar/internal/history"
	"github.com/tomncooper/dunst-waybar/internal/menu"
)

var (
	configPath = flag.String("config", "", "Path to config file (default: $XDG_CONFIG_HOME/waybar/dunst-waybar.json)")
	count      = flag.Int("count", 0, "Number of notifications to show (0 = all, default: from config or 0)")
	menuTool   = flag.String("menu-tool", "", "Menu tool to use: rofi, wofi, or dmenu (default: auto-detect)")
	version    = flag.Bool("version", false, "Print version information")
)

// Build-time version information (set via ldflags)
var (
	versionString = "dev"
	commit        = "none"
	date          = "unknown"
	builtBy       = "unknown"
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("dunst-waybar-history %s\n", versionString)
		if commit != "none" {
			fmt.Printf("  commit: %s\n", commit)
		}
		if date != "unknown" {
			fmt.Printf("  built: %s\n", date)
		}
		if builtBy != "unknown" {
			fmt.Printf("  by: %s\n", builtBy)
		}
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override config with command-line flags if provided
	historyCount := cfg.History.Count
	if *count > 0 {
		historyCount = *count
	}

	historyMenuTool := cfg.History.MenuTool
	if *menuTool != "" {
		historyMenuTool = *menuTool
	}

	// Fetch notification history
	notifications, err := dunst.FetchHistory()
	if err != nil {
		log.Fatalf("Failed to fetch notification history: %v", err)
	}

	// Limit to configured count
	if historyCount > 0 && len(notifications) > historyCount {
		notifications = notifications[:historyCount]
	}

	// Format notifications for display
	formatter := history.NewFormatter(cfg)
	formattedLines := formatter.Format(notifications)

	// Create menu launcher
	launcher, err := menu.NewLauncher(historyMenuTool)
	if err != nil {
		log.Fatalf("Failed to create menu launcher: %v", err)
	}

	// Show menu and get selection
	selected, err := launcher.Show(formattedLines, "Notification History")
	if err != nil {
		log.Fatalf("Failed to show menu: %v", err)
	}

	// Handle selection
	if err := dunst.HandleSelection(selected, notifications); err != nil {
		log.Fatalf("Failed to handle selection: %v", err)
	}
}

