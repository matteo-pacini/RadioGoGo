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
		model := NewSelectorModel(Theme{}, "Title", items, 0)
		assert.Equal(t, items[0], model.Selection())
	})

	t.Run("selection returns the correct item after changing selection", func(t *testing.T) {
		model := NewSelectorModel(Theme{}, "Title", items, 0)
		model.selection = 1
		assert.Equal(t, items[1], model.Selection())
	})

	t.Run("focused returns false by default", func(t *testing.T) {
		model := NewSelectorModel(Theme{}, "Title", items, 0)
		assert.False(t, model.Focused())
	})

	t.Run("focus sets focus to true", func(t *testing.T) {
		model := NewSelectorModel(Theme{}, "Title", items, 0)
		model.Focus()
		assert.True(t, model.Focused())
	})

	t.Run("blur sets focus to false", func(t *testing.T) {
		model := NewSelectorModel(Theme{}, "Title", items, 0)
		model.Blur()
		assert.False(t, model.Focused())
	})

	t.Run("view returns the correct string", func(t *testing.T) {
		model := NewSelectorModel(Theme{}, "Title", items, 0)
		expected := "Title\n\n  [•] Item 1\n  [ ] Item 2\n  [ ] Item 3\n"
		assert.Equal(t, expected, model.View())
	})

	t.Run("view returns the correct string with a different selection", func(t *testing.T) {
		model := NewSelectorModel(Theme{}, "Title", items, 0)
		model.selection = 2
		expected := "Title\n\n  [ ] Item 1\n  [ ] Item 2\n  [•] Item 3\n"
		assert.Equal(t, expected, model.View())
	})

	t.Run("view returns the correct string on focus", func(t *testing.T) {
		model := NewSelectorModel(Theme{}, "Title", items, 0)
		model.Focus()
		expected := "Title\n\n> [•] Item 1\n  [ ] Item 2\n  [ ] Item 3\n"
		assert.Equal(t, expected, model.View())
	})

	t.Run("view returns the correct string with a different selection on focus", func(t *testing.T) {
		model := NewSelectorModel(Theme{}, "Title", items, 0)
		model.Focus()
		model.selection = 2
		expected := "Title\n\n  [ ] Item 1\n  [ ] Item 2\n> [•] Item 3\n"
		assert.Equal(t, expected, model.View())
	})

	t.Run("up key moves the selection up", func(t *testing.T) {

		model := NewSelectorModel(Theme{}, "Title", items, 2)
		model.Focus()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")}

		model, _ = model.Update(msg)

		assert.Equal(t, 1, model.selection)

	})

	t.Run("up key does not move the selection up if we're at index zero", func(t *testing.T) {

		model := NewSelectorModel(Theme{}, "Title", items, 0)
		model.Focus()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")}

		model, _ = model.Update(msg)

		assert.Equal(t, 0, model.selection)

	})

	t.Run("down key moves the selection down", func(t *testing.T) {

		model := NewSelectorModel(Theme{}, "Title", items, 1)
		model.Focus()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")}

		model, _ = model.Update(msg)

		assert.Equal(t, 2, model.selection)

	})

	t.Run("down key does not move selection down if we're at max index", func(t *testing.T) {

		model := NewSelectorModel(Theme{}, "Title", items, 2)
		model.Focus()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")}

		model, _ = model.Update(msg)

		assert.Equal(t, 2, model.selection)

	})

}
