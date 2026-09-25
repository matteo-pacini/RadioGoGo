## Context

See proposal.md for motivation and specs/dev-control-socket/spec.md for behaviour.

Current shape relevant to this change:

- `main.go` builds `models.Model` and runs `tea.NewProgram(model, tea.WithAltScreen())`. There is no argument parsing.
- `models.Model` is a value-receiver BubbleTea model. Its state (`modelState`, child models, playback info) is unexported.
- Keys are matched by `tea.KeyMsg.String()`. Config keybindings are stored in that same format (`config/keybindings.go`).
- BubbleTea v1.3.10. `tea.Program.Send` is safe to call from any goroutine. `Update` and `View` run on the program goroutine.
- `charmbracelet/x/ansi` is already an indirect dependency.

## Goals / Non-Goals

**Goals:**
- Keep `models/` unchanged except for two small, tag-independent hooks: a state snapshot and a presence message.
- Never block `Update` on socket I/O.
- Use one build path for `models/` (no tag-specific files there), so existing tests cover it as-is.

**Non-Goals:**
- Headless mode, TTY-less runs, or size spoofing. The developer's terminal is the only screen.
- An MCP server. `ctl` through a shell is enough. MCP can wrap the socket later without protocol changes.
- Authentication beyond socket file permissions.
- Windows support.
- Arbitration between human and remote input.

## Decisions

### 1. Build tag split lives in `package main` and `devtools/`

```
main.go                  calls runDevtoolsCtl(args) → (exitCode, handled), then setupDevtools(model) → (tea.Model, attach, cleanup)
devtools_on.go   //go:build devtools    → delegates to package devtools
devtools_off.go  //go:build !devtools   → returns model unchanged, no-op attach/cleanup
devtools/        //go:build devtools on every file
  server.go      listener, connection handling, stale-socket check
  wrapper.go     remoteModel wrapping models.Model
  protocol.go    request/response/event types
  keys.go        key-name → tea.KeyMsg parser
  ctl.go         `radiogogo ctl` client
```

`main.go` first hands `os.Args` to `runDevtoolsCtl`, before config or i18n are loaded. If `ctl` is the first argument, the on-variant runs the client and `main` exits with its code. The off-variant never handles args, so release behaviour is unchanged. `setupDevtools` then opens the socket. Instead of the bare model, `main.go` passes the returned wrapped model to `tea.NewProgram`. It then calls `attach(p)` to give the server the `*tea.Program` and calls `cleanup()` after `p.Run` returns, before any `os.Exit`.

*Alternative considered:* a runtime flag. Rejected because the user chose build-time gating to keep input injection out of release binaries.

### 2. Wrapper model owns all remote state

`remoteModel{inner tea.Model, ...}` implements `tea.Model`. Its `Update`:

1. Handles control messages (types private to `devtools`). They are never forwarded:
   - `snapshotReq{reply chan}` replies with `ansi.Strip(inner.View())` and `inner.Snapshot()`.
   - `waitReq{id, text, reply}` is checked immediately. If there is no match, it is stored in a pending map.
   - `waitCancel{id}` removes the pending wait and replies with a timeout and the last screen.
   - `watchReq`/`watchDrop` adds or removes a subscriber channel.
2. Forwards every other message to `inner.Update` and stores the returned model.
3. After forwarding, if waits are pending or watchers exist, renders `inner.View()` once. It then resolves matching waits and emits events to watchers. It detects view changes by comparing `Snapshot().View` before and after.

Reply channels are buffered (size 1), and event sends use `select` with a `default` case. `Update` never blocks. This satisfies CLAUDE.md rule 1. A dropped-event counter per watcher is attached to the next delivered event.

Waits need no polling. The screen can only change after a message is processed, so checking after each message is sufficient. The timeout runs in the connection goroutine with `time.After`, which then sends `waitCancel`.

*Alternative considered:* a mutex-guarded copy of the last `View()` output, updated by the wrapper and read by the server. Rejected: this still needs the wrapper, and it adds shared state for no gain over request messages.

### 3. `models.Model.Snapshot()`: exported, tag-independent

```go
type Snapshot struct {
    View          string // "boot" | "search" | "loading" | "stations" | "error" | "terminal_too_small"
    Width, Height int
    Stations      *StationsSnapshot // nil unless View == "stations"
}
type StationsSnapshot struct {
    Cursor      int
    Selected    *StationRef
    Playing     *StationRef
    Volume      int
    Recording   bool
    ModalOpen   bool
}
```

