# dunst-waybar

A custom [Waybar](https://github.com/Alexays/Waybar) module for [Dunst](https://github.com/dunst-project/dunst) notification daemon integration, written in Go.

## Features

- 🔔 **Real-time status display** - Shows current notification state (paused/unpaused) with appropriate unicode icons
- 📊 **Waiting notification count** - Displays the number of queued notifications when paused
- 🖱️ **Click to toggle** - Easily pause/unpause notifications by clicking the module
- ⚡ **Efficient D-Bus integration** - Uses D-Bus signals for instant updates (no polling)
- ⚙️ **Highly configurable** - Customize icons, formats, and tooltips via JSON config
- 🛑 **Error handling** - Gracefully displays error state when Dunst is not running

## Screenshots

States:
- **Unpaused**: 🔔 (green) - Notifications are active
- **Paused**: 🔕 3 (red) - Notifications paused with 3 waiting
- **Error**: 🛑 (red) - Dunst is not running

## Installation

### Prerequisites

- Go 1.21 or later
- [Task](https://taskfile.dev) (optional, for convenient building)
  - Install: `go install github.com/go-task/task/v3/cmd/task@latest`
  - Or see [installation docs](https://taskfile.dev/installation)
- Dunst notification daemon
- Waybar
- D-Bus session bus

### From Source

#### Using Task (recommended)
```bash
git clone https://github.com/tcooper/dunst-waybar.git
cd dunst-waybar
task build
sudo task install
```

#### Using Go directly
```bash
git clone https://github.com/tcooper/dunst-waybar.git
cd dunst-waybar
go build -o dunst-waybar ./cmd/dunst-waybar
sudo install -Dm755 dunst-waybar /usr/local/bin/dunst-waybar
```

### Verify Installation

```bash
dunst-waybar --version
```

## Usage

### Quick Start

1. **Run the module** (test it first):
   ```bash
   dunst-waybar
   ```
   
   You should see JSON output like:
   ```json
   {"text":"🔔","tooltip":"Notifications active","class":"unpaused","alt":"unpaused"}
   ```

2. **Add to Waybar configuration** (`~/.config/waybar/config`):
   ```json
   {
     "modules-right": ["custom/dunst", "clock", "tray"],
     
     "custom/dunst": {
       "exec": "/usr/local/bin/dunst-waybar",
       "return-type": "json",
       "on-click": "dunstctl set-paused toggle",
       "tooltip": true
     }
   }
   ```

3. **Add styling** (`~/.config/waybar/style.css`):
   ```css
   #custom-dunst {
     padding: 0 10px;
     margin: 0 5px;
   }

   #custom-dunst.unpaused {
     color: #a6e3a1;
   }

   #custom-dunst.paused {
     color: #f38ba8;
   }

   #custom-dunst.error {
     color: #f38ba8;
     background: rgba(243, 139, 168, 0.2);
   }
   ```

4. **Restart Waybar**:
   ```bash
   killall waybar
   waybar &
   ```

## Configuration

### Config File Location

The config file is optional. By default, dunst-waybar looks for:
```
$XDG_CONFIG_HOME/waybar/dunst-waybar.json
```

Or if `XDG_CONFIG_HOME` is not set:
```
~/.config/waybar/dunst-waybar.json
```

You can also specify a custom path:
```bash
dunst-waybar --config /path/to/config.json
```

### Configuration Options

Create `~/.config/waybar/dunst-waybar.json`:

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

#### Configuration Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `icon-paused` | string | "🔕" | Icon shown when notifications are paused |
| `icon-unpaused` | string | "🔔" | Icon shown when notifications are active |
| `icon-error` | string | "🛑" | Icon shown when Dunst is not running |
| `format-paused` | string | "{icon} {waiting_count}" | Format for paused state text |
| `format-unpaused` | string | "{icon}" | Format for unpaused state text |
| `format-error` | string | "{icon}" | Format for error state text |
| `tooltip-format-paused` | string | "Notifications paused ({waiting_count} waiting)" | Tooltip when paused |
| `tooltip-format-unpaused` | string | "Notifications active" | Tooltip when unpaused |
| `tooltip-format-error` | string | "Dunst is not running" | Tooltip when error |
| `show-waiting-count` | boolean | true | Show the count of waiting notifications |
| `waiting-length-max` | integer | 9 | Maximum number to show before displaying "N+" |

#### Format Variables

The following variables can be used in format strings:

- `{icon}` - The current state icon
- `{waiting_count}` - Number of waiting notifications (or empty if 0)

### Example Configurations

**Minimal (no waiting count)**:
```json
{
  "format-paused": "{icon}",
  "show-waiting-count": false
}
```

**Custom icons (Font Awesome)**:
```json
{
  "icon-paused": "",
  "icon-unpaused": "",
  "icon-error": ""
}
```

**Verbose tooltips**:
```json
{
  "tooltip-format-paused": "🔕 Dunst notifications are paused. You have {waiting_count} notification(s) waiting.",
  "tooltip-format-unpaused": "🔔 Dunst notifications are active and will appear immediately."
}
```

## Waybar Integration

### Complete Waybar Module Example

```json
{
  "custom/dunst": {
    "exec": "/usr/local/bin/dunst-waybar",
    "return-type": "json",
    "on-click": "dunstctl set-paused toggle",
    "on-click-right": "dunstctl history-pop",
    "on-click-middle": "dunstctl close-all",
    "tooltip": true,
    "restart-interval": 1
  }
}
```

### Advanced Click Actions

- **Left click**: Toggle pause state
- **Right click**: Show last notification from history
- **Middle click**: Close all notifications

### CSS Styling Examples

See `examples/waybar-style.css` for complete examples. Here are some themes:

**Catppuccin Mocha**:
```css
#custom-dunst.unpaused { color: #a6e3a1; }
#custom-dunst.paused { color: #f38ba8; }
#custom-dunst.error { color: #f38ba8; }
```

**Material Design**:
```css
#custom-dunst.unpaused { color: #4caf50; }
#custom-dunst.paused { color: #f44336; }
#custom-dunst.error { color: #ff5252; }
```

## Troubleshooting

### Module doesn't appear in Waybar

1. Check that the binary is executable:
   ```bash
   chmod +x /usr/local/bin/dunst-waybar
   ```

2. Test the module manually:
   ```bash
   dunst-waybar
   ```
   
   You should see JSON output immediately.

3. Check Waybar logs:
   ```bash
   killall waybar
   waybar -l trace
   ```

### "Dunst is not running" error

1. Ensure Dunst is running:
   ```bash
   ps aux | grep dunst
   ```

2. Start Dunst if needed:
   ```bash
   dunst &
   ```

3. Check that Dunst is accessible via D-Bus:
   ```bash
   dunstctl is-paused
   ```

### Module shows old state

The module uses D-Bus signals for instant updates. If you see delays:

1. Check that D-Bus session bus is running
2. Verify Dunst version (signals added in recent versions)
3. Restart both Dunst and Waybar

### Click actions not working

1. Ensure `dunstctl` is in your PATH:
   ```bash
   which dunstctl
   ```

2. Test the toggle command manually:
   ```bash
   dunstctl set-paused toggle
   ```

3. Check Waybar's `on-click` configuration syntax

### High CPU usage

The module should use minimal CPU when idle (D-Bus signals only). If CPU is high:

1. Check for D-Bus connection issues
2. Verify Dunst is not spamming property changes
3. Review Waybar logs for errors

## How It Works

1. **Startup**: Connects to D-Bus session bus and queries initial Dunst state
2. **Monitoring**: Subscribes to `PropertiesChanged` signals from Dunst
3. **Output**: Formats state as JSON and writes to stdout (waybar reads this)
4. **Updates**: On property changes, immediately outputs new state
5. **Toggle**: Waybar's `on-click` calls `dunstctl` to toggle pause state
6. **Loop**: Dunst emits signal → module updates → Waybar re-renders

This design avoids polling, making it very efficient.

## Technical Details

### D-Bus Interface

- **Service**: `org.freedesktop.Notifications`
- **Object Path**: `/org/freedesktop/Notifications`
- **Interface**: `org.dunstproject.cmd0`
- **Properties**:
  - `paused` (boolean): Whether notifications are paused
  - `waitingLength` (uint32): Number of queued notifications

### Output Format

The module outputs JSON in Waybar's custom module format:

```json
{
  "text": "🔕 3",
  "tooltip": "Notifications paused (3 waiting)",
  "class": "paused",
  "alt": "paused"
}
```

- `text`: Displayed in the bar
- `tooltip`: Shown on hover
- `class`: CSS class for styling (unpaused/paused/error)
- `alt`: Alternative text (matches class)

## Comparison with waybar-dunst (Python)

This project is inspired by [waybar-dunst](https://github.com/CelDaemon/waybar-dunst) but rewritten in Go:

| Feature | dunst-waybar (Go) | waybar-dunst (Python) |
|---------|-------------------|----------------------|
| Language | Go | Python |
| Dependencies | godbus | dbus-fast |
| Binary Size | ~5MB | N/A (script) |
| Memory Usage | ~5-10MB | ~15-20MB |
| Startup Time | <10ms | ~50-100ms |
| Configuration | JSON | JSON |
| Installation | Single binary | Python + dependencies |

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Related Projects

- [Dunst](https://github.com/dunst-project/dunst) - Lightweight notification daemon
- [Waybar](https://github.com/Alexays/Waybar) - Highly customizable Wayland bar
- [waybar-dunst](https://github.com/CelDaemon/waybar-dunst) - Python implementation
- [godbus](https://github.com/godbus/dbus) - Go D-Bus library

## Acknowledgments

- Inspired by [waybar-dunst](https://github.com/CelDaemon/waybar-dunst) by CelDaemon
- Built using [godbus](https://github.com/godbus/dbus) for D-Bus integration
- Made for [Waybar](https://github.com/Alexays/Waybar) and [Dunst](https://github.com/dunst-project/dunst)
