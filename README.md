# RadioGoGo 📻

<div style="display:flex;justify-content:center;">
    <img src="./logo.png" alt="RadioGoGo Logo" width="200" height="200">
</div>
<br>

RadioGoGo is a CLI application developed in Go, enabling seamless access to a wide array of radio stations from around the world directly from your terminal.

Leveraging the streamlined [BubbleTea](https://github.com/charmbracelet/bubbletea) TUI (Terminal User Interface) and the expansive capabilities of [RadioBrowser API](http://www.radio-browser.info/), your desired stations are merely a keystroke away. 

## ⭐️ Features

- Sleek and intuitive TUI that's a joy to navigate.
- Search, browse, and play radio stations from a vast global database.
- Enjoy cross-platform compatibility, because radio waves know no bounds.
- Integrated playback using `ffplay`.

## 📋 Upcoming Features

- Bookmark your favorite stations for easy access.
- Refine your searches to find the perfect station.
- Record your favorite broadcasts for later listening.
- Integrated playback using `mpv`.

## ⚒️ Installation

### Dependencies: Installing FFmpeg

For seamless playback, ensure `ffplay` is installed:

#### Windows:

Download FFmpeg from the [official website](https://ffmpeg.org/download.html) and add it to your system's PATH.

#### Linux:

##### For apt-based distros (like Ubuntu and Debian):

```bash
sudo apt update
sudo apt install ffmpeg
```

##### For dnf-based distros (like Fedora):

```bash
sudo dnf install ffmpeg
```

##### For pacman-based distros (like Arch):

```bash
sudo pacman -S ffmpeg
```

##### For Gentoo:

```bash
emerge --ask --quiet --verbose media-video/ffmpeg
```

#### macOS:

```bash
brew install ffmpeg
```

#### FreeBSD:

```bash
pkg install ffmpeg
```

#### NetBSD:

```bash
pkg_add ffmpeg
```

#### OpenBSD:

```bash
doas pkg_add ffmpeg
```

### Installing via Go

Ensure you have [Go](https://golang.org/dl/) installed (version 1.18 or later).

```bash
go install github.com/Zi0P4tch0/RadioGoGo@latest
```

### Downloading the Binary

Navigate to the `Releases` section of the project repository. 

Find the appropriate binary for your OS, download it, and place it in your system's PATH for easy access.

## 🚀 Usage

1. Launch RadioGoGo by executing the following command:

```bash
radiogogo
```

## ❤️ Contributing

All contributions, big or small, are warmly welcomed. Whether it's a typo fix, new feature, or bug report, I appreciate your effort to make RadioGoGo even better!

## ⚖️ License(s)

RadioGoGo is licensed under the [MIT License](LICENSE).

### Third-party dependencies

BubbleTea TUI license (MIT):

```
MIT License

Copyright (c) 2020-2023 Charmbracelet, Inc

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
