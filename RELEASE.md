# Release Process

This document describes how to create and publish releases for dunst-waybar.

## Overview

Releases are automated using [GoReleaser](https://goreleaser.com/) and GitHub Actions. When you push a git tag, GitHub Actions automatically builds:

- Multi-platform binaries (Linux amd64, arm64, arm)
- RPM packages (Fedora, RHEL, CentOS, etc.)
- DEB packages (Debian, Ubuntu, etc.)
- APK packages (Alpine Linux)
- Checksums and release notes

## Prerequisites

### For Testing Locally
Install GoReleaser:
```bash
go install github.com/goreleaser/goreleaser/v2@latest
```

### For Publishing (via GitHub Actions)
No additional setup needed! GitHub Actions handles everything automatically.

## Release Workflow

### 1. Prepare for Release

Ensure all changes are committed and tests pass:
```bash
task release-check
```

This runs:
- Linting (fmt + vet)
- All unit tests
- Build verification

### 2. Test the Release Locally (Optional but Recommended)

Build a snapshot release to verify everything works:
```bash
task release-snapshot
```

Or test the full release process (without publishing):
```bash
task release-test
```

This creates all artifacts in the `dist/` directory for inspection.

### 3. Create and Push a Release Tag

When ready to release:

```bash
# Create an annotated tag (use monotonic versioning)
git tag -a v1 -m "Release v1"

# Push the tag to GitHub
git push origin v1
```

**Important:** Tags should start with 'v' followed by a number (v1, v2, v3, etc.).

### 4. GitHub Actions Takes Over

Once the tag is pushed:
1. GitHub Actions workflow triggers automatically
2. Runs tests and builds all artifacts
3. Creates a GitHub Release with:
   - Release notes (auto-generated from commits)
   - Binary archives for all platforms
   - RPM, DEB, and APK packages
   - Checksums file

### 5. Monitor the Release

Watch the release process:
- Go to: https://github.com/tomncooper/dunst-waybar/actions
- Click on the "Release" workflow run
- Monitor progress and check for errors

Once complete, the release appears at:
https://github.com/tomncooper/dunst-waybar/releases

## Release Versioning

This project uses **monotonic versioning** with simple incrementing numbers:
- **v1** - First release
- **v2** - Second release  
- **v3** - Third release
- etc.

This is simpler than semantic versioning and works perfectly with GoReleaser.

For pre-releases or testing, you can use suffixes:
- **v1-beta** - Beta version
- **v2-rc1** - Release candidate

## Release Versioning

This project uses **monotonic versioning** with simple incrementing numbers:
- **v1** - First release
- **v2** - Second release  
- **v3** - Third release
- etc.

This is simpler than semantic versioning and works perfectly with GoReleaser.

For pre-releases or testing, you can use suffixes:
- **v1-beta** - Beta version
- **v2-rc1** - Release candidate

## Changelog Generation

GoReleaser automatically generates changelogs from commit messages. For best results, use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat: Add new feature` → Features section
- `fix: Fix bug` → Bug fixes section
- `perf: Performance improvement` → Performance section
- `docs: Update documentation` → Excluded from changelog
- `test: Add tests` → Excluded from changelog
- `chore: Maintenance` → Excluded from changelog

Example workflow:
```bash
git commit -m "feat: Add notification grouping"
git commit -m "fix: Handle dunst restart gracefully"
git tag -a v2 -m "Release v2"
git push origin v2
```

## Manual Local Release (Advanced)

If you need to publish a release manually (not recommended):

```bash
# Set your GitHub token
export GITHUB_TOKEN="your-github-personal-access-token"

# Run the release
task release-local
```

Get a token from: https://github.com/settings/tokens (needs `repo` scope)

## Package Installation

After release, users can install via:

### Binary Installation
```bash
# Download from releases page
tar xzf dunst-waybar_1_Linux_x86_64.tar.gz
sudo install -Dm755 dunst-waybar /usr/local/bin/dunst-waybar
```

### RPM (Fedora/RHEL/CentOS)
```bash
sudo rpm -i dunst-waybar_1_linux_amd64.rpm
```

### DEB (Debian/Ubuntu)
```bash
sudo dpkg -i dunst-waybar_1_linux_amd64.deb
```

### APK (Alpine Linux)
```bash
sudo apk add --allow-untrusted dunst-waybar_1_linux_amd64.apk
```

## Troubleshooting

### Release Workflow Failed
1. Check the Actions tab for error details
2. Common issues:
   - Tests failed → Fix tests and push another tag
   - Build errors → Fix and push a new tag (bump patch version)
   - GITHUB_TOKEN permissions → Check repository settings

### Fix a Bad Release
If you need to fix a release:
1. Delete the tag locally: `git tag -d v2`
2. Delete the tag remotely: `git push origin :refs/tags/v2`
3. Delete the GitHub Release (if created)
4. Fix the issue
5. Create a new tag with the next version number

### Test Without Publishing
Always use `task release-test` before creating a real tag to catch issues early.

## Task Commands Reference

- `task release-check` - Run pre-release checks
- `task release-snapshot` - Build snapshot (no publish)
- `task release-test` - Full release test (no publish)
- `task release-local` - Manual release (requires GITHUB_TOKEN)
- `task release-clean` - Clean release artifacts

## GoReleaser Configuration

Configuration is in `.goreleaser.yaml`. Key sections:

- **builds** - Binary compilation settings
- **archives** - Archive creation (tar.gz)
- **nfpms** - Linux packages (RPM, DEB, APK)
- **checksum** - Checksum file generation
- **changelog** - Changelog generation rules
- **release** - GitHub Release settings

## References

- [GoReleaser Documentation](https://goreleaser.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Monotonic Versioning](https://en.wikipedia.org/wiki/Software_versioning#Incrementing_sequences)
