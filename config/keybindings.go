// Copyright (c) 2023-2026 Matteo Pacini
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"fmt"
	"strings"
)

// Keybindings holds all customizable key bindings for the application.
// Keys are stored as strings matching BubbleTea's msg.String() format.
type Keybindings struct {
	Quit           string `yaml:"quit"`
	Search         string `yaml:"search"`
	Record         string `yaml:"record"`
	BookmarkToggle string `yaml:"bookmarkToggle"`
	BookmarksView  string `yaml:"bookmarksView"`
	HideStation    string `yaml:"hideStation"`
	ManageHidden   string `yaml:"manageHidden"`
	ChangeLanguage string `yaml:"changeLanguage"`
	VolumeDown     string `yaml:"volumeDown"`
	VolumeUp       string `yaml:"volumeUp"`
	NavigateDown   string `yaml:"navigateDown"`
	NavigateUp     string `yaml:"navigateUp"`
	StopPlayback   string `yaml:"stopPlayback"`
	Vote           string `yaml:"vote"`
}

// reservedKeys contains keys that cannot be remapped because they would break
// BubbleTea components (table, textinput) or violate system/terminal conventions.
var reservedKeys = map[string]bool{
	// Arrow keys - used by table and textinput components
	"up": true, "down": true, "left": true, "right": true,
	// Standard UI keys
	"tab": true, "enter": true, "esc": true,
	// Editing keys - used by textinput component
	"backspace": true, "delete": true,
	// Page navigation - used by table component
	"pgup": true, "pgdown": true, "home": true, "end": true,
	// System signals
	"ctrl+c": true, "ctrl+z": true,
	// Terminal control
	"ctrl+s": true, "ctrl+q": true, "ctrl+l": true,
	// TextInput editing shortcuts
	"ctrl+a": true, "ctrl+e": true, "ctrl+u": true,
	"ctrl+w": true, "ctrl+d": true, "ctrl+h": true,
}

// IsReserved returns true if the key is reserved and cannot be used as a custom keybinding.
func IsReserved(key string) bool {
	return reservedKeys[strings.ToLower(key)]
}

// NewDefaultKeybindings returns the default keybindings for RadioGoGo.
func NewDefaultKeybindings() Keybindings {
	return Keybindings{
		Quit:           "q",
		Search:         "s",
		Record:         "r",
		BookmarkToggle: "b",
		BookmarksView:  "B",
		HideStation:    "h",
		ManageHidden:   "H",
		ChangeLanguage: "L",
		VolumeDown:     "9",
		VolumeUp:       "0",
		NavigateDown:   "j",
		NavigateUp:     "k",
		StopPlayback:   "ctrl+k",
		Vote:           "v",
	}
}

// ValidationWarning represents a warning about an invalid keybinding.
type ValidationWarning struct {
	Key     string
	Value   string
	Reason  string
	Default string
}

func (w ValidationWarning) String() string {
	return fmt.Sprintf("keybinding '%s' has invalid value '%s' (%s), using default '%s'",
		w.Key, w.Value, w.Reason, w.Default)
}

// Validate checks the keybindings for reserved keys and duplicates.
// It returns a list of warnings and a corrected Keybindings struct with defaults
// applied for any invalid keys.
func (k Keybindings) Validate() (Keybindings, []ValidationWarning) {
	defaults := NewDefaultKeybindings()
	warnings := []ValidationWarning{}
	result := k

	// Helper to validate a single key
	type keyField struct {
		name     string
		value    *string
		defValue string
	}

	fields := []keyField{
		{"quit", &result.Quit, defaults.Quit},
		{"search", &result.Search, defaults.Search},
		{"record", &result.Record, defaults.Record},
		{"bookmarkToggle", &result.BookmarkToggle, defaults.BookmarkToggle},
		{"bookmarksView", &result.BookmarksView, defaults.BookmarksView},
		{"hideStation", &result.HideStation, defaults.HideStation},
		{"manageHidden", &result.ManageHidden, defaults.ManageHidden},
		{"changeLanguage", &result.ChangeLanguage, defaults.ChangeLanguage},
		{"volumeDown", &result.VolumeDown, defaults.VolumeDown},
		{"volumeUp", &result.VolumeUp, defaults.VolumeUp},
		{"navigateDown", &result.NavigateDown, defaults.NavigateDown},
		{"navigateUp", &result.NavigateUp, defaults.NavigateUp},
		{"stopPlayback", &result.StopPlayback, defaults.StopPlayback},
		{"vote", &result.Vote, defaults.Vote},
	}

	// Check for reserved keys
	for _, f := range fields {
		if *f.value == "" {
			*f.value = f.defValue
		} else if IsReserved(*f.value) {
			warnings = append(warnings, ValidationWarning{
				Key:     f.name,
				Value:   *f.value,
				Reason:  "reserved key",
				Default: f.defValue,
			})
			*f.value = f.defValue
		}
	}

	// Check for duplicates (excluding already-defaulted keys)
	seen := make(map[string]string)
	for _, f := range fields {
		if existing, ok := seen[*f.value]; ok {
			warnings = append(warnings, ValidationWarning{
				Key:     f.name,
				Value:   *f.value,
				Reason:  fmt.Sprintf("duplicate of '%s'", existing),
				Default: f.defValue,
			})
			*f.value = f.defValue
		} else {
			seen[*f.value] = f.name
		}
	}

	return result, warnings
}
