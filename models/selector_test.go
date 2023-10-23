package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

type MockRenderable struct {
	value string
}

func (m MockRenderable) Render() string {
	return m.value
}

func TestSelectorModel(t *testing.T) {

	items := []MockRenderable{
		{value: "Item 1"},
		{value: "Item 2"},
		{value: "Item 3"},
	}

	t.Run("selection returns the correct item", func(t *testing.T) {
		model := NewSelectorModel("Title", items, 0)
		assert.Equal(t, items[0], model.Selection())
	})

	t.Run("selection returns the correct item after changing selection", func(t *testing.T) {
		model := NewSelectorModel("Title", items, 0)
		model.selection = 1
		assert.Equal(t, items[1], model.Selection())
	})

	t.Run("focused returns false by default", func(t *testing.T) {
		model := NewSelectorModel("Title", items, 0)
		assert.False(t, model.Focused())
	})

	t.Run("focus sets focus to true", func(t *testing.T) {
		model := NewSelectorModel("Title", items, 0)
		model.Focus()
		assert.True(t, model.Focused())
	})

	t.Run("blur sets focus to false", func(t *testing.T) {
		model := NewSelectorModel("Title", items, 0)
		model.Blur()
		assert.False(t, model.Focused())
	})

	t.Run("view returns the correct string", func(t *testing.T) {
		model := NewSelectorModel("Title", items, 0)
		expected := "Title\n\n  [•] Item 1\n  [ ] Item 2\n  [ ] Item 3\n"
		assert.Equal(t, expected, model.View())
	})

	t.Run("view returns the correct string with a different selection", func(t *testing.T) {
		model := NewSelectorModel("Title", items, 0)
		model.selection = 2
		expected := "Title\n\n  [ ] Item 1\n  [ ] Item 2\n  [•] Item 3\n"
		assert.Equal(t, expected, model.View())
	})

	t.Run("view returns the correct string on focus", func(t *testing.T) {
		model := NewSelectorModel("Title", items, 0)
		model.Focus()
		expected := "Title\n\n> [•] Item 1\n  [ ] Item 2\n  [ ] Item 3\n"
		assert.Equal(t, expected, model.View())
	})

	t.Run("view returns the correct string with a different selection on focus", func(t *testing.T) {
		model := NewSelectorModel("Title", items, 0)
		model.Focus()
		model.selection = 2
		expected := "Title\n\n  [ ] Item 1\n  [ ] Item 2\n> [•] Item 3\n"
		assert.Equal(t, expected, model.View())
	})

	t.Run("up key moves the selection up", func(t *testing.T) {

		model := NewSelectorModel("Title", items, 2)
		model.Focus()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")}

		model, _ = model.Update(msg)

		assert.Equal(t, 1, model.selection)

	})

	t.Run("up key does not move the selection up if we're at index zero", func(t *testing.T) {

		model := NewSelectorModel("Title", items, 0)
		model.Focus()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")}

		model, _ = model.Update(msg)

		assert.Equal(t, 0, model.selection)

	})

	t.Run("down key moves the selection down", func(t *testing.T) {

		model := NewSelectorModel("Title", items, 1)
		model.Focus()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")}

		model, _ = model.Update(msg)

		assert.Equal(t, 2, model.selection)

	})

	t.Run("down key does not move selection down if we're at max index", func(t *testing.T) {

		model := NewSelectorModel("Title", items, 2)
		model.Focus()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")}

		model, _ = model.Update(msg)

		assert.Equal(t, 2, model.selection)

	})

}
