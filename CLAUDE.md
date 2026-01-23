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
6. **terminalTooSmallState** - Displays when terminal is below minimum size (115x29)

### Package Structure

- **`main.go`** - Entry point: loads config, creates model, runs BubbleTea program
- **`models/`** - TUI components and main state machine:
  - `model.go` - Root state machine coordinating all views
  - `stations.go`, `stations_commands.go`, `stations_modal.go` - Stations view (split for maintainability)
  - `search.go`, `loading.go`, `error.go` - Individual view models
  - `header.go` - Header bar with playback/recording indicators
  - `theme.go` - Centralized lipgloss styling configuration
  - `layout.go` - Terminal height calculations
  - `selector.go` - Generic selector component
- **`api/`** - RadioBrowser API client with DNS-based load balancing
- **`config/`** - YAML configuration management with platform-specific paths
- **`common/`** - Shared data models (Station, StationQuery, URL types)
- **`playback/`** - Audio playback via FFplay and recording via FFmpeg
- **`storage/`** - Persistent storage for bookmarks and hidden stations
- **`data/`** - Version information and user agent string
- **`i18n/`** - Internationalization support using go-i18n:
  - `i18n.go` - Core functions: T(), Tf(), Tn(), SetLanguage()
  - `locales/*.yaml` - Translation files (en.yaml, it.yaml)
- **`mocks/`** - Test mocks for interfaces

### Key Patterns

**Message Passing (BubbleTea Elm Architecture):**
- State transitions via typed messages: `switchToSearchModelMsg`, `switchToStationsModelMsg`, etc.
- Commands return `tea.Cmd` (functions that produce messages asynchronously)
- Use `tea.Batch()` for parallel commands, `tea.Sequence()` for ordered execution

**Interfaces for Testability:**
- `RadioBrowserService` - API client interface
- `PlaybackManagerService` - Audio playback interface
- `StationStorageService` - Persistence interface
- `HTTPClientService` - HTTP client interface

**Theme-based Styling:**
- All styles defined in `models/theme.go` via the `Theme` struct
- Use `theme.PrimaryText`, `theme.ErrorText`, etc. - never inline `lipgloss.NewStyle()`

**Platform-specific Code:**
- FFplay/FFmpeg process management differs between Windows and Unix
- Windows uses `taskkill /T /F` for process tree termination
- Unix uses signals (SIGKILL for stop, SIGINT for graceful recording stop)
- See `playback/ffplay.go` for implementation details

**Configuration:**
- YAML-based with platform-aware paths:
  - Windows: `%LOCALAPPDATA%\radiogogo\config.yaml`
  - Others: `~/.config/radiogogo/config.yaml`
- `language` field controls UI language (default: "en", available: "en", "it")

**Internationalization (i18n):**
- All user-facing strings use `i18n.T("message_id")` or `i18n.Tf()` for templates
- Locale files in `i18n/locales/*.yaml` using go-i18n format
- Language persisted in config, switchable at runtime with "L" key on search screen
- To add a language: create `i18n/locales/XX.yaml`, app auto-discovers it

**Version Handling:**
- Version defined as `var` in `data/version.go` for ldflags injection
- Local builds show "dev", release builds show actual version
- Release script injects version: `-ldflags="-s -w -X github.com/zi0p4tch0/radiogogo/data.Version=$1"`

### Key Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components (spinner, table, textinput)
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `gopkg.in/yaml.v3` - YAML config parsing
- `github.com/stretchr/testify` - Testing assertions
- `github.com/google/uuid` - UUID handling for station IDs
- `github.com/nicksnyder/go-i18n/v2` - Internationalization and pluralization

## Testing

### Mock Usage
Mocks are in the `mocks/` package with function-based configuration:
```go
mockPM := &mocks.MockPlaybackManagerService{
    NameResult: "ffplay",
    IsPlayingResult: true,
    PlayStationFunc: func(station common.Station, volume int) error {
        return nil
    },
}
```

### Test Patterns
- Table-driven tests with `t.Run()` subtests
- Use `github.com/stretchr/testify/assert` for assertions
- Tests live alongside source files (`*_test.go`)

## Code Style Guidelines

- Use `strconv.FormatUint()` / `strconv.FormatBool()` instead of `fmt.Sprintf()` for conversions
- Centralize all lipgloss styles in `theme.go` - avoid inline style definitions
- Add detailed comments for platform-specific logic (Windows vs Unix)
- Keep files under ~500 lines; split large files by responsibility (see `stations_*.go`)
- Document public interfaces and complex functions with godoc comments

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
