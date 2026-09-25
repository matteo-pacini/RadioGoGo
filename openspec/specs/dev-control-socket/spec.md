# dev-control-socket Specification

## Purpose

Lets a developer run RadioGoGo interactively while an external client (typically an LLM assistant) attaches to the running instance to read its screen and state, inject keypresses, and observe its internal message flow, for development and debugging only.

## Requirements

### Requirement: Developer-only build gating
The control socket, the `ctl` client subcommand, and the presence indicator SHALL be available only in binaries built with the `devtools` build tag. Binaries built without the tag SHALL NOT open a socket, SHALL NOT accept the `ctl` subcommand, and SHALL NOT show the presence indicator.

#### Scenario: Release build has no control socket
- **WHEN** radiogogo is built without the `devtools` tag and started
- **THEN** no control socket file is created and the TUI behaves exactly as before this change

#### Scenario: Release build rejects ctl
- **WHEN** a binary built without the `devtools` tag is run as `radiogogo ctl screen`
- **THEN** the `ctl` subcommand is not recognised and the argument has no effect on the TUI

#### Scenario: Devtools build starts socket
- **WHEN** radiogogo is built with `-tags devtools` and started in a terminal
- **THEN** the TUI starts normally and the control socket accepts connections

### Requirement: Socket location and lifecycle
In a `devtools` build, radiogogo SHALL listen on a Unix domain socket at `$XDG_RUNTIME_DIR/radiogogo-dev.sock`, or `radiogogo-dev.sock` in the OS temporary directory when `XDG_RUNTIME_DIR` is unset. The socket file SHALL have permissions `0600`. Only one instance SHALL own the socket at a time. The socket file SHALL be removed when radiogogo exits normally.

#### Scenario: Socket permissions
- **WHEN** a devtools build starts
- **THEN** the socket file exists at the resolved path with mode `0600`

#### Scenario: Stale socket from crashed instance
- **WHEN** a devtools build starts and the socket file exists but no process accepts connections on it
- **THEN** the stale file is replaced and the new instance listens on the path

#### Scenario: Second live instance
- **WHEN** a devtools build starts while another instance accepts connections on the socket
- **THEN** the new instance runs the TUI without a control socket and prints a warning to stderr before entering the TUI

#### Scenario: Clean exit
- **WHEN** the user quits radiogogo
- **THEN** the socket file is removed

### Requirement: Request/response protocol
The socket SHALL use newline-delimited JSON. Each request SHALL be one JSON object with an `op` field. Each response SHALL be one JSON object with a boolean `ok` field; when `ok` is false the response SHALL contain an `error` string. A connection MAY carry multiple sequential requests. Multiple clients MAY be connected at the same time.

#### Scenario: Unknown op
- **WHEN** a client sends `{"op":"fly"}`
- **THEN** the response is `{"ok":false,"error":...}` naming the unknown op, and the connection stays open

#### Scenario: Malformed JSON
- **WHEN** a client sends a line that is not valid JSON
- **THEN** the response has `ok` false with an error, and radiogogo keeps running

### Requirement: Key injection
The `key` op SHALL accept a `keys` array of key names and deliver them to the application, in order, as if pressed by the user. Key names SHALL use the same format as keybindings in the config file (for example `j`, `enter`, `esc`, `up`, `ctrl+c`, `alt+x`, `space`). If any key name is invalid, no keys from the request SHALL be delivered.

#### Scenario: Navigate the stations list
- **WHEN** the stations list is shown with the cursor on row 0 and a client sends `{"op":"key","keys":["down","down"]}`
- **THEN** the response has `ok` true and the cursor is on row 2

#### Scenario: Invalid key name
- **WHEN** a client sends `{"op":"key","keys":["down","hyperspace"]}`
- **THEN** the response has `ok` false naming `hyperspace`, and the cursor does not move

#### Scenario: Remote key triggers quit
- **WHEN** a client sends the configured quit key
- **THEN** radiogogo quits exactly as if the user pressed it

### Requirement: Text injection
The `type` op SHALL accept a `text` string and deliver each character, in order, as an individual character keypress.

