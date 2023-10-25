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

func TestErrorModel_Init(t *testing.T) {

	model := NewErrorModel(Theme{}, "this is an error")

	t.Run("broadcasts a quitTickMsg", func(t *testing.T) {

		cmd := model.Init()
		assert.NotNil(t, cmd)

		msg := cmd()
		assert.IsType(t, quitTickMsg{}, msg)

	})

}

func TestErrorModel_Update(t *testing.T) {

	model := NewErrorModel(Theme{}, "this is an error")

	t.Run("broadcasts a quitMsg whe 'q' is pressed", func(t *testing.T) {

		input := tea.Msg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

		_, cmd := model.Update(input)
		assert.NotNil(t, cmd)

		msg := cmd()

		assert.IsType(t, quitMsg{}, msg)

	})

	t.Run("broadcasts a quitTickMsg after receiving onem and increments the tick count", func(t *testing.T) {

		input := tea.Msg(quitTickMsg{})
		newModel, cmd := model.Update(input)

		assert.Equal(t, newModel.(ErrorModel).tickCount, 1)
		assert.NotNil(t, cmd)

		msg := cmd()
		assert.IsType(t, quitTickMsg{}, msg)

	})

	t.Run("quits after 30 ticks", func(t *testing.T) {

		model.tickCount = 29

		_, cmd := model.Update(quitTickMsg{})
		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, quitMsg{}, msg)

	})

}
