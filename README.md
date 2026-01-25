# dunst-waybar

Inspired by [waybar-dunst](https://github.com/CelDaemon/waybar-dunst)

A custom Waybar module for Dunst notification daemon integration, written in Go. 
This module displays the current notification state (paused/unpaused) and the number of waiting notifications directly in your Waybar. 
It includes a companion history browser (`dunst-waybar-history`) for browsing recent notifications interactively.
It uses efficient D-Bus signal-based communication for instant updates without polling, and supports customizable icons, formats, and click actions through a simple JSON configuration file.

## Installation

**Prerequisites:** Dunst, Waybar, D-Bus session bus

### From Pre-built Packages (Recommended)

Download the latest release from the [releases page](https://github.com/tomncooper/dunst-waybar/releases).

#### RPM (Fedora/RHEL/CentOS/AlmaLinux)
```bash
sudo rpm -i dunst-waybar_*.rpm
```

#### DEB (Debian/Ubuntu)
```bash
sudo dpkg -i dunst-waybar_*.deb
```

#### APK (Alpine Linux)
```bash
sudo apk add --allow-untrusted dunst-waybar_*.apk
```

#### Binary Archive
```bash
tar xzf dunst-waybar_*_Linux_x86_64.tar.gz
sudo install -Dm755 dunst-waybar /usr/local/bin/dunst-waybar
```

### From Source

**Prerequisites:** Go 1.21+

```bash
git clone https://github.com/tomncooper/dunst-waybar.git
cd dunst-waybar
go build -o dunst-waybar ./cmd/dunst-waybar
go build -o dunst-waybar-history ./cmd/dunst-waybar-history
sudo install -Dm755 dunst-waybar /usr/local/bin/dunst-waybar
sudo install -Dm755 dunst-waybar-history /usr/local/bin/dunst-waybar-history
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
    "on-click-middle": "dunst-waybar-history",
    "on-click-right": "dunstctl history-pop",
    "tooltip": true
  }
}
```

**Click Actions:**
- **Left click**: Toggle notification pause state
- **Middle click**: Show notification history browser (requires rofi/wofi/dmenu)
- **Right click**: Pop last notification from history

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

## Notification History

The `dunst-waybar-history` command provides an interactive notification history browser. Middle-click the Waybar module to view recent notifications in a menu (rofi/wofi/dmenu).

**Features:**
- Browse last N notifications (default: 10, configurable)
- Selectable menu with customizable format
- Auto-detects available menu tool (rofi → wofi → dmenu)
- Clicking a notification re-displays it or executes its action

**Command-line options:**
```bash
dunst-waybar-history [OPTIONS]

Options:
  --count N          Number of notifications to show (0 = all, default: from config or 0)
  --menu-tool TOOL   Menu tool to use: rofi, wofi, or dmenu (default: auto-detect)
  --config PATH      Path to config file
  --version          Print version information
```

**Configuration** (add to `~/.config/waybar/dunst-waybar.json`):
```json
{
  "history": {
    "count": 0,
    "format": "{time} {summary}",
    "menu-tool": "auto",
    "time-format": "relative",
    "max-line-length": 100,
    "truncate-suffix": "..."
  }
}
```

**Format variables:**
- `{time}` - Time since notification (e.g., "2m ago", "1h ago", "3d ago")
- `{icon}` - Icon based on urgency (🔔 normal, 🚨 critical, ℹ️ low)
- `{summary}` - Notification summary text
- `{body}` - Notification body
- `{appname}` - Application name
- `{urgency}` - LOW, NORMAL, or CRITICAL
- `{id}` - Notification ID

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
  "waiting-length-max": 9,
  "history": {
    "count": 10,
    "format": "{time} {summary}",
    "menu-tool": "auto",
    "time-format": "relative",
    "max-line-length": 100,
    "truncate-suffix": "..."
  }
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

## Releases

For maintainers: See [RELEASE.md](RELEASE.md) for the release process.

## License

MIT License - see LICENSE file for details
