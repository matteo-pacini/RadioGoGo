# Devtools control socket

A developer build lets an external tool (typically an LLM assistant) attach to a
running RadioGoGo. It can read the screen and state, press keys, and watch the
message flow while the developer uses the TUI in their own terminal.

Everything lives behind the `devtools` build tag. Release builds contain none of it.

## Build and run

```bash
go build -tags devtools -o radiogogo
./radiogogo                     # developer runs this in their terminal
```

The instance listens on `$XDG_RUNTIME_DIR/radiogogo-dev.sock`, or on the same name
in the OS temp dir if `XDG_RUNTIME_DIR` is unset. The socket has mode `0600`.
Only one instance owns the socket. A second instance prints a warning and runs
without it. While a client is connected, the bottom bar shows a `remote` badge.

## ctl client

```bash
./radiogogo ctl screen                      # screen as plain text
./radiogogo ctl state                       # JSON: view, size, stations cursor/selected/playing/volume/recording/modal
./radiogogo ctl key down down enter         # key names use the config keybinding format
./radiogogo ctl type jazz                   # one keypress per character
./radiogogo ctl wait "Now playing" --timeout 5000
./radiogogo ctl watch                       # one JSON event per processed message, until Ctrl+C
```

Exit codes: `0` ok, `1` error or no instance running, `2` bad usage.

Key names are `j`, `B`, `enter`, `esc`, `tab`, `up`, `down`, `pgdown`, `ctrl+k`,
`alt+x`, and `space`. A request with any unknown name delivers no keys.

## Driving it from an LLM

Keys are asynchronous: the reply to `key` only means the keys were queued. To
act on the result, use `wait` for text the next screen will show, not sleeps:

```bash
./radiogogo ctl type jazz
./radiogogo ctl key enter
./radiogogo ctl wait "Jazz" --timeout 10000   # blocks until results render
./radiogogo ctl state                         # cursor, selection, playback
```

To see why the UI did something, run `./radiogogo ctl watch > /tmp/rgg-events.log &`
before acting. The log has message types such as `models.playbackStartedMsg` and
view changes such as `{"type":"view","from":"loading","to":"stations"}`.
Human and remote keypresses interleave with no locking.

## Protocol

Newline-delimited JSON over the socket. Try it with
`socat - UNIX-CONNECT:$XDG_RUNTIME_DIR/radiogogo-dev.sock`:

```
→ {"op":"key","keys":["down","enter"]}
← {"ok":true}
→ {"op":"wait","text":"Now playing","timeout_ms":5000}
← {"ok":true,"screen":"..."}
→ {"op":"watch"}
← {"type":"msg","name":"tea.KeyMsg","ts":"..."}   (stream; "dropped" appears if the client fell behind)
```

`wait` without `timeout_ms` defaults to 5 seconds.

## Code

| File | Contents |
|------|----------|
| `devtools_on.go` / `devtools_off.go` | Build-tag split in `package main` |
| `devtools/server.go` | Socket lifecycle, connection handling |
| `devtools/wrapper.go` | Wrapper model: answers control messages on the Update goroutine |
| `devtools/keys.go` | Key name parser (inverse of `tea.KeyMsg.String()`) |
| `devtools/ctl.go` | `radiogogo ctl` client |
| `models/snapshot.go` | `Model.Snapshot()` used by `state` |

Test: `go test -tags devtools -race ./devtools/`
