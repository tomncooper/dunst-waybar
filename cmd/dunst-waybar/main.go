package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tomncooper/dunst-waybar/internal/config"
	"github.com/tomncooper/dunst-waybar/internal/dunst"
	"github.com/tomncooper/dunst-waybar/internal/waybar"
)

var (
	configPath  = flag.String("config", "", "Path to config file (default: $XDG_CONFIG_HOME/waybar/dunst-waybar.json)")
	showVersion = flag.Bool("version", false, "Print version information")
)

// Build-time version information (set via ldflags)
var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	builtBy   = "unknown"
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("dunst-waybar %s\n", version)
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

	// Create formatter
	formatter := waybar.NewFormatter(cfg)

	// Create dunst client
	client, err := dunst.NewClient()
	if err != nil {
		// If we can't connect to D-Bus at all, output error state and exit
		outputError(formatter)
		log.Fatalf("Failed to create dunst client: %v", err)
	}
	defer client.Close()

	// Start listening for updates
	if err := client.Start(); err != nil {
		// Dunst not running, output error state but continue running
		// (dunst might start later)
		outputError(formatter)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main event loop
	for {
		select {
		case update := <-client.Updates():
			handleUpdate(formatter, update)

		case <-sigChan:
			// Graceful shutdown
			return
		}
	}
}

// handleUpdate processes a state update from dunst
func handleUpdate(formatter *waybar.Formatter, update dunst.StateUpdate) {
	var output *waybar.Output

	if update.IsError {
		output = formatter.Format(waybar.StateError, 0)
	} else if update.Paused {
		output = formatter.Format(waybar.StatePaused, update.WaitingLength)
	} else {
		output = formatter.Format(waybar.StateUnpaused, update.WaitingLength)
	}

	if err := output.Write(); err != nil {
		log.Printf("Failed to write output: %v", err)
	}
}

// outputError outputs an error state
func outputError(formatter *waybar.Formatter) {
	output := formatter.Format(waybar.StateError, 0)
	if err := output.Write(); err != nil {
		log.Printf("Failed to write error output: %v", err)
	}
}
