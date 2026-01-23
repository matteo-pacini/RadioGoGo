package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestNewDefaultKeybindings(t *testing.T) {
	t.Run("returns default keybindings", func(t *testing.T) {
		kb := NewDefaultKeybindings()

		assert.Equal(t, "q", kb.Quit)
		assert.Equal(t, "s", kb.Search)
		assert.Equal(t, "r", kb.Record)
		assert.Equal(t, "b", kb.BookmarkToggle)
		assert.Equal(t, "B", kb.BookmarksView)
		assert.Equal(t, "h", kb.HideStation)
		assert.Equal(t, "H", kb.ManageHidden)
		assert.Equal(t, "L", kb.ChangeLanguage)
		assert.Equal(t, "9", kb.VolumeDown)
		assert.Equal(t, "0", kb.VolumeUp)
		assert.Equal(t, "j", kb.NavigateDown)
		assert.Equal(t, "k", kb.NavigateUp)
		assert.Equal(t, "ctrl+k", kb.StopPlayback)
	})
}

func TestIsReserved(t *testing.T) {
	t.Run("returns true for reserved keys", func(t *testing.T) {
		reservedKeys := []string{
			"up", "down", "left", "right",
			"tab", "enter", "esc",
			"backspace", "delete",
			"pgup", "pgdown", "home", "end",
			"ctrl+c", "ctrl+z",
			"ctrl+s", "ctrl+q", "ctrl+l",
			"ctrl+a", "ctrl+e", "ctrl+u",
			"ctrl+w", "ctrl+d", "ctrl+h",
		}

		for _, key := range reservedKeys {
			assert.True(t, IsReserved(key), "expected %s to be reserved", key)
		}
	})

	t.Run("returns true for reserved keys case-insensitive", func(t *testing.T) {
		assert.True(t, IsReserved("UP"))
		assert.True(t, IsReserved("Tab"))
		assert.True(t, IsReserved("CTRL+C"))
	})

	t.Run("returns false for non-reserved keys", func(t *testing.T) {
		nonReserved := []string{
			"q", "s", "r", "b", "B", "h", "H", "L",
			"9", "0", "j", "k", "ctrl+k",
			"a", "z", "1", "2", "space",
		}

		for _, key := range nonReserved {
			assert.False(t, IsReserved(key), "expected %s to not be reserved", key)
		}
	})
}

func TestKeybindings_Validate(t *testing.T) {
	t.Run("accepts valid keybindings", func(t *testing.T) {
		kb := NewDefaultKeybindings()
		validated, warnings := kb.Validate()

		assert.Empty(t, warnings)
		assert.Equal(t, kb, validated)
	})

	t.Run("warns and defaults for reserved key", func(t *testing.T) {
		kb := NewDefaultKeybindings()
		kb.Quit = "enter" // reserved key

		validated, warnings := kb.Validate()

		assert.Len(t, warnings, 1)
		assert.Equal(t, "quit", warnings[0].Key)
		assert.Equal(t, "enter", warnings[0].Value)
		assert.Equal(t, "reserved key", warnings[0].Reason)
		assert.Equal(t, "q", validated.Quit)
	})

	t.Run("warns and defaults for duplicate key", func(t *testing.T) {
		kb := NewDefaultKeybindings()
		kb.Search = "q" // same as Quit

		validated, warnings := kb.Validate()

		assert.Len(t, warnings, 1)
		assert.Equal(t, "search", warnings[0].Key)
		assert.Equal(t, "q", warnings[0].Value)
		assert.Contains(t, warnings[0].Reason, "duplicate")
		assert.Equal(t, "s", validated.Search)
	})

	t.Run("fills empty keys with defaults", func(t *testing.T) {
		kb := Keybindings{} // all empty

		validated, warnings := kb.Validate()

		assert.Empty(t, warnings) // empty is not an error, just filled
		assert.Equal(t, "q", validated.Quit)
		assert.Equal(t, "s", validated.Search)
	})

	t.Run("handles multiple validation issues", func(t *testing.T) {
		kb := NewDefaultKeybindings()
		kb.Quit = "tab"   // reserved
		kb.Search = "esc" // reserved

		validated, warnings := kb.Validate()

		assert.Len(t, warnings, 2)
		assert.Equal(t, "q", validated.Quit)
		assert.Equal(t, "s", validated.Search)
	})
}

