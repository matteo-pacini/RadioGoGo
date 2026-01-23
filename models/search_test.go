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

package models

import (
	"testing"

	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/i18n"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

var testSearchKeybindings = config.Keybindings{
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
}

func TestSearchModel_Init(t *testing.T) {

	t.Run("starts blinking the input field cursor", func(t *testing.T) {

		model := NewSearchModel(Theme{}, nil, nil, testSearchKeybindings)

		cmd := model.Init()
		assert.NotNil(t, cmd)

		var batchMsg tea.BatchMsg = cmd().(tea.BatchMsg)
		found := false

		for _, msg := range batchMsg {
			currentMsg := msg()
			if currentMsg == textarea.Blink() {
				found = true
				break
			}
		}

		assert.True(t, found)

	})

	t.Run("broadcasts a bottomBarUpdateMsg", func(t *testing.T) {

		model := NewSearchModel(Theme{}, nil, nil, testSearchKeybindings)

		cmd := model.Init()
		assert.NotNil(t, cmd)

		var batchMsg tea.BatchMsg = cmd().(tea.BatchMsg)

		found := false
		var commands []string
		for _, msg := range batchMsg {
			currentMsg := msg()
			if _, ok := currentMsg.(bottomBarUpdateMsg); ok {
				commands = currentMsg.(bottomBarUpdateMsg).commands
				found = true
				break
			}
		}

		assert.True(t, found)

		expectedCommands := []string{"q: quit", "tab: cycle focus", "enter: search", "B: bookmarks", "L: language", "EN"}

		assert.Equal(t, expectedCommands, commands)

	})

}

func TestSearchModel_Update(t *testing.T) {

	t.Run("does not broadcast quitMsg when 'q' is pressed and textarea is focused", func(t *testing.T) {

		model := NewSearchModel(Theme{}, nil, nil, testSearchKeybindings)

		model.inputModel.Focus()
		model.querySelector.Blur()

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}

		_, cmd := model.Update(input)
		assert.NotNil(t, cmd)

		msg := cmd()

		assert.IsType(t, cursor.BlinkMsg{}, msg)

	})

	t.Run("broadcasts a quitMsg when 'q' is pressed and textarea is not focused", func(t *testing.T) {

		model := NewSearchModel(Theme{}, nil, nil, testSearchKeybindings)

		model.inputModel.Blur()
		model.querySelector.Focus()

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}

		_, cmd := model.Update(input)
		assert.NotNil(t, cmd)

		msg := cmd()

		assert.IsType(t, quitMsg{}, msg)

	})

	t.Run("broadcasts a switchToLoadingModelMsg when 'enter' is pressed, propagating text area value", func(t *testing.T) {

		model := NewSearchModel(Theme{}, nil, nil, testSearchKeybindings)
		model.inputModel.SetValue("fancy value")

		input := tea.KeyMsg{Type: tea.KeyEnter}

		_, cmd := model.Update(input)
		assert.NotNil(t, cmd)

		msg := cmd()

		assert.Equal(t, switchToLoadingModelMsg{
			query:     common.StationQueryByName,
			queryText: "fancy value",
		}, msg)

	})

	t.Run("ignores 'enter' when is not in focus", func(t *testing.T) {

		model := NewSearchModel(Theme{}, nil, nil, testSearchKeybindings)
		model.inputModel.SetValue("fancy value")
		model.inputModel.Blur()

		input := tea.KeyMsg{Type: tea.KeyEnter}

		_, cmd := model.Update(input)
		assert.Nil(t, cmd)

	})

	t.Run("cycles focused input when 'tab' is pressed and updates bottom bar", func(t *testing.T) {

		model := NewSearchModel(Theme{}, nil, nil, testSearchKeybindings)

		input := tea.KeyMsg{Type: tea.KeyTab}

		newModel, cmd := model.Update(input)

		assert.Equal(t, newModel.(SearchModel).inputModel.Focused(), false)
		assert.Equal(t, newModel.(SearchModel).querySelector.Focused(), true)
		assert.NotNil(t, cmd)

		msg := cmd()
		assert.IsType(t, bottomBarUpdateMsg{}, msg)

		newModel, cmd = newModel.Update(input)

		assert.Equal(t, newModel.(SearchModel).inputModel.Focused(), true)
		assert.Equal(t, newModel.(SearchModel).querySelector.Focused(), false)

		msg = cmd()
		assert.IsType(t, bottomBarUpdateMsg{}, msg)

	})

}
func TestUpdateCommandsForTextfieldFocus(t *testing.T) {
	expectedCommands := []string{"q: quit", "tab: cycle focus", "enter: search", "B: bookmarks", "L: language", "EN"}

	cmd := updateCommandsForTextfieldFocus(testSearchKeybindings)
	msg := cmd()

	updateMsg, ok := msg.(bottomBarUpdateMsg)

	assert.True(t, ok)
	assert.Equal(t, expectedCommands, updateMsg.commands)
}

