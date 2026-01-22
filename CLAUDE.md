# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RadioGoGo is a terminal UI application written in Go that lets users search, browse, and play radio stations from the RadioBrowser API. It uses BubbleTea for the TUI framework and FFplay for audio playback.

## Build & Development Commands

```bash
# Build
go build -o radiogogo

# Install
go install github.com/zi0p4tch0/radiogogo@latest

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./api
go test ./models
go test ./config

# Run a single test
go test -run TestBrowserImplGetStations ./api

# Format code
go fmt ./...

# Vet code
go vet ./...

# Multi-platform release build
./make_release.sh <version>
```

### Nix Development Environment

```bash
nix flake update   # Update dependencies
direnv allow       # Enable automatic environment loading
```

The flake provides: Go compiler, Delve debugger, Gopls, Go tools, and FFmpeg.

## Architecture

RadioGoGo is a **state machine TUI application** with these states:

1. **bootState** - Initialization, checks if playback (FFplay) is available
2. **searchState** - User enters search criteria (name, country, codec, etc.)
3. **loadingState** - Fetches stations from RadioBrowser API
4. **stationsState** - Displays results in a table, allows selection and playback
5. **errorState** - Shows error messages

### Package Structure

- **`main.go`** - Entry point: loads config, creates model, runs BubbleTea program
- **`models/`** - TUI components and main state machine (`model.go` is the core)
- **`api/`** - RadioBrowser API client with DNS-based load balancing
- **`config/`** - YAML configuration management with platform-specific paths
- **`common/`** - Shared data models (Station, StationQuery, URL types)
- **`playback/`** - Audio playback via FFplay
- **`assets/`** - ASCII art logo and static text
- **`mocks/`** - Test mocks for interfaces

### Key Patterns

- **Message Passing:** BubbleTea's Elm-inspired architecture with Msg types for events
- **Interfaces:** API uses mocking-friendly interfaces (`RadioBrowserService`, `PlaybackManagerService`)
- **Configuration:** YAML-based with platform-aware paths:
  - Windows: `%LOCALAPPDATA%\radiogogo\config.yaml`
  - Others: `~/.config/radiogogo/config.yaml`

### Key Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components (spinner, table, textinput)
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `gopkg.in/yaml.v3` - YAML config parsing
- `github.com/stretchr/testify` - Testing assertions

## Release Notes Format

When writing GitHub release notes, use this emoji-based format:

### Emoji Categories
- 🌟 Features and updates
- 🎵 Audio-related features
- 🔝 UI improvements
- 🐞 Bug fixes
- 🎨 Customization features
- 🪟 Platform-specific fixes

### Structure
1. Version header with descriptive subtitle (e.g., "v0.3.3 - Cleanup build")
2. Bulleted list with emoji prefixes
3. Nested sub-bullets for details
4. SHA256 checksums section at the end

### Example
```
## v0.3.0 - New Features

🌟 Bumped up Go version to 1.22
🌟 Updated dependencies to their latest versions
🎵 MPV Support: Users now have more choice for audio playback
🐞 Fixed volume slider not responding on certain terminals
```
