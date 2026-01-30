# Common Pitfalls

## 1. Blocking in Update()

**Wrong:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    stations, _ := browser.GetStations(...) // Blocks UI!
}
```

**Right:**
```go
return m, func() tea.Msg {
    stations, err := browser.GetStations(...)
    return stationsLoadedMsg{stations: stations, err: err}
}
```

## 2. Inline lipgloss Styles

**Wrong:** `lipgloss.NewStyle().Foreground(...)` in view code

**Right:** `theme.PrimaryText.Render("text")` using `models/theme.go`

## 3. Forgetting Window Resize Updates

When handling `tea.WindowSizeMsg`, update ALL child models' dimensions.

**File:** `models/model_handlers.go:handleGlobalMessages()`

## 4. Not Stopping Playback on State Transitions

Always call `m.playbackManager.StopStation()` when leaving stations view.

## 5. Config Changes Without Saving

```go
m.config.Language = newLang
_ = m.config.Save(config.ConfigFile())  // Don't forget this!
```

## 6. Volume Changes Without Debouncing

Rapid volume changes restart FFplay repeatedly. Use debouncing.

**Implementation:** `volumeDebounceExpiredMsg` in `models/stations_handlers.go`

## 7. Forgetting to Filter Hidden Stations

Always filter via `filterHiddenStations()` before displaying search results.

## 8. Layout Recalculation After Playback Changes

When playback starts/stops, the status bar height changes (single line vs multi-line now-playing box).

**Required:**
1. Call `updateTableDimensions()` to recalculate table height
2. Rebuild table with `newStationsTableModel()` to update `▶` prefix

## 9. Type Assertions Without Guards

Child model `Update()` returns `tea.Model`. Type-assert safely:

```go
newModel, cmd := m.childModel.Update(msg)
m.childModel = newModel.(ChildModel)  // Safe if Update always returns same type
```

## 10. Ignoring Storage Errors

Storage operations can fail. Check and handle errors from:
- `AddBookmark()`, `RemoveBookmark()`
- `AddHidden()`, `RemoveHidden()`
- `SetLastVoteTimestamp()`
