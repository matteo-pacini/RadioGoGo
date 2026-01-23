# RadioGoGo

<div style="display:flex;justify-content:center;">
    <img src="./logo.png" alt="RadioGoGo Logo" width="200" height="200">
</div>

A terminal UI for browsing and playing internet radio stations. Built with Go using [BubbleTea](https://github.com/charmbracelet/bubbletea) and the [RadioBrowser API](http://www.radio-browser.info/).

<img src="./screen1.png" alt="RadioGoGo Search View" width="500" height="320">
<img src="./screen2.png" alt="RadioGoGo Station List View" width="500" height="320">

## Features

- Search stations by name, country, language, or codec
- Browse results in a navigable table
- Stream playback via `ffplay`
- Real-time volume control during playback
- Record streams to disk via `ffmpeg`
- Customizable color themes and keybindings
- Bookmark favorite stations for quick access
- Hide unwanted stations from search results
- Cross-platform (Linux, macOS, Windows, *BSD)
- Multi-language UI (English, German, Greek, Spanish, Italian, Japanese, Portuguese, Russian, Chinese)

## How It Works

RadioGoGo uses FFmpeg tools for audio:

- **Playback**: `ffplay` handles audio streaming. Volume changes restart the player with the new level (with debouncing to avoid rapid restarts).
- **Recording**: `ffmpeg` runs alongside `ffplay` when recording. Both connect to the stream independently—audio keeps playing while the stream saves to disk.

The header shows two status indicators:
- `(●) ffplay` — green when playing, yellow during volume restart, gray when idle
- `(●) rec` — red when recording, gray when idle

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Play selected station |
| `Ctrl+K` | Stop playback |
| `9` / `0` | Volume down / up |
| `r` | Toggle recording (while playing) |
| `↑` / `↓` or `j` / `k` | Navigate station list |
| `b` | Toggle bookmark on selected station |
| `B` | View bookmarks / back to stations |
| `h` | Hide station from results |
| `H` | Manage hidden stations |
| `s` | Back to search |
| `L` | Cycle UI language (search screen) |
| `q` | Quit |

Most keys are customizable via config (see [Custom Keybindings](#custom-keybindings) below). Keys that cannot be changed: arrow keys, Enter, Tab, Escape, and common editing keys (Backspace, Delete, Ctrl+C, etc.).

## Recording

Press `r` while a station is playing to start recording. The file saves to your current directory with the format:

```
station_name-YYYY-MM-DD-HH-MM-SS.codec
```

For example: `bbc_radio_1-2026-01-22-18-32-00.mp3`

Press `r` again to stop recording. The recording continues even if you adjust volume (only the player restarts, not the recorder).

## Bookmarks & Hidden Stations

**Bookmarks:** Press `b` on any station to bookmark it (⭐ appears next to name). Press `B` to view all bookmarks. Press `B` again to return to your search results.

**Hidden Stations:** Press `h` to hide a station from search results. Press `H` to manage hidden stations and unhide them if needed.

Bookmarks and hidden stations persist across sessions.

## Installation

### Dependencies

You need `ffplay` for playback. For recording, you also need `ffmpeg`. Both come with the FFmpeg package.

**Windows:**
```
choco install ffmpeg
```
or
```
scoop install ffmpeg
```

**Linux (apt):**
```
sudo apt install ffmpeg
```

**Linux (dnf/Fedora):**
```
sudo dnf install https://download1.rpmfusion.org/free/fedora/rpmfusion-free-release-$(rpm -E %fedora).noarch.rpm
sudo dnf install ffmpeg
```

**Linux (pacman):**
```
sudo pacman -S ffmpeg
```

**macOS:**
```
brew install ffmpeg
```

**FreeBSD:**
```
pkg install ffmpeg
```

### Install RadioGoGo

**Via Go:**
```bash
go install github.com/zi0p4tch0/radiogogo@latest
```

Make sure `$(go env GOPATH)/bin` is in your PATH.

**Via releases:**

Download the binary for your platform from the [Releases](https://github.com/zi0p4tch0/radiogogo/releases) page.

## Configuration

Config file location:
- **Windows:** `%LOCALAPPDATA%\radiogogo\config.yaml`
- **Linux/macOS/*BSD:** `~/.config/radiogogo/config.yaml`

Created automatically on first run.

### Theme

```yaml
theme:
    textColor: '#ffffff'
    primaryColor: '#5a4f9f'
    secondaryColor: '#8b77db'
    tertiaryColor: '#4e4e4e'
    errorColor: '#ff0000'
```

### Language

Set the UI language:

```yaml
language: en
```

Available: `de` (German), `el` (Greek), `en` (English), `es` (Spanish), `it` (Italian), `ja` (Japanese), `pt` (Portuguese), `ru` (Russian), `zh` (Chinese)

Press `L` on the search screen to cycle through languages.

### Custom Keybindings

Most keys can be customized. Changes require restarting the app.

```yaml
keybindings:
  quit: q
  search: s
  record: r
  bookmarkToggle: b
  bookmarksView: B
  hideStation: h
  manageHidden: H
  changeLanguage: L
  volumeDown: "9"
  volumeUp: "0"
  navigateDown: j
  navigateUp: k
  stopPlayback: ctrl+k
```

**Reserved keys** (cannot be remapped): arrow keys (`up`, `down`, `left`, `right`), `tab`, `enter`, `esc`, `backspace`, `delete`, `pgup`, `pgdown`, `home`, `end`, and terminal control keys (`ctrl+c`, `ctrl+z`, `ctrl+s`, `ctrl+q`, `ctrl+l`, `ctrl+a`, `ctrl+e`, `ctrl+u`, `ctrl+k`, `ctrl+w`, `ctrl+d`, `ctrl+h`).

If you set an invalid key or duplicate, the app warns at startup and uses the default for that key.

Example alternate theme:

```yaml
theme:
    textColor: '#f0e6e6'
    primaryColor: '#c41230'
    secondaryColor: '#e4414f'
    tertiaryColor: '#f58b8d'
    errorColor: '#ff0000'
```

<img src="./screen3.png" alt="RadioGoGo Alternate Theme" width="500" height="320">
<img src="./screen4.png" alt="RadioGoGo Alternate Theme" width="500" height="320">

## FAQ

**Station takes a while to start playing?**

Some streams need time to buffer depending on server location and connection. Wait a few seconds.

**Station doesn't work at all?**

Stations go offline or change URLs. RadioBrowser is community-maintained, so some entries may be stale.

**Recording requires ffmpeg?**

Yes. Playback only needs `ffplay`, but recording needs `ffmpeg` installed and in your PATH.

## Mentions

- [Golang Weekly Issue 481](https://golangweekly.com/issues/481)

## Contributing

Bug reports, fixes, and feature ideas welcome. For new features, open an issue first to discuss.

## License

MIT. See [LICENSE](LICENSE).
