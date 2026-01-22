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
- Customizable color themes
- Cross-platform (Linux, macOS, Windows, *BSD)

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
| `s` | Back to search |
| `q` | Quit |

## Recording

Press `r` while a station is playing to start recording. The file saves to your current directory with the format:

```
station_name-YYYY-MM-DD-HH-MM-SS.codec
```

For example: `bbc_radio_1-2026-01-22-18-32-00.mp3`

Press `r` again to stop recording. The recording continues even if you adjust volume (only the player restarts, not the recorder).

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
