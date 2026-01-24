# AGENTS.md - Project Documentation for AI Coding Agents

This document provides essential information for AI coding agents working on the dunst-waybar project.

## Project Overview

**dunst-waybar** is a custom Waybar module for Dunst notification daemon integration, written in Go. It provides real-time notification status display in Waybar through D-Bus integration.

### Key Features
- Real-time D-Bus integration with Dunst notification daemon
- JSON output for Waybar custom module consumption
- Configurable icons and formats via JSON config
- Efficient event-driven updates (no polling)
- Shows notification state (paused/unpaused) and waiting notification count

### Technology Stack
- **Language**: Go 1.25.5
- **Build Tool**: Task (go-task)
- **Key Dependencies**: 
  - `github.com/godbus/dbus/v5` - D-Bus communication
  - `github.com/testcontainers/testcontainers-go` - Integration testing
- **Target Environment**: Linux with Wayland/Waybar

## Project Structure

```
.
├── cmd/
│   └── dunst-waybar/          # Main application entry point
├── internal/
│   ├── config/                # Configuration loading and parsing
│   ├── dunst/                 # D-Bus client for Dunst interaction
│   └── waybar/                # Output formatting for Waybar
├── test/
│   └── integration/           # Integration tests with testcontainers
├── testdata/                  # Test fixtures and data
├── examples/                  # Example configuration files
├── Taskfile.yml               # Task runner configuration
├── go.mod                     # Go module definition
└── README.md                  # User documentation
```

## Building and Running

### Prerequisites
- Go 1.25.5 or later
- Task runner (optional but recommended)
- Dunst notification daemon
- D-Bus session bus

### Using Task (Recommended)

Task is the primary build tool. All common operations are defined in `Taskfile.yml`:

```bash
# Build the binary
task build

# Run the program
task run

# Clean build artifacts
task clean
```

### Release Management

The project uses **GoReleaser** for automated releases with GitHub Actions:

```bash
# Test the release configuration locally
task release-test

# Build a snapshot release (no publishing)
task release-snapshot

# Check if ready to release (runs tests and builds)
task release-check

# Clean release artifacts
task release-clean
```

**Creating a Release:**
1. Ensure all tests pass: `task release-check`
2. Test locally (optional): `task release-snapshot`
3. Create and push an annotated tag: `git tag -a v1 -m "Release v1" && git push origin v1`
4. GitHub Actions automatically builds and publishes the release

**Release Artifacts Generated:**
- Multi-platform binaries (Linux amd64, arm64, arm)
- RPM packages (Fedora/RHEL/CentOS)
- DEB packages (Debian/Ubuntu)
- APK packages (Alpine Linux)
- Archives with documentation and examples
- SHA256 checksums

See `RELEASE.md` for detailed release process documentation.

### Using Go Directly

```bash
# Build
go build -o dunst-waybar ./cmd/dunst-waybar

# Run
./dunst-waybar

# Clean
go clean
```

### Installation

```bash
# Using Task (installs to /usr/local/bin)
sudo task install

# Manual installation
sudo install -Dm755 dunst-waybar /usr/local/bin/dunst-waybar
```

## Testing

### Test Structure
- **Unit tests**: `internal/*/` packages contain `*_test.go` files
- **Integration tests**: `test/integration/` contains tests using testcontainers
- **Test data**: `testdata/` contains fixtures and test configurations

### Running Tests

```bash
# Run all unit tests
task test
# or
go test -v ./...

# Run with coverage
task test-coverage

# Generate HTML coverage report
task test-coverage-html

# Run with race detector
task test-race

# Run integration tests (requires Docker/Podman)
task test-integration

# Run integration tests with Podman
task test-integration-podman

# Run all tests (unit + integration)
task test-all

# Run short tests only
task test-short

# Run verbose tests
task test-verbose
```

### Integration Test Notes
- Integration tests use testcontainers-go
- Require Docker or Podman socket available
- Tagged with `//go:build integration` to separate from unit tests
- Run with `-tags=integration` flag

## Code Quality

### Linting and Formatting

```bash
# Format code
task fmt
# or
go fmt ./...

# Run vet
task vet
# or
go vet ./...

# Run both fmt and vet
task lint

# Tidy dependencies
task tidy
# or
go mod tidy
```

### Pre-Release Checks

Before creating a release, run:
```bash
task release-check
```

This ensures:
- Code is formatted correctly (`task lint`)
- All unit tests pass (`task test`)
- Binary builds successfully (`task build`)

### Code Standards
- Follow standard Go formatting (enforced by `go fmt`)
- Run `go vet` before committing
- Maintain test coverage for new features
- Use meaningful variable and function names
- Add comments for exported functions and types

## Configuration

### Config File Location
Default: `$XDG_CONFIG_HOME/waybar/dunst-waybar.json` or `~/.config/waybar/dunst-waybar.json`

### Config Schema
The configuration is defined in `internal/config/` package and includes:
- Icon customization (paused, unpaused, error)
- Format strings with variables: `{icon}`, `{waiting_count}`
- Tooltip formats for each state
- Display options (show-waiting-count, waiting-length-max)

See `examples/dunst-waybar.json` for a complete example.

## D-Bus Integration

### D-Bus Details
- **Service**: `org.freedesktop.Notifications`
- **Object Path**: `/org/freedesktop/Notifications`
- **Interface**: `org.dunstproject.cmd0`
- **Properties**:
  - `paused` (boolean): Whether notifications are paused
  - `waitingLength` (uint32): Number of queued notifications

