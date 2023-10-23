// Copyright (c) 2023 Matteo Pacini
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
	"radiogogo/common"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestSearchModel_Init(t *testing.T) {

	t.Run("starts blinking the input field cursor", func(t *testing.T) {

		model := NewSearchModel()

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

		model := NewSearchModel()

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

		expectedCommands := []string{"q: quit", "tab: cycle focus", "enter: search"}

		assert.Equal(t, expectedCommands, commands)

	})

}

func TestSearchModel_Update(t *testing.T) {

	t.Run("does not broadcast quitMsg when 'q' is pressed and textarea is focused", func(t *testing.T) {

		model := NewSearchModel()

		model.inputModel.Focus()
		model.querySelector.Blur()

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}

		_, cmd := model.Update(input)
		assert.NotNil(t, cmd)

		msg := cmd()

		assert.IsType(t, tea.BatchMsg{}, msg)

	})

	t.Run("broadcasts a quitMsg when 'q' is pressed and textarea is not focused", func(t *testing.T) {

		model := NewSearchModel()

		model.inputModel.Blur()
		model.querySelector.Focus()

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}

		_, cmd := model.Update(input)
		assert.NotNil(t, cmd)

		msg := cmd()

		assert.IsType(t, quitMsg{}, msg)

	})

	t.Run("broadcasts a switchToLoadingModelMsg when 'enter' is pressed, propagating text area value", func(t *testing.T) {

		model := NewSearchModel()
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

		model := NewSearchModel()
		model.inputModel.SetValue("fancy value")
		model.inputModel.Blur()

		input := tea.KeyMsg{Type: tea.KeyEnter}

		_, cmd := model.Update(input)
		assert.Nil(t, cmd)

	})

	t.Run("cycles focused input when 'tab' is pressed and updates bottom bar", func(t *testing.T) {

		model := NewSearchModel()

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
	expectedCommands := []string{"q: quit", "tab: cycle focus", "enter: search"}

	msg := updateCommandsForTextfieldFocus()

	updateMsg, ok := msg.(bottomBarUpdateMsg)

	assert.True(t, ok)
	assert.Equal(t, expectedCommands, updateMsg.commands)
}

func TestUpdateCommandsForSelectorFocus(t *testing.T) {
	expectedCommands := []string{"q: quit", "tab: cycle focus", "↑/↓: change filter"}

	msg := updateCommandsForSelectorFocus()

	updateMsg, ok := msg.(bottomBarUpdateMsg)

	assert.True(t, ok)
	assert.Equal(t, expectedCommands, updateMsg.commands)
}
