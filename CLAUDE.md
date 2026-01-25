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

### Supported Platforms

Release builds are available for:
- **macOS**: amd64, arm64
- **Linux**: 386, amd64, arm64, armv6 (Pi 1/Zero), armv7 (Pi 2/3/4 32-bit)
- **Windows**: 386, amd64

**Unsupported platforms** (due to `modernc.org/sqlite` libc constraints):
- FreeBSD, OpenBSD, NetBSD (all architectures)
- Windows ARM (arm, arm64)

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

### State Transitions

```
┌──────────┐  ffplay available   ┌─────────────┐
│ bootState│───────────────────>│ searchState │<──────────────────┐
└──────────┘                     └─────────────┘                   │
      │                                │                           │
      │ ffplay not available          │ enter search              │ quit/search key
      v                                v                           │
┌──────────┐                     ┌──────────────┐   error     ┌────────────┐
│errorState│<────────────────────│ loadingState │───────────>│ errorState │
└──────────┘                     └──────────────┘             └────────────┘
                                       │                           │
                                       │ success                   │ recoverable
                                       v                           │
                                 ┌───────────────┐                 │
                                 │ stationsState │─────────────────┘
                                 └───────────────┘
                                       │
                                       │ (any state)
                                       v
                                 ┌──────────────────────┐
                                 │ terminalTooSmallState│
                                 └──────────────────────┘
```

**Valid State Transitions:**
- `bootState` -> `searchState` (FFplay available) or `errorState` (not available)
- `searchState` -> `loadingState` (enter search) or `stationsState` (view bookmarks)
- `loadingState` -> `stationsState` (success) or `errorState` (failure)
- `stationsState` -> `searchState` (search key pressed)
- `errorState` -> `searchState` (recoverable error, user presses key)
- Any state -> `terminalTooSmallState` (terminal resized below minimum)

### Package Structure

```
radiogogo/
├── main.go           # Entry point: loads config, creates model, runs BubbleTea
├── api/              # RadioBrowser API client
│   ├── browser.go    # RadioBrowserService interface and implementation
│   ├── http.go       # HTTPClientService interface
│   └── utils.go      # String conversion helpers
├── common/           # Shared data models
│   ├── station.go    # Station struct with all RadioBrowser fields
│   ├── station_query.go  # StationQuery enum with 14 filter types
│   ├── url.go        # RadioGoGoURL custom URL type
│   ├── click_station_response.go
│   └── vote_station_response.go
├── config/           # Configuration management
│   ├── config.go     # Config struct (Language, Theme, Keybindings, PlayerPreferences)
│   ├── keybindings.go    # Keybindings struct with validation
│   └── paths.go      # Platform-specific config paths
├── data/             # Application metadata
│   └── version.go    # Version string (injected via ldflags)
├── i18n/             # Internationalization
│   ├── i18n.go       # T(), Tf(), Tn(), Tfn() functions
│   └── locales/*.yaml    # Translation files (9 languages)
├── models/           # TUI components and state machine
│   ├── model.go          # Root Model coordinating all views
│   ├── model_handlers.go # Message handling (global, state transitions, delegation)
│   ├── search.go         # SearchModel - search form
│   ├── loading.go        # LoadingModel - spinner while fetching
│   ├── stations.go       # StationsModel - results table
│   ├── stations_handlers.go  # StationsModel message handlers
│   ├── stations_commands.go  # StationsModel tea.Cmd functions
│   ├── stations_modal.go     # Hidden stations modal
│   ├── error_fatal.go    # ErrorModel - error display
│   ├── header.go         # HeaderModel - playback/recording indicators
│   ├── selector.go       # Generic selector component
│   ├── theme.go          # Theme struct with lipgloss styles
│   └── layout.go         # Height calculations and filler rendering
├── playback/         # Audio playback and recording
│   ├── manager.go    # PlaybackManagerService interface
│   ├── ffplay.go     # FFPlayPlaybackManager implementation
│   └── filename.go   # Recording filename generation
├── storage/          # Persistent storage
│   ├── storage.go    # StationStorageService interface
│   └── sqlite_storage.go  # SQLite implementation with caching
└── mocks/            # Test mocks
    ├── browser_mock.go
    ├── playback_manager_mock.go
    ├── storage_mock.go
    └── http_client_mock.go
```

