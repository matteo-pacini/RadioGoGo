package ui

import (
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

		expectedCommands := []string{"q: quit", "enter: search"}

		assert.Equal(t, expectedCommands, commands)

	})

}

func TestSearchModel_Update(t *testing.T) {

	t.Run("broadcasts a quitMsg whe 'q' is pressed", func(t *testing.T) {

		model := NewSearchModel()

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
			query: "fancy value",
		}, msg)

	})

}