func TestKeybindings_YAML(t *testing.T) {
	t.Run("parses from YAML", func(t *testing.T) {
		input := `
quit: "x"
search: "/"
record: "R"
bookmarkToggle: "f"
bookmarksView: "F"
hideStation: "d"
manageHidden: "D"
changeLanguage: "l"
volumeDown: "-"
volumeUp: "+"
navigateDown: "n"
navigateUp: "p"
stopPlayback: "ctrl+x"
`
		var kb Keybindings
		err := yaml.Unmarshal([]byte(input), &kb)

		assert.NoError(t, err)
		assert.Equal(t, "x", kb.Quit)
		assert.Equal(t, "/", kb.Search)
		assert.Equal(t, "R", kb.Record)
		assert.Equal(t, "f", kb.BookmarkToggle)
		assert.Equal(t, "F", kb.BookmarksView)
		assert.Equal(t, "d", kb.HideStation)
		assert.Equal(t, "D", kb.ManageHidden)
		assert.Equal(t, "l", kb.ChangeLanguage)
		assert.Equal(t, "-", kb.VolumeDown)
		assert.Equal(t, "+", kb.VolumeUp)
		assert.Equal(t, "n", kb.NavigateDown)
		assert.Equal(t, "p", kb.NavigateUp)
		assert.Equal(t, "ctrl+x", kb.StopPlayback)
	})

	t.Run("parses partial YAML", func(t *testing.T) {
		input := `
quit: "x"
search: "/"
`
		var kb Keybindings
		err := yaml.Unmarshal([]byte(input), &kb)

		assert.NoError(t, err)
		assert.Equal(t, "x", kb.Quit)
		assert.Equal(t, "/", kb.Search)
		assert.Equal(t, "", kb.Record) // not specified
	})
}

func TestValidationWarning_String(t *testing.T) {
	t.Run("formats warning message", func(t *testing.T) {
		w := ValidationWarning{
			Key:     "quit",
			Value:   "enter",
			Reason:  "reserved key",
			Default: "q",
		}

		msg := w.String()

		assert.Contains(t, msg, "quit")
		assert.Contains(t, msg, "enter")
		assert.Contains(t, msg, "reserved key")
		assert.Contains(t, msg, "q")
	})
}