### Package Dependency Graph

```
                    ┌──────────┐
                    │   main   │
                    └────┬─────┘
          ┌──────────────┼──────────────┐
          v              v              v
     ┌────────┐    ┌──────────┐    ┌────────┐
     │ config │    │  models  │    │  i18n  │
     └────┬───┘    └────┬─────┘    └────────┘
          │             │                ↑
          │    ┌────────┼────────┬───────┤
          │    v        v        v       │
          │  ┌─────┐ ┌────────┐ ┌─────────┐
          │  │ api │ │playback│ │ storage │
          │  └──┬──┘ └────┬───┘ └────┬────┘
          │     │         │          │
          └─────┴────┬────┴──────────┘
                     v
               ┌──────────┐
               │  common  │
               └──────────┘
               ┌──────────┐
               │   data   │ (version info, used by api)
               └──────────┘
```

**Dependency Rules:**
- `common` has no internal dependencies (only stdlib + google/uuid)
- `data` has no internal dependencies
- `i18n` depends only on external packages
- `config` depends only on stdlib
- `api`, `playback`, `storage` depend on `common`, `config`, `i18n`
- `models` depends on all packages except `main`
- `main` orchestrates everything

## Interfaces

### Core Service Interfaces

**RadioBrowserService** (`api/browser.go`) - API client:
```go
type RadioBrowserService interface {
    GetStations(stationQuery StationQuery, searchTerm, order string,
                reverse bool, offset, limit uint64, hideBroken bool) ([]Station, error)
    GetStationsByUUIDs(uuids []uuid.UUID) ([]Station, error)
    ClickStation(station Station) (ClickStationResponse, error)
    VoteStation(station Station) (VoteStationResponse, error)
}
```

**PlaybackManagerService** (`playback/manager.go`) - Audio playback:
```go
type PlaybackManagerService interface {
    Name() string
    IsAvailable() bool
    NotAvailableErrorString() string
    IsPlaying() bool
    PlayStation(station Station, volume int) error
    StopStation() error
    VolumeMin() int
    VolumeDefault() int  // Returns 80 for FFplay
    VolumeMax() int
    VolumeIsPercentage() bool
    CurrentStation() Station
    IsRecordingAvailable() bool
    RecordingNotAvailableErrorString() string
    IsRecording() bool
    StartRecording(outputPath string) error
    StopRecording() (string, error)
    CurrentRecordingPath() string
}
```

**StationStorageService** (`storage/storage.go`) - Persistence:
```go
type StationStorageService interface {
    GetBookmarks() ([]uuid.UUID, error)
    AddBookmark(stationUUID uuid.UUID) error
    RemoveBookmark(stationUUID uuid.UUID) error
    IsBookmarked(stationUUID uuid.UUID) bool

    GetHidden() ([]uuid.UUID, error)
    AddHidden(stationUUID uuid.UUID) error
    RemoveHidden(stationUUID uuid.UUID) error
    IsHidden(stationUUID uuid.UUID) bool

    GetLastVoteTimestamp() (time.Time, bool)
    SetLastVoteTimestamp(timestamp time.Time) error
}
```

**HTTPClientService** (`api/http.go`) - HTTP client (for testing):
```go
type HTTPClientService interface {
    Do(req *http.Request) (*http.Response, error)
}
```

### Playback Abstraction Interfaces

