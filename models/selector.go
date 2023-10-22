package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

type SelectorRenderable interface {
	Render() string
}

type SelectorModel[T SelectorRenderable] struct {
	title     string
	items     []T
	selection int
	focus     bool
}

func NewSelectorModel[T SelectorRenderable](title string, items []T, initialSelection int) SelectorModel[T] {
	return SelectorModel[T]{
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

	v := StyleSetSectionTitle(m.title) + "\n\n"

	for i, item := range m.items {
		if i == m.selection {
			if m.focus {
				v += "> [*] "
			} else {
				v += "  [*] "
			}
		} else {
			v += "  [ ] "
		}
		v += item.Render()
		v += "\n"
	}

	return v

}
