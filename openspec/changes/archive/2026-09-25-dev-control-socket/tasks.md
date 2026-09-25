## 1. models hooks (tag-independent)

- [x] 1.1 Add `models/snapshot.go` with `Snapshot`, `StationsSnapshot`, `StationRef` and `func (m Model) Snapshot() Snapshot` built from existing fields; verify with table tests in `models/snapshot_test.go` covering every view name, stations cursor/selected/playing/volume/recording/modal, and `Stations == nil` outside the stations view (`go test ./models/`)
- [x] 1.2 Export `RemoteClientsChangedMsg{Count int}`, handle it in `handleGlobalMessages`, store the count on `Model`; verify with a test that the count updates and the message is not delegated to child models
- [x] 1.3 Add the presence indicator style to `models/theme.go` and render it inline on the bottom bar when count > 0, dropped first when the row overflows; verify with tests that the indicator appears/disappears and that total rendered height is unchanged at min terminal size
- [x] 1.4 Add the indicator i18n key to all 9 files in `i18n/locales/`; verify with a test that loads each locale and asserts `i18n.T` for the new key returns a non-key value in every language (no locale-completeness test exists today)
- [x] 1.5 Verify `go build ./... && go vet ./... && go test ./...` pass without the `devtools` tag

## 2. devtools package: keys and protocol

- [x] 2.1 Create `devtools/` with `//go:build devtools` on every file and `protocol.go` defining request, response and event types; verify `go vet -tags devtools ./devtools/` passes
- [x] 2.2 Implement `keys.go` (reverse table from `tea.KeyType`, `alt+` prefix, single-rune fallback, error on unknown); verify a round-trip test asserting `parse(s).String() == s` for every table entry, sample runes, `alt+x`, and rejection of `hyperspace`. Fall back to an explicit table if `KeyType` iteration proves unsound, and note the decision in design.md
- [x] 2.3 Verify every default keybinding in `config.NewDefaultConfig()` parses via the key parser (test)

## 3. devtools package: wrapper model

- [x] 3.1 Implement `wrapper.go` `remoteModel` that forwards non-control messages to the inner model and intercepts `snapshotReq`; verify with a fake inner model that forwarded messages reach it and control messages do not
- [x] 3.2 Implement `waitReq` / `waitCancel` with immediate match, match after a later message, and cancel returning the last screen; verify with unit tests for all three paths
- [x] 3.3 Implement watcher subscribe/unsubscribe, `%T` message events, view-change events from `Snapshot().View`, non-blocking sends and per-watcher dropped counter; verify with tests including a full (never-read) subscriber channel where `Update` still returns and the next delivered event carries `dropped`
- [x] 3.4 Verify `View()` is rendered after forwarding only when waits or watchers exist (test with a counting fake inner model)

## 4. devtools package: server and ctl

- [x] 4.1 Implement socket path resolution (`$XDG_RUNTIME_DIR`, fallback `os.TempDir()`), stale-socket detection, listen with mode `0600`, and cleanup; verify with tests using a temp dir: fresh start, stale file replaced, live listener causes "run without socket" result, file mode is `0600`, file removed after cleanup
- [x] 4.2 Implement connection handling: JSON-lines loop, ops `key`, `type`, `screen`, `state`, `wait`, `watch`, errors for unknown op and malformed JSON with the connection kept open, invalid key rejects the whole request; verify with tests driving a real `tea.Program` (no renderer, `tea.WithInput(nil)`, `tea.WithOutput(io.Discard)`) wrapping a fake model over a temp socket
- [x] 4.3 Send `RemoteClientsChangedMsg` on each connect/disconnect with the current count; verify by test with two concurrent clients (counts 1, 2, 1, 0)
- [x] 4.4 Implement `ctl.go` for `key`, `type`, `screen`, `state`, `wait [--timeout]`, `watch` with the output formats and exit codes from the spec; verify with tests against a test server, including "no instance running" exit non-zero with a stderr message
- [x] 4.5 Verify `go test -tags devtools -race ./devtools/` passes

## 5. main wiring

- [x] 5.1 Add `devtools_on.go` / `devtools_off.go` in package main exposing `devtoolsSetup`, and wire `main.go` to use the wrapped model, `attach(p)`, deferred `cleanup()`, and early `ctl` dispatch; verify `go build -o radiogogo` and `go build -tags devtools -o radiogogo` both succeed
- [x] 5.2 Verify the release binary contains no devtools code: `go build -o /tmp/rgg-release . && go tool nm /tmp/rgg-release | grep -c devtools` prints 0
- [x] 5.3 Verify full suite: `go fmt ./... && go vet ./... && go vet -tags devtools ./... && go test ./... && go test -tags devtools -race ./...`

