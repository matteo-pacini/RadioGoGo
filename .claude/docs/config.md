# Configuration

## File Location

- **Unix/macOS:** `~/.config/radiogogo/config.yaml`
- **Windows:** `%LOCALAPPDATA%\radiogogo\config.yaml`

## Structure

```yaml
language: en  # de, el, en, es, it, ja, pt, ru, zh

theme:
  textColor: "#ffffff"
  primaryColor: "#5a4f9f"
  secondaryColor: "#8b77db"
  tertiaryColor: "#4e4e4e"
  errorColor: "#ff0000"

keybindings:
  quit: "q"
  search: "s"
  record: "r"
  bookmarkToggle: "b"
  bookmarksView: "B"
  hideStation: "h"
  manageHidden: "H"
  changeLanguage: "L"
  volumeDown: "9"
  volumeUp: "0"
  navigateDown: "j"
  navigateUp: "k"
  stopPlayback: "ctrl+k"
  vote: "v"

playerPreferences:
  defaultVolume: 80  # 0-100
```

## Keybinding Rules

**Reserved keys (cannot be remapped):**
- Navigation: up, down, left, right, tab, enter, esc
- Editing: backspace, delete, pgup, pgdown, home, end
- System: ctrl+c, ctrl+z, ctrl+s, ctrl+q, ctrl+l
- TextInput: ctrl+a, ctrl+e, ctrl+u, ctrl+w, ctrl+d, ctrl+h

**Validation:** `main.go` at startup - rejects reserved/duplicate keys with warnings.

## Internationalization

**Languages:** de, el, en, es, it, ja, pt, ru, zh

**Usage:**
```go
i18n.T("message_id")                    // Simple translation
i18n.Tf("message_id", map[string]interface{}{"Key": value})  // With data
i18n.Tn("message_id", count)            // Pluralization
```

**Adding a language:**
1. Create `i18n/locales/XX.yaml` (copy from `en.yaml`)
2. Translate all message strings
3. App auto-discovers via `//go:embed`

**Runtime switching:** Press "L" on search screen.