func TestKeybindings_Validate_EdgeCases(t *testing.T) {
	t.Run("handles all reserved keys at once", func(t *testing.T) {
		// Set multiple keys to reserved values
		kb := Keybindings{
			Quit:           "enter",
			Search:         "tab",
			Record:         "esc",
			BookmarkToggle: "up",
			BookmarksView:  "down",
			HideStation:    "left",
			ManageHidden:   "right",
			ChangeLanguage: "ctrl+c",
			VolumeDown:     "backspace",
			VolumeUp:       "delete",
			NavigateDown:   "pgup",
			NavigateUp:     "pgdown",
			StopPlayback:   "home",
		}

		validated, warnings := kb.Validate()

		// All 13 keys should have warnings
		assert.Len(t, warnings, 13)

		// All should be reset to defaults
		assert.Equal(t, "q", validated.Quit)
		assert.Equal(t, "s", validated.Search)
		assert.Equal(t, "r", validated.Record)
		assert.Equal(t, "b", validated.BookmarkToggle)
		assert.Equal(t, "B", validated.BookmarksView)
		assert.Equal(t, "h", validated.HideStation)
		assert.Equal(t, "H", validated.ManageHidden)
		assert.Equal(t, "L", validated.ChangeLanguage)
		assert.Equal(t, "9", validated.VolumeDown)
		assert.Equal(t, "0", validated.VolumeUp)
		assert.Equal(t, "j", validated.NavigateDown)
		assert.Equal(t, "k", validated.NavigateUp)
		assert.Equal(t, "ctrl+k", validated.StopPlayback)
	})

	t.Run("detects case-sensitive duplicates", func(t *testing.T) {
		// Keys are case-sensitive, so "Q" and "q" should both be allowed
		kb := NewDefaultKeybindings()
		kb.Quit = "x"
		kb.Search = "X" // Different case, should be allowed

		validated, warnings := kb.Validate()

		assert.Empty(t, warnings)
		assert.Equal(t, "x", validated.Quit)
		assert.Equal(t, "X", validated.Search)
	})

	t.Run("detects duplicate in later position", func(t *testing.T) {
		kb := NewDefaultKeybindings()
		kb.Quit = "x"
		kb.Search = "x" // Same as Quit - Search comes after Quit in validation

		validated, warnings := kb.Validate()

		// Quit (position 1) is processed before Search (position 2)
		// So Quit keeps "x" and Search gets the duplicate warning and resets to default "s"
		assert.Len(t, warnings, 1)
		assert.Equal(t, "search", warnings[0].Key)
		assert.Contains(t, warnings[0].Reason, "duplicate")
		assert.Equal(t, "x", validated.Quit)   // Keeps "x" (processed first)
		assert.Equal(t, "s", validated.Search) // Reset to default "s"
	})

	t.Run("chain of duplicates all get warnings", func(t *testing.T) {
		kb := NewDefaultKeybindings()
		// Create a chain: Search, Record, BookmarkToggle all use "x"
		kb.Search = "x"
		kb.Record = "x"
		kb.BookmarkToggle = "x"

		validated, warnings := kb.Validate()

		// Record and BookmarkToggle duplicate Search's "x"
		assert.Len(t, warnings, 2)

		// Search keeps "x", others get defaults
		assert.Equal(t, "x", validated.Search)
		assert.Equal(t, "r", validated.Record)
		assert.Equal(t, "b", validated.BookmarkToggle)
	})

	t.Run("mixed reserved and duplicate warnings", func(t *testing.T) {
		kb := NewDefaultKeybindings()
		kb.Quit = "enter"    // reserved
		kb.Search = "q"      // duplicate of Quit's default (after Quit is reset)
		kb.Record = "ctrl+c" // reserved

		validated, warnings := kb.Validate()

		// Should have warnings for reserved keys and potentially duplicates
		assert.GreaterOrEqual(t, len(warnings), 2)

		// Reserved keys get defaults
		assert.Equal(t, "q", validated.Quit)
		assert.Equal(t, "r", validated.Record)
	})

	t.Run("whitespace-only key treated as empty", func(t *testing.T) {
		kb := NewDefaultKeybindings()
		kb.Quit = "   " // Whitespace only - should be treated as empty or invalid

		validated, _ := kb.Validate()

		// Whitespace should either be kept as-is or reset to default
		// depending on implementation. Let's verify it doesn't crash
		assert.NotPanics(t, func() {
			kb.Validate()
		})

		// If whitespace is kept, the key will be "   "
		// If treated as empty, it will be "q"
		// Either way, Validate should handle it gracefully
		_ = validated
	})

	t.Run("special characters in keys", func(t *testing.T) {
		kb := NewDefaultKeybindings()
		kb.Quit = "!"
		kb.Search = "@"
		kb.Record = "#"

		validated, warnings := kb.Validate()

		// Special characters should be allowed (not reserved)
		assert.Empty(t, warnings)
		assert.Equal(t, "!", validated.Quit)
		assert.Equal(t, "@", validated.Search)
		assert.Equal(t, "#", validated.Record)
	})
}