## 6. Docs

- [x] 6.1 Document `go build -tags devtools`, the socket path, `radiogogo ctl` usage and a sample LLM loop in `CLAUDE.md` (Commands) and a new `.claude/docs/devtools.md`; verify the docs table in `CLAUDE.md` links the new file
- [ ] 6.2 Optionally add a `devtools` build alias to `flake.nix`; verify `nix build` or `direnv reload` still works

## 7. Live end-to-end validation (developer runs radiogogo, Claude drives via ctl)

Setup: the developer builds with `go build -tags devtools -o radiogogo` and runs `./radiogogo` in their own terminal. Claude runs `./radiogogo ctl ...` through Bash. Record each result (pass/fail with the output) in the apply summary. Any test that changes storage (bookmark, hide) MUST be reverted within the same task.

- [x] 7.1 Socket exists at the resolved path with mode `0600` (`stat`); pass if mode is `srw-------`
- [x] 7.2 Presence: developer confirms the indicator is visible while `ctl watch` is running and gone after it is interrupted
- [x] 7.3 `ctl screen` output matches what the developer sees (developer confirms), contains no ESC bytes (`grep -c $'\x1b'` prints 0)
- [x] 7.4 `ctl state` on the search view returns `view: search` with the terminal width/height; developer resizes the terminal and a second `ctl state` reflects the new size
- [x] 7.5 `ctl type jazz` then `ctl screen` shows `jazz` in the search input
- [x] 7.6 Start `ctl watch` in the background to a log file, then `ctl key enter` and `ctl wait <a station name or results marker> --timeout 10000` succeed; the watch log contains a view change `search→loading` and `loading→stations`
- [x] 7.7 `ctl state` on stations view returns cursor 0; `ctl key down down` then `ctl state` returns cursor 2 and the selected station matches row 2 in `ctl screen`
- [x] 7.8 Invalid key: `ctl key down hyperspace` exits non-zero naming `hyperspace` and `ctl state` shows the cursor unchanged
- [x] 7.9 Play: `ctl key enter` then `ctl wait` for the now-playing indicator; `ctl state` returns the playing station name/UUID; developer confirms audio
- [x] 7.10 Volume: press the configured volume-up key twice and volume-down once; `ctl state` volume changes by the expected net step
- [x] 7.11 Recording: toggle record on, `ctl state` shows recording true, toggle off, recording false; delete the produced recording file if the developer agrees
- [x] 7.12 Stop playback via the configured key; `ctl state` shows no playing station
- [x] 7.13 Bookmark: toggle bookmark on the selected station, open bookmarks view, `ctl screen` lists it; toggle it off again and confirm removal (revert)
- [x] 7.14 Hide/unhide: hide the selected station, confirm it disappears from `ctl screen`; open the hidden-stations modal, `ctl state` shows modal open, unhide it and confirm it is back (revert)
- [x] 7.15 Wait timeout: `ctl wait never-shown-xyz --timeout 300` exits non-zero after ~300 ms (`time`) and prints the current screen
- [x] 7.16 Protocol errors via raw socket (`socat - UNIX-CONNECT:<path>`; both `nc` and `socat` are installed): unknown op returns `ok:false`, malformed JSON returns `ok:false`, and a following valid `{"op":"state"}` on the same connection succeeds
- [x] 7.17 Concurrent clients: run `ctl watch` twice in the background plus `ctl state` loops; TUI remains responsive (developer confirms) and the indicator stays visible until both watches are stopped
- [x] 7.18 Slow watcher: `ctl watch | sleep 60` in the background while the developer navigates; developer confirms no lag, and after it ends a new `ctl watch` shows a `dropped` count or normal flow
- [x] 7.19 Human + remote interleave: developer presses keys while Claude sends `ctl key down` repeatedly; no crash and final `ctl state` is consistent with `ctl screen`
- [x] 7.20 Second instance: Claude runs `timeout 2 ./radiogogo </dev/null` (or the developer starts one in another terminal); stderr contains the "another instance" warning and the first instance's socket keeps working (`ctl state`)
- [x] 7.21 Remote quit: `ctl key <quit key>` exits radiogogo and the socket file no longer exists
- [x] 7.22 Stale socket: developer starts radiogogo, Claude kills it with `kill -9`, socket file remains; developer restarts it and `ctl state` succeeds
- [x] 7.23 No instance: with radiogogo stopped, `ctl screen` exits non-zero with the "no instance running" message
- [x] 7.24 Release build: `go build -o /tmp/rgg-release .` and developer runs `/tmp/rgg-release`; no socket file is created and no indicator appears; developer runs `/tmp/rgg-release ctl screen` and the TUI starts normally