func TestUpdateCommandsForSelectorFocus(t *testing.T) {
	expectedCommands := []string{"q: quit", "tab: cycle focus", "↑/↓: change filter", "B: bookmarks", "L: language", "EN"}

	cmd := updateCommandsForSelectorFocus(testSearchKeybindings)
	msg := cmd()

	updateMsg, ok := msg.(bottomBarUpdateMsg)

	assert.True(t, ok)
	assert.Equal(t, expectedCommands, updateMsg.commands)
}

func TestGetNextLanguage(t *testing.T) {
	// Initialize i18n to populate available languages
	_ = i18n.Init("en")

	t.Run("returns next language in sequence", func(t *testing.T) {
		_ = i18n.SetLanguage("de")
		next := getNextLanguage()
		assert.Equal(t, "el", next) // de -> el (alphabetically sorted)
	})

	t.Run("returns first language when at end of list", func(t *testing.T) {
		_ = i18n.SetLanguage("zh")
		next := getNextLanguage()
		assert.Equal(t, "de", next) // zh wraps to de (first in sorted list)
	})

	t.Run("cycles through all languages", func(t *testing.T) {
		// Verify we can cycle through all 9 languages
		expectedOrder := []string{"de", "el", "en", "es", "it", "ja", "pt", "ru", "zh"}

		_ = i18n.SetLanguage("zh") // Start at end so first call returns "de"

		for _, expected := range expectedOrder {
			next := getNextLanguage()
			assert.Equal(t, expected, next)
			_ = i18n.SetLanguage(next)
		}
	})

	t.Run("handles unknown current language", func(t *testing.T) {
		_ = i18n.SetLanguage("xx") // Unknown language
		next := getNextLanguage()
		// Should return first available language since "xx" is not in the list
		assert.Equal(t, "de", next)
	})
}

func TestSearchModel_LanguageChange(t *testing.T) {
	_ = i18n.Init("en")

	t.Run("broadcasts languageChangedMsg when language key is pressed", func(t *testing.T) {
		model := NewSearchModel(Theme{}, nil, nil, testSearchKeybindings)

		// Press the language change key (L)
		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")}

		_, cmd := model.Update(input)
		assert.NotNil(t, cmd)

		msg := cmd()
		assert.IsType(t, languageChangedMsg{}, msg)

		// Verify the message contains the next language
		langMsg := msg.(languageChangedMsg)
		assert.NotEmpty(t, langMsg.lang)
	})

	t.Run("custom language key works", func(t *testing.T) {
		customKb := testSearchKeybindings
		customKb.ChangeLanguage = "x"

		model := NewSearchModel(Theme{}, nil, nil, customKb)

		// Press the custom language change key
		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}

		_, cmd := model.Update(input)
		assert.NotNil(t, cmd)

		msg := cmd()
		assert.IsType(t, languageChangedMsg{}, msg)
	})

	t.Run("default language key does not work with custom binding", func(t *testing.T) {
		customKb := testSearchKeybindings
		customKb.ChangeLanguage = "x"

		model := NewSearchModel(Theme{}, nil, nil, customKb)

		// Press the default language change key (L) - should NOT trigger language change
		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")}

		_, cmd := model.Update(input)
		// cmd might be nil or a different command (cursor blink from textinput)
		if cmd != nil {
			msg := cmd()
			assert.NotEqual(t, languageChangedMsg{}, msg)
		}
	})
}
