# dunst-waybar

Inspired by [waybar-dunst](https://github.com/CelDaemon/waybar-dunst)

A custom Waybar module for Dunst notification daemon integration, written in Go. 
This module displays the current notification state (paused/unpaused) and the number of waiting notifications directly in your Waybar. 
It uses efficient D-Bus signal-based communication for instant updates without polling, and supports customizable icons, formats, and click actions through a simple JSON configuration file.

## Installation

**Prerequisites:** Go 1.21+, Dunst, Waybar, D-Bus session bus

**From source:**
```bash
git clone https://github.com/tcooper/dunst-waybar.git
cd dunst-waybar
go build -o dunst-waybar ./cmd/dunst-waybar
sudo install -Dm755 dunst-waybar /usr/local/bin/dunst-waybar
```

Or using [Task](https://taskfile.dev):
```bash
task build && sudo task install
```

## Waybar Configuration

Add the module to `~/.config/waybar/config`:
```json
{
  "modules-right": ["custom/dunst", "clock", "tray"],
  
  "custom/dunst": {
    "exec": "/usr/local/bin/dunst-waybar",
    "return-type": "json",
    "on-click": "dunstctl set-paused toggle",
    "on-click-right": "dunstctl history-pop",
    "tooltip": true
  }
}
```

Add styling to `~/.config/waybar/style.css`:
```css
#custom-dunst.unpaused { color: #a6e3a1; }
#custom-dunst.paused { color: #f38ba8; }
#custom-dunst.error { color: #f38ba8; }
```

Restart Waybar:
```bash
killall waybar && waybar &
```

## Customization

Configuration is optional. Create `$XDG_CONFIG_HOME/waybar/dunst-waybar.json` to customize icons, formats, and behavior. For most linux setups, this is `~/.config/waybar/dunst-waybar.json`:

```json
{
  "icon-paused": "🔕",
  "icon-unpaused": "🔔",
  "icon-error": "🛑",
  "format-paused": "{icon} {waiting_count}",
  "format-unpaused": "{icon}",
  "format-error": "{icon}",
  "tooltip-format-paused": "Notifications paused ({waiting_count} waiting)",
  "tooltip-format-unpaused": "Notifications active",
  "tooltip-format-error": "Dunst is not running",
  "show-waiting-count": true,
  "waiting-length-max": 9
}
```

See `examples/` directory for more configuration examples.

## Development

**Testing:**
```bash
# Run unit tests
go test -v ./...

# Run with coverage
task test-coverage

# Run integration tests (requires Docker/Podman)
task test-integration
```

**Code quality:**
```bash
task fmt    # Format code
task vet    # Run linter
task lint   # Run both fmt and vet
```

**Contributing:**

Contributions are welcome! Please fork the repository, create a feature branch, test your changes, and submit a pull request. Ensure all tests pass and code is formatted with `go fmt`.

## License

MIT License - see LICENSE file for details
