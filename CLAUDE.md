# CLAUDE.md

Terminal UI application for searching, browsing, and playing internet radio stations.

**Stack:** Go + BubbleTea (Elm architecture) + FFplay (playback) + SQLite (storage)

## Commands

```bash
go build -o radiogogo    # Build
go test ./...            # Test all
go fmt ./...             # Format
go vet ./...             # Lint
./make_release.sh v1.0   # Release build
```

**Nix users:** `direnv allow` to load the development environment.

## Architecture

RadioGoGo is a **state machine TUI**:

```
bootState → searchState → loadingState → stationsState
                ↓              ↓              ↓
            bookmarks      errorState    searchState
```

State transitions are handled via typed messages in the BubbleTea Elm architecture.

### Package Structure

```
radiogogo/
├── main.go       # Entry point
├── api/          # RadioBrowser API client
├── common/       # Shared models (Station, StationQuery)
├── config/       # Config management
├── i18n/         # Internationalization (9 languages)
├── models/       # TUI components and state machine
├── playback/     # FFplay audio playback
├── storage/      # SQLite persistence
└── mocks/        # Test mocks
```

### Key Files by Task

| Task | Start Here |
|------|------------|
| Add keyboard shortcut | `config/keybindings.go`, `models/stations_handlers.go` |
| New API endpoint | `api/browser.go` |
| UI component | `models/` (see existing `*Model` types) |
| Styling | `models/theme.go` |
| Playback changes | `playback/ffplay.go` |
| Storage/bookmarks | `storage/sqlite_storage.go` |
| Add translation | `i18n/locales/XX.yaml` |

## Critical Rules

1. **Never block in Update()** - Return `tea.Cmd` for async operations
2. **Use theme styles** - Never inline `lipgloss.NewStyle()`, use `models/theme.go`
3. **Filter hidden stations** - Call `filterHiddenStations()` before displaying results
4. **Recalculate layout** - Call `updateTableDimensions()` after playback state changes

## Detailed Documentation

Read these files based on task relevance:

| File | Contents |
|------|----------|
| `.claude/docs/architecture.md` | State machine, packages, interfaces, SQLite schema |
| `.claude/docs/patterns.md` | BubbleTea patterns, message types, theme styling |
| `.claude/docs/testing.md` | Mocks, dependency injection, test organization |
| `.claude/docs/pitfalls.md` | Common mistakes with solutions |
| `.claude/docs/config.md` | Configuration structure, keybindings, i18n |
| `.claude/docs/releases.md` | Release process, platform support |