```go
// CommandExecutor - allows mocking exec.Command
type CommandExecutor interface {
    Command(name string, args ...string) Cmd
    LookPath(file string) (string, error)
}

// Cmd - wraps exec.Cmd
type Cmd interface {
    Start() error
    Run() error
    Process() Process
    SetStderr(w *os.File)
    SetStdout(w *os.File)
}

// Process - wraps os.Process
type Process interface {
    Kill() error
    Signal(sig os.Signal) error
    Wait() (*os.ProcessState, error)
    Pid() int
}
```

## Message Types

### State Transition Messages (models/model.go)

```go
switchToErrorModelMsg{err string, recoverable bool}  // -> errorState
switchToSearchModelMsg{}                              // -> searchState
switchToLoadingModelMsg{query, queryText}            // -> loadingState
switchToStationsModelMsg{stations, query, queryText} // -> stationsState
switchToBookmarksMsg{stations}                        // -> stationsState (bookmark view)
```

### UI Messages

```go
bottomBarUpdateMsg{commands, secondaryCommands []string}  // Update bottom bar
languageChangedMsg{lang string}                           // Language change
quitMsg{}                                                 // Trigger quit
```

### Playback Messages (models/stations_handlers.go)

```go
playbackStatusMsg{station Station, status PlaybackStatus}  // Playing/stopped
recordingStatusMsg{isRecording bool}                       // Recording state
stationCursorMovedMsg{offset, totalStations int}          // Table cursor moved
```

### Async Operation Messages

```go
stationsLoadedMsg{stations []Station, err error}          // API response
bookmarkToggledMsg{stationUUID uuid.UUID, added bool}    // Bookmark changed
stationHiddenMsg{stationUUID uuid.UUID}                  // Station hidden
stationUnhiddenMsg{stationUUID uuid.UUID}                // Station unhidden
voteResultMsg{station Station, success bool, message string}
volumeDebounceExpiredMsg{changeID int64}                 // Volume change debounce
```

## SQLite Storage Schema

