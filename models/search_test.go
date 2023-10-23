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
