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
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type StringRenderable interface {
	Render() string
}

type SelectorModel[T StringRenderable] struct {
	theme     Theme
	title     string
	items     []T
	selection int
	focus     bool
}

func NewSelectorModel[T StringRenderable](theme Theme, title string, items []T, initialSelection int) SelectorModel[T] {
	return SelectorModel[T]{
		theme:     theme,
		title:     title,
		items:     items,
		selection: initialSelection,
		focus:     false,
	}
}

// Selection

func (m SelectorModel[T]) Selection() T {
	return m.items[m.selection]
}

// Focus

func (m SelectorModel[T]) Focused() bool {
	return m.focus
}

func (m *SelectorModel[T]) Focus() {
	m.focus = true
}

func (m *SelectorModel[T]) Blur() {
	m.focus = false
}

// Bubbletea

func (m SelectorModel[T]) Init() tea.Cmd {
	return nil
}

func (m SelectorModel[T]) Update(msg tea.Msg) (SelectorModel[T], tea.Cmd) {

	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.selection > 0 {
				m.selection--
			}
		case "down":
			if m.selection < len(m.items)-1 {
				m.selection++
			}
		}
	}

	return m, nil
}

func (m SelectorModel[T]) View() string {

	v := m.theme.SecondaryText.Bold(true).Render(m.title) + "\n\n"

	for i, item := range m.items {
		if i == m.selection {
			if m.focus {
				v += fmt.Sprintf(
					"%s%s%s ",
					m.theme.Text.Render("> ["),
					m.theme.SecondaryText.Render("•"),
					m.theme.Text.Render("]"),
				)
			} else {
				v += fmt.Sprintf(
					"%s%s%s ",
					m.theme.Text.Render("  ["),
					m.theme.SecondaryText.Render("•"),
					m.theme.Text.Render("]"),
				)
			}
		} else {
			v += m.theme.Text.Render("  [ ] ")
		}
		v += m.theme.Text.Render(item.Render())
		v += "\n"
	}

	return v

}