The application subscribes to `PropertiesChanged` signals for real-time updates.

## Output Format

The program outputs JSON to stdout in Waybar's custom module format:

```json
{
  "text": "🔔",
  "tooltip": "Notifications active",
  "class": "unpaused",
  "alt": "unpaused"
}
```

States:
- **unpaused**: Notifications active
- **paused**: Notifications paused (shows waiting count)
- **error**: Dunst not running or connection error

## Development Workflow

1. **Make changes** to source files in `cmd/` or `internal/`
2. **Format code**: `task fmt`
3. **Run linter**: `task vet`
4. **Run tests**: `task test`
5. **Build**: `task build`
6. **Test manually**: `./dunst-waybar` (requires Dunst running)
7. **Run integration tests**: `task test-integration` (optional)

## Common Development Tasks

### Adding a New Configuration Option
1. Update the config struct in `internal/config/`
2. Add default value in config initialization
3. Update config tests in `internal/config/config_test.go`
4. Use the new option in relevant code
5. Update examples and documentation

### Modifying Output Format
1. Edit output formatting logic in `internal/waybar/`
2. Update unit tests in `internal/waybar/output_test.go`
3. Test with actual Waybar integration

### Changing D-Bus Interaction
1. Modify D-Bus client in `internal/dunst/`
2. Ensure signal handling is preserved
3. Test with running Dunst instance
4. Consider adding integration tests

## Debugging

### Running with Debug Output
The application outputs JSON to stdout for Waybar consumption. For debugging:

```bash
# Run directly to see output
./dunst-waybar

# Test with dunstctl commands
dunstctl set-paused toggle
dunstctl is-paused
```

### D-Bus Debugging
```bash
# Check Dunst is running
ps aux | grep dunst

# Test D-Bus connection
dunstctl is-paused

# Monitor D-Bus signals
dbus-monitor "type='signal',interface='org.freedesktop.DBus.Properties'"
```

## Important Notes for Agents

1. **No Polling**: The application uses D-Bus signals, not polling. Preserve this architecture.
2. **Single Binary**: The project compiles to a single static binary with no runtime dependencies (except system libs).
3. **Minimal Output**: The application must output only valid JSON to stdout (Waybar requirement).
4. **Error Handling**: Gracefully handle Dunst not running (error state with appropriate icon/message).
5. **Configuration Optional**: The application works with defaults if no config file exists.
6. **Testing**: Always run unit tests. Integration tests require Docker/Podman.
7. **Build Tool**: Use Task for consistent builds across environments.
8. **Releases**: Use GoReleaser for releases - GitHub Actions automates the entire process when tags are pushed.
9. **Versioning**: The project uses monotonic versioning (v1, v2, v3...) - simpler than semantic versioning.

## GoReleaser Configuration

The project uses GoReleaser for automated releases:

### Configuration Files
- **`.goreleaser.yaml`** - Main GoReleaser configuration
- **`.github/workflows/release.yml`** - GitHub Actions workflow for automated releases
- **`RELEASE.md`** - Detailed release process documentation

### Key GoReleaser Features
1. **Multi-platform Builds**: Automatic cross-compilation for Linux (amd64, arm64, arm)
2. **Package Generation**: Creates RPM, DEB, and APK packages with proper metadata
3. **Archive Creation**: Bundles binaries with LICENSE, README, AGENTS.md, and examples
4. **Changelog Generation**: Auto-generates changelogs from conventional commits
5. **GitHub Integration**: Automatically creates GitHub releases with all artifacts
6. **Pre-release Hooks**: Runs `go mod tidy` and tests before building
7. **Build Metadata**: Injects version, commit, and build date into binaries

### Release Workflow
1. **Local Testing**: Use `task release-snapshot` or `task release-test` to verify configuration
2. **Tag Creation**: Push an annotated tag (e.g., `v1`, `v2`) to trigger release
3. **Automated Build**: GitHub Actions runs GoReleaser on tag push
4. **Artifact Publishing**: All binaries, packages, and checksums uploaded to GitHub Release

### Package Configuration
GoReleaser packages include:
- **Dependencies**: `dunst` (required)
- **Recommendations**: `waybar` (recommended)
- **Documentation**: Installed to `/usr/share/doc/dunst-waybar/`
- **Examples**: Installed to `/usr/share/doc/dunst-waybar/examples/`
- **Binary**: Installed to `/usr/bin/dunst-waybar`

### Making Changes to Releases
- Modify `.goreleaser.yaml` for build configuration, package metadata, or file inclusion
- Update GitHub Actions workflow in `.github/workflows/release.yml` for CI/CD changes
- Test locally with `task release-snapshot` before pushing tags
- See `RELEASE.md` for complete workflow documentation

## References

- Main documentation: [README.md](README.md)
- Implementation notes: [IMPLEMENTATION.md](IMPLEMENTATION.md)
- Release process: [RELEASE.md](RELEASE.md)
- GoReleaser config: [.goreleaser.yaml](.goreleaser.yaml)
- GitHub Actions workflow: [.github/workflows/release.yml](.github/workflows/release.yml)
- Example configs: [examples/](examples/)
- Waybar custom module docs: https://github.com/Alexays/Waybar/wiki/Module:-Custom
- Dunst D-Bus interface: https://github.com/dunst-project/dunst
- GoReleaser documentation: https://goreleaser.com/
