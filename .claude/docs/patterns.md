# Patterns

## BubbleTea Elm Architecture

All UI updates flow through the message-passing pattern:

```go
// Commands produce messages asynchronously
func fetchStationsCmd(browser RadioBrowserService) tea.Cmd {
    return func() tea.Msg {
        stations, err := browser.GetStations(...)
        return stationsLoadedMsg{stations: stations, err: err}
    }
}

// Parallel execution
return m, tea.Batch(cmd1, cmd2, cmd3)

// Sequential execution
return m, tea.Sequence(cmd1, cmd2)
```

**Key rule:** Never block in `Update()`. All I/O must return a `tea.Cmd`.

## Handler Chain

The root Model delegates messages through three stages:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 1. Global messages (window resize, quit, playback status)
    if handled, newM, cmd := m.handleGlobalMessages(msg); handled { return newM, cmd }

    // 2. State transitions (switchToSearchModelMsg, etc.)
    if handled, newM, cmd := m.handleStateTransitions(msg); handled { return newM, cmd }

    // 3. Delegate to current state's model
    return m.delegateToCurrentState(msg)
}
```

**File:** `models/model_handlers.go`

## Message Types

**State transitions** (`models/model.go`):
- `switchToSearchModelMsg{}` - Return to search
- `switchToLoadingModelMsg{query, queryText}` - Start loading
- `switchToStationsModelMsg{stations, query, queryText}` - Show results
- `switchToErrorModelMsg{err, recoverable}` - Show error

**Async results** (`models/stations_handlers.go`):
- `stationsLoadedMsg{stations, err}` - API response
- `playbackStatusMsg{station, status}` - Playback state
- `bookmarkToggledMsg{stationUUID, added}` - Bookmark change

## Theme-based Styling

All styles live in `models/theme.go` via the `Theme` struct.

**Usage:** `theme.PrimaryText.Render("text")`

**Never:** Inline `lipgloss.NewStyle()` in view code.

**Key styles:**
- `PrimaryText`, `SecondaryText`, `TertiaryText` - Text variants
- `ErrorText`, `SuccessText` - Status colors
- `StationsTableStyle` - Table styling
- `QualityHighStyle`, `QualityMediumStyle`, `QualityLowStyle` - Bitrate indicators

## Error Handling

| Type | Example | Behavior |
|------|---------|----------|
| Fatal | FFplay not installed | `errorState` with `recoverable: false` |
| Non-fatal | API timeout | Display error, auto-clear after 3 seconds |

**Non-fatal pattern:**
```go
m.err = msg.err.Error()
return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
    return clearErrorMsg{}
})
```

## Stations Table

**Columns:** Name (35), Country (10), Quality (12), Clicks (10), Votes (8), Status (6)

**Quality tiers:** 256+ kbps = `★★`, 128-255 = `★`, <128 = (none)

**Now-playing indicator:** `▶` prefix on station name

**Helper functions:**
- `formatNumber()` - 1000 → "1.0K"
- `removeStationByUUID()` - Filter slice
- `rebuildTablePreservingCursor()` - Refresh without losing position