#### Scenario: Type a search query
- **WHEN** the search input is focused and a client sends `{"op":"type","text":"jazz"}`
- **THEN** the search input contains `jazz`

### Requirement: Screen snapshot
The `screen` op SHALL return a `screen` string containing the currently rendered view with all ANSI escape sequences removed, at the size of the developer's terminal.

#### Scenario: Snapshot matches the terminal
- **WHEN** a client sends `{"op":"screen"}`
- **THEN** the `screen` field contains the same text visible in the developer's terminal, without colour or style codes

### Requirement: Structured state
The `state` op SHALL return a `state` object describing at least: the current view (`boot`, `search`, `loading`, `stations`, `error`, or `terminal_too_small`), terminal width and height, and for the stations view the cursor row index, the selected station's name and UUID, the playing station's name and UUID (or null), the volume, whether recording is active, and whether a modal is open.

#### Scenario: State while playing
- **WHEN** a station is playing on the stations view and a client sends `{"op":"state"}`
- **THEN** the response has `view` `stations` and the playing station's name and UUID

#### Scenario: State on search view
- **WHEN** the search view is shown and a client sends `{"op":"state"}`
- **THEN** the response has `view` `search` and no playing station unless playback continues

### Requirement: Wait for text
The `wait` op SHALL accept a `text` string and a `timeout_ms` integer. It SHALL respond when the rendered screen, with ANSI sequences removed, contains `text`, returning `ok` true and the `screen`. If the timeout expires first, it SHALL respond with `ok` false, an error stating the timeout, and the last `screen`. If the screen already contains `text` when the request arrives, it SHALL respond immediately.

#### Scenario: Wait for search results
- **WHEN** a client submits a search and then sends `{"op":"wait","text":"Jazz FM","timeout_ms":5000}`
- **THEN** the response arrives once results containing `Jazz FM` are rendered, with `ok` true

#### Scenario: Wait times out
- **WHEN** a client sends `{"op":"wait","text":"never-shown","timeout_ms":200}`
- **THEN** after about 200 ms the response has `ok` false and includes the current `screen`

### Requirement: Event stream
The `watch` op SHALL switch the connection into streaming mode. For each message the application processes, radiogogo SHALL write one JSON line containing the message type name and a timestamp, and SHALL write a line when the current view changes, containing the old and new view names. Streaming SHALL continue until the client disconnects. A slow client SHALL NOT delay the application; events that cannot be delivered SHALL be dropped and the next delivered event SHALL report the number dropped.

#### Scenario: Observe a search
- **WHEN** a client is watching and the user submits a search
- **THEN** the client receives events that include the loading transition and the message that delivers results, followed by a view change from `loading` to `stations`

#### Scenario: Slow watcher
- **WHEN** a watching client stops reading from the socket
- **THEN** the TUI stays responsive to the developer's keypresses

### Requirement: Presence indicator
While at least one client is connected to the control socket, the bottom bar SHALL show an indicator that a remote client is attached. The indicator SHALL disappear when the last client disconnects.

#### Scenario: Client attaches and detaches
- **WHEN** a client connects to the socket and then disconnects
- **THEN** the indicator appears while the client is connected and disappears after it disconnects

### Requirement: ctl client
A `devtools` build SHALL provide `radiogogo ctl <op> [args]`, which connects to the socket, sends one request, prints the response to stdout, and exits. It SHALL support `key <name>...`, `type <text>`, `screen`, `state`, `wait <text> [--timeout <ms>]`, and `watch`. `screen` and a successful `wait` SHALL print the screen as plain text. `state` SHALL print JSON. `watch` SHALL print one event per line until interrupted. The process SHALL exit with status 0 when `ok` is true and non-zero otherwise, including when no instance is listening.

#### Scenario: No running instance
- **WHEN** `radiogogo ctl screen` is run and no devtools instance is listening
- **THEN** it prints an error to stderr stating that no instance is running and exits non-zero

#### Scenario: Scripted step
- **WHEN** `radiogogo ctl key down enter` is run against a running instance
- **THEN** the keys are delivered and the command exits 0
