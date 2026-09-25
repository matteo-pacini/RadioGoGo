## Why

Developing and debugging RadioGoGo with an LLM assistant is limited by the fact that the assistant cannot see or drive the running TUI. A developer-only control socket lets the developer run radiogogo in their own terminal while an LLM attaches, reads the screen and state, injects keypresses, and watches the internal message flow, making exploratory testing and debugging collaborative.

## What Changes

- New `devtools` build tag. All control-socket code is compiled only with `-tags devtools`; release binaries are unchanged and contain no remote input path.
- With the tag, radiogogo listens on a Unix socket at a fixed path (`$XDG_RUNTIME_DIR/radiogogo-dev.sock`, mode `0600`) while the developer uses the TUI normally. A single dev instance is supported; a stale socket is cleaned up at startup, and the socket is removed on exit.
- JSON-lines request/response protocol with verbs:
  - `key` — inject one or more keypresses (e.g. `down`, `enter`, `ctrl+c`, `j`).
  - `type` — inject a string as individual rune keypresses.
  - `screen` — return the current rendered view with ANSI sequences stripped.
  - `state` — return structured state (current state name, cursor/selection, playing station, volume, recording).
  - `wait` — block until the screen contains given text, or a timeout expires.
  - `watch` — stream an event per message processed (message type name, state transitions) until the client disconnects.
- New `radiogogo ctl <verb> ...` client subcommand (also behind the tag) so LLMs and scripts can use the socket from a shell.
- Presence indicator in the bottom bar while at least one client is attached; shown only in `devtools` builds.
- Human and remote keypresses interleave without locking; this is accepted for a dev-only tool.

## Capabilities

### New Capabilities
- `dev-control-socket`: developer-only control socket and `ctl` client — build gating, socket lifecycle, protocol verbs (`key`, `type`, `screen`, `state`, `wait`, `watch`), and attached-client presence indicator.

### Modified Capabilities
<!-- none -->

## Impact

- `main.go`: wires the control socket and the `ctl` subcommand under the `devtools` tag.
- New package (e.g. `devtools/`) containing the socket server, protocol, key parser, and a wrapper `tea.Model` that intercepts control messages and forwards everything else to `models.Model`.
- `models/`: minimal hook for the presence indicator and for exposing structured state; must stay inert in non-`devtools` builds.
- Build/dev environment: `flake.nix` / `direnv` and `.claude/docs` updated to document building with `-tags devtools` and using `radiogogo ctl`.
- No new runtime dependencies expected; ANSI stripping may use `charmbracelet/x/ansi` (already transitive via lipgloss).