It is read-only and built from existing fields. It lives in `models/snapshot.go` with unit tests. It is compiled into release builds but unused there, which is a small cost for a single build path. `Snapshot` has no JSON tags. `devtools` maps it to protocol types, so the wire format stays out of `models/`.

*Alternative considered:* a `//go:build devtools` file in `models/`. Rejected because it creates a second build configuration for the package that holds the most logic.

### 4. Presence indicator: exported inert message

`models` exports `RemoteClientsChangedMsg{Count int}`. It is handled in `handleGlobalMessages`, which stores `remoteClients int`. When the count is non-zero, `View()` adds a short indicator to the bottom bar. The indicator is styled via a new `theme.go` style (CLAUDE.md rule 2) with a new i18n key in all locales. Only `devtools` sends this message, so release builds never show the indicator. The server sends it through `p.Send` on each connect and disconnect.

Layout: the indicator is added inline to the existing bottom-bar row, so the bar height is unchanged and `CalculateFillerHeight` is unaffected. If the row overflows at narrow widths, the indicator is dropped before any command hints.

### 5. Key names parsed as the inverse of `KeyMsg.String()`

At init, build a reverse table by iterating `tea.KeyType` values and recording `KeyType.String()` for every named key (`enter`, `esc`, `up`, `ctrl+c`, `space`, …). Parsing rules:
- An `alt+` prefix sets `Alt: true` on the rest of the name.
- A name found in the table becomes that `KeyType`.
- A single rune becomes `KeyRunes`.
- `space` is accepted as an alias, because `KeySpace` stringifies as `" "`.
- Anything else is an error.

A round-trip test asserts `parse(s).String() == s` for every table entry and for sample runes. This keeps parity with config keybindings without a hand-kept list. If iterating `KeyType` proves unsound in v1.3.10, fall back to an explicit table, still covered by the round-trip test.

`type` sends one `KeyRunes` message per rune. It does not send one multi-rune message, because models handle single-key input.

### 6. Socket path and stale detection

The path is `$XDG_RUNTIME_DIR/radiogogo-dev.sock`, falling back to `os.TempDir()`. At startup:
- `net.Dial` to the path.
- If the dial succeeds, another instance is live. Warn on stderr and run without a socket.
- If the dial fails, remove the file, `net.Listen("unix")`, and `os.Chmod(0600)`.

A small race remains between listen and chmod. It is mitigated by setting `umask` around listen, or by relying on `$XDG_RUNTIME_DIR` already being `0700`. Chosen: rely on the directory being private (`$XDG_RUNTIME_DIR` is `0700`; the macOS per-user `$TMPDIR` is private too) and `os.Chmod` after listen. `umask` is avoided because `syscall.Umask` is not portable. Cleanup closes the listener, which removes the file for Unix listeners in Go, and also removes the file explicitly.

### 7. Protocol

The protocol is newline-delimited JSON, with one `bufio.Scanner` per connection and a raised buffer limit. The request is `{"op": ..., ...}`. The response is `{"ok": bool, "error"?: string, "screen"?: string, "state"?: {...}}`. After `watch`, the server only writes events: `{"type":"msg","name":"models.playbackStartedMsg","ts":...,"dropped"?:n}` and `{"type":"view","from":"loading","to":"stations","ts":...}`. The message name comes from `fmt.Sprintf("%T", msg)`, which also works for unexported types.

## Risks / Trade-offs

- [`KeyType` iteration misses or duplicates names] → The round-trip test catches it, with an explicit table as fallback.
- [Rendering `View()` after every message costs CPU while waits or watchers are active] → Only render when pending waits or watchers exist. Normal runs pay nothing.
- [`watch` floods with spinner and tick messages] → This is accepted. The `ctl watch` client can filter with `grep`. Add server-side filtering only if it hurts in practice.
- [Remote keys interleave with human keys] → Accepted for dev. Documented.
- [`Snapshot()` drifts from real UI state as models evolve] → Unit tests per view in `models/snapshot_test.go`.
- [Two terminals racing to start] → The stale check is best-effort. The loser runs without a socket and prints a warning.

## Migration Plan

Additive only. Release builds are unaffected. Document `go build -tags devtools` and `radiogogo ctl` in CLAUDE.md and `.claude/docs/`, and optionally add a `devtools` build alias in `flake.nix`. Rollback means removing the package and the two `main` files. The `models/` hooks are inert on their own.