**Database location:** `~/.config/radiogogo/radiogogo.db` (or `%LOCALAPPDATA%\radiogogo\` on Windows)

```sql
-- Schema version tracking (current: 3)
CREATE TABLE schema_version (
    version INTEGER PRIMARY KEY
);

-- Bookmarked stations
CREATE TABLE bookmarks (
    station_uuid TEXT PRIMARY KEY,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Hidden stations (filtered from search results)
CREATE TABLE hidden (
    station_uuid TEXT PRIMARY KEY,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Global vote cooldown (RadioBrowser enforces 10-min per IP)
CREATE TABLE last_vote (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    voted_at TEXT NOT NULL
);
```

**Storage Features:**
- WAL mode enabled for better concurrent access
- In-memory caching with RWMutex for fast reads
- Automatic database integrity validation on startup
- Automatic recovery from corrupted databases (renamed to `.corrupted.<timestamp>`)
- Schema migrations handled transparently

## Key Patterns

### Message Passing (BubbleTea Elm Architecture)

```go
// State transitions via typed messages
case switchToSearchModelMsg:
    m.state = searchState
    return m, m.searchModel.Init()

// Commands return tea.Cmd (functions that produce messages asynchronously)
func fetchStationsCmd(browser RadioBrowserService, query StationQuery) tea.Cmd {
    return func() tea.Msg {
        stations, err := browser.GetStations(query, ...)
        return stationsLoadedMsg{stations: stations, err: err}
    }
}

// Use tea.Batch() for parallel commands
return m, tea.Batch(cmd1, cmd2, cmd3)

// Use tea.Sequence() for ordered execution
return m, tea.Sequence(cmd1, cmd2)
```

### Handler Pattern (models/model_handlers.go)

The root Model delegates message handling through a chain:
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 1. Handle global messages (cursor, playback, window resize, quit)
    if handled, newM, cmd := m.handleGlobalMessages(msg); handled {
        return newM, cmd
    }

    // 2. Handle state transitions
    if handled, newM, cmd := m.handleStateTransitions(msg); handled {
        return newM, cmd
    }

    // 3. Delegate to current state's model
    return m.delegateToCurrentState(msg)
}
```

### Dependency Injection for Testing

```go
// Production: uses real dependencies
model, err := models.NewDefaultModel(cfg)

// Testing: inject mocks
model := models.NewModel(cfg, mockBrowser, mockPlayback, mockStorage)
```

### Theme-based Styling

All styles defined in `models/theme.go` via the `Theme` struct:
```go
type Theme struct {
    PrimaryBlock   lipgloss.Style  // Colored background blocks
    SecondaryBlock lipgloss.Style
    Text           lipgloss.Style  // Default text
    PrimaryText    lipgloss.Style  // Purple accent text
    SecondaryText  lipgloss.Style  // Light purple text
    TertiaryText   lipgloss.Style  // Gray text
    ErrorText      lipgloss.Style  // Red text
    SuccessText    lipgloss.Style  // Green text (#50FA7B)
    StationsTableStyle table.Styles
    ModalStyle     lipgloss.Style
    QualityHighStyle   lipgloss.Style  // Green (#50FA7B) for 256+ kbps
    QualityMediumStyle lipgloss.Style  // Orange (#FFB86C) for 128-192 kbps
    QualityLowStyle    lipgloss.Style  // Gray for <128 kbps
    StatusBoxStyle     lipgloss.Style  // Rounded border box for now-playing
    SecondaryColor string              // For dynamic border colors
}
```

**Usage:** `theme.PrimaryText.Render("text")` - never inline `lipgloss.NewStyle()`.

### Stations Table Visual Elements

The stations table (`models/stations.go`) displays these columns:

| Column | Width | Description |
|--------|-------|-------------|
| Name | 35 | Station name with `▶` (now playing) and `⭐` (bookmarked) prefixes |
| Country | 10 | ISO 3166-1 alpha-2 country code |
| Quality | 12 | Bitrate + codec + star tier (e.g., "320k MP3 ★★", "128k AAC ★", "64k OGG") |
| Clicks | 10 | Listener count (formatted: 45234 → "45.2K") |
| Votes | 8 | Vote count (formatted with K/M suffixes) |
| Status | 6 | Online status: "✓" (OK) or "✗" (broken) |

**Quality Tier Stars:**
| Bitrate | Indicator | Example |
|---------|-----------|---------|
| 256+ kbps | ★★ (high) | `320k MP3 ★★` |
| 128-255 kbps | ★ (medium) | `128k AAC ★` |
| <128 kbps | (none) | `64k OGG` |
| No info | — | `—` |

> **Note:** Stars are used instead of colored text because the bubbles/table component's Cell style overrides pre-rendered ANSI color codes, making per-cell coloring impossible.

**Helper functions (`models/stations.go`):**
- `formatNumber(n uint64)` - Formats large numbers: 1000+ → "1.0K", 1000000+ → "1.0M"
- `removeStationByUUID(stations, uuid)` - Returns new slice with station removed
- `setCursorSafely(cursor int)` - Bounds-checked cursor setting on StationsModel
- `rebuildTablePreservingCursor(cursorOverride int)` - Rebuilds table, restores cursor position

**Helper functions (`models/stations_commands.go`):**
- `clearErrorAfterDelayCmd()` - Returns tea.Cmd to clear errors after 3 seconds

**Now-Playing Box** (`renderNowPlayingBox`):
When a station is playing, the status bar shows a multi-line bordered box:
```
╭──────────────────────────────────────────────────────────────────╮
│ ▶ Station Name • 128 kbps MP3 • 🎧 45.2K listeners               │
│ 📍 US • jazz, smooth, relaxing • ⭐ Bookmarked                   │
╰──────────────────────────────────────────────────────────────────╯
```

### Platform-specific Code

FFplay/FFmpeg process management differs between Windows and Unix:

```go
// Windows: taskkill /T (tree kill) /F (force) to kill process tree
killCmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", pid))

// Unix/macOS: SIGKILL for stop, SIGINT for graceful recording stop
process.Kill()           // SIGKILL - immediate termination
process.Signal(os.Interrupt)  // SIGINT - graceful (ffmpeg finalizes file)
```

## Configuration

### Config File Location
- **Windows:** `%LOCALAPPDATA%\radiogogo\config.yaml`
- **Others:** `~/.config/radiogogo/config.yaml`

### Full Config Structure
```yaml
language: en  # de, el, en, es, it, ja, pt, ru, zh

theme:
  textColor: "#ffffff"
  primaryColor: "#5a4f9f"
  secondaryColor: "#8b77db"
  tertiaryColor: "#4e4e4e"
  errorColor: "#ff0000"

keybindings:
  quit: "q"
  search: "s"
  record: "r"
  bookmarkToggle: "b"
  bookmarksView: "B"
  hideStation: "h"
  manageHidden: "H"
  changeLanguage: "L"
  volumeDown: "9"
  volumeUp: "0"
  navigateDown: "j"
  navigateUp: "k"
  stopPlayback: "ctrl+k"
  vote: "v"

playerPreferences:
  defaultVolume: 80  # Initial volume level (0-100)
```

### Reserved Keys (cannot be remapped)

```
Navigation: up, down, left, right, tab, enter, esc
Editing: backspace, delete, pgup, pgdown, home, end
System: ctrl+c, ctrl+z, ctrl+s, ctrl+q, ctrl+l
TextInput: ctrl+a, ctrl+e, ctrl+u, ctrl+w, ctrl+d, ctrl+h
```

### Keybinding Validation

Validation happens in `main.go` at startup:
- Reserved keys are rejected with a warning
- Duplicate keys are rejected with a warning
- Invalid keys fall back to defaults
- Empty keys are filled with defaults

### Player Preferences

The `playerPreferences` section allows customizing playback behavior:

- **defaultVolume**: Initial volume level when starting the application (0-100, default: 80)

Values are validated and clamped to valid ranges. Missing preferences use sensible defaults.

## Internationalization (i18n)

### Available Languages
de (German), el (Greek), en (English), es (Spanish), it (Italian), ja (Japanese), pt (Portuguese), ru (Russian), zh (Chinese)

### Usage Functions
```go
i18n.T("message_id")                           // Simple translation
i18n.Tf("message_id", map[string]interface{}{  // With template data
    "Key": value,
})
i18n.Tn("message_id", count)                   // Pluralization
i18n.Tfn("message_id", count, data)            // Both
```

### Adding a New Language
1. Create `i18n/locales/XX.yaml` (copy from en.yaml)
2. Translate all message strings
3. App auto-discovers new locale files (embedded via `//go:embed`)

### Runtime Language Switching
Press "L" on search screen -> selector shows available languages -> selection saved to config.

## Error Handling

### Fatal vs Non-Fatal Errors

| Error Type | Example | Behavior |
|------------|---------|----------|
| Fatal | FFplay not installed | `errorState` with `recoverable: false`, app unusable |
| Non-fatal | API timeout, playback failure | Display in current view, auto-clear after 3 seconds |

### Non-Fatal Error Flow
```go
case playbackError:
    m.err = msg.err.Error()
    return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
        return clearErrorMsg{}
    })
```

### API Error Handling
- HTTP 4xx/5xx: `fmt.Errorf("API request failed with status %d", statusCode)`
- Network errors: Propagate underlying error
- JSON parsing errors: Returned as-is

## Testing

### Mock Pattern
Mocks use function fields for flexible behavior configuration:
```go
mockBrowser := &mocks.MockRadioBrowserService{
    GetStationsFunc: func(query StationQuery, ...) ([]Station, error) {
        return []Station{{Name: "Test Radio"}}, nil
    },
    ClickStationFunc: func(station Station) (ClickStationResponse, error) {
        return ClickStationResponse{Ok: true}, nil
    },
}
```

### Test Patterns
```go
// Table-driven tests
func TestStationQuery(t *testing.T) {
    tests := []struct {
        name     string
        query    StationQuery
        expected string
    }{
        {"by name", StationQueryByName, "byname"},
        {"by country", StationQueryByCountry, "bycountry"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, string(tt.query))
        })
    }
}
```

### Test File Organization
- Tests live alongside source: `foo.go` + `foo_test.go`
- Mocks in dedicated `mocks/` package
- Use `github.com/stretchr/testify/assert` for assertions

## Common Developer Pitfalls

1. **Inline lipgloss styles** - Always use `theme.go` styles. Never write `lipgloss.NewStyle()` in view code.

2. **Blocking in Update()** - Commands run synchronously. For async operations, return a `tea.Cmd`:
   ```go
   // WRONG: blocks the UI
   func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       stations, _ := browser.GetStations(...) // Blocking!
   }

   // CORRECT: async via command
   return m, func() tea.Msg {
       stations, err := browser.GetStations(...)
       return stationsLoadedMsg{stations: stations, err: err}
   }
   ```

3. **Type assertions without guards** - Child model Update() returns `tea.Model`. Always type-assert safely:
   ```go
   newModel, cmd := m.childModel.Update(msg)
   m.childModel = newModel.(ChildModel) // Safe if Update always returns same type
   ```

4. **Forgetting to update child dimensions** - When handling `tea.WindowSizeMsg`, update ALL child models' dimensions.

5. **Not stopping playback on state transitions** - Always call `m.playbackManager.StopStation()` when transitioning away from stations view.

6. **Modifying config without saving** - Config changes must be persisted:
   ```go
   m.config.Language = newLang
   _ = m.config.Save(config.ConfigFile())
   ```

7. **Ignoring storage errors** - Storage operations can fail; handle errors appropriately.

8. **Volume change without debouncing** - Rapid volume changes should be debounced to avoid excessive process restarts.

9. **Forgetting to filter hidden stations** - Always filter via `filterHiddenStations()` before displaying search results.

10. **Using fmt.Sprintf for conversions** - Use `strconv.FormatUint()` / `strconv.FormatBool()` instead.

11. **Forgetting layout recalculation** - When status bar height changes (e.g., playback start/stop changes from single line to multi-line box), call `updateTableDimensions()` to recalculate the table height. Also rebuild the table with `newStationsTableModel()` to update visual indicators like the `▶` now-playing prefix.

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework (Elm architecture) |
| `github.com/charmbracelet/bubbles` | TUI components (spinner, table, textinput) |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `gopkg.in/yaml.v3` | YAML config parsing |
| `github.com/stretchr/testify` | Testing assertions |
| `github.com/google/uuid` | UUID handling for station IDs |
| `github.com/nicksnyder/go-i18n/v2` | Internationalization and pluralization |
| `modernc.org/sqlite` | Pure Go SQLite driver (no CGO required) |
| `golang.org/x/text/language` | Language tag parsing for i18n |

## Code Style Guidelines

- Use `strconv.FormatUint()` / `strconv.FormatBool()` instead of `fmt.Sprintf()` for conversions
- Centralize all lipgloss styles in `theme.go` - avoid inline style definitions
- Add detailed comments for platform-specific logic (Windows vs Unix)
- Keep files under ~500 lines; split large files by responsibility (see `stations_*.go`)
- Document public interfaces and complex functions with godoc comments
- Use consistent handler pattern: `func (m Model) handleXxx(msg) (bool, Model, tea.Cmd)`

## Version Handling

- Version defined as `var` in `data/version.go` for ldflags injection
- Local builds show "dev", release builds show actual version
- Release script injects version: `-ldflags="-s -w -X github.com/zi0p4tch0/radiogogo/data.Version=$1"`

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

### Guidelines for Release Notes Content
- Only list features/fixes that existed before this release cycle
- Don't list bugs that were introduced and fixed within the same release
- If a feature was removed and reintroduced, don't list it as "new"
- Focus on what users experience, not implementation details (e.g., "Bookmarks" not "SQLite storage backend")
