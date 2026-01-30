# Architecture

RadioGoGo is a **state machine TUI application** built with BubbleTea's Elm architecture.

## State Machine

```
bootState ──> searchState ──> loadingState ──> stationsState
    │              │                │               │
    v              v                v               v
errorState    (bookmarks)      errorState     searchState
```

**States:**
- `bootState` - Initialization, checks FFplay availability
- `searchState` - Search form (name, country, codec)
- `loadingState` - Fetching from RadioBrowser API
- `stationsState` - Results table with playback controls
- `errorState` - Error display (fatal vs recoverable)
- `terminalTooSmallState` - Below minimum size (115x29)

**Transitions defined in:** `models/model_handlers.go:handleStateTransitions()`

## Package Structure

```
radiogogo/
├── main.go           # Entry point, config loading, BubbleTea setup
├── api/              # RadioBrowser API client
├── common/           # Shared data models (Station, StationQuery)
├── config/           # Configuration management
├── data/             # Version info (ldflags injection)
├── i18n/             # Internationalization (9 languages)
├── models/           # TUI components and state machine
├── playback/         # FFplay audio playback and recording
├── storage/          # SQLite persistence (bookmarks, hidden stations)
└── mocks/            # Test mocks
```

## Key Interfaces

| Interface | File | Purpose |
|-----------|------|---------|
| `RadioBrowserService` | `api/browser.go:15` | API client for station search/vote/click |
| `PlaybackManagerService` | `playback/manager.go:10` | Audio playback and recording |
| `StationStorageService` | `storage/storage.go:12` | Bookmarks and hidden stations |
| `HTTPClientService` | `api/http.go:10` | HTTP client abstraction for testing |

## Package Dependencies

```
main ──> config, models, i18n
models ──> api, playback, storage, common, config, i18n
api, playback, storage ──> common, config, i18n
common ──> (stdlib only)
```

**Rules:**
- `common` has no internal dependencies
- `models` coordinates all other packages
- Circular dependencies are prohibited

## SQLite Storage

**Location:** `~/.config/radiogogo/radiogogo.db` (Windows: `%LOCALAPPDATA%\radiogogo\`)

**Tables:**
- `bookmarks` - Station UUIDs with timestamps
- `hidden` - Stations filtered from search results
- `last_vote` - Global vote cooldown (10min per RadioBrowser rules)
- `schema_version` - Migration tracking (current: v3)

**Schema details:** `storage/sqlite_storage.go:30-60`

## Platform Differences

FFplay process termination differs by OS:
- **Windows:** `taskkill /T /F /PID` (tree kill)
- **Unix/macOS:** `SIGKILL` for stop, `SIGINT` for graceful recording end

**Implementation:** `playback/ffplay.go:StopStation()`, `playback/ffplay.go:StopRecording()`
