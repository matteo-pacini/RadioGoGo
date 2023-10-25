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
	"fmt"
	"strings"

	"github.com/zi0p4tch0/radiogogo/assets"
	"github.com/zi0p4tch0/radiogogo/common"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SearchModel struct {
	theme         Theme
	inputModel    textinput.Model
	querySelector SelectorModel[common.StationQuery]
	width         int
	height        int
}

func NewSearchModel(theme Theme) SearchModel {
	i := textinput.New()
	i.Placeholder = "Name"
	i.Width = 30
	i.TextStyle = theme.Text
	i.PlaceholderStyle = theme.TertiaryText
	i.Focus()

	selector := NewSelectorModel[common.StationQuery](
		theme,
		"Filter:",
		[]common.StationQuery{
			common.StationQueryByName,
			common.StationQueryByNameExact,
			common.StationQueryByCodec,
			common.StationQueryByCodecExact,
			common.StationQueryByCountry,
			common.StationQueryByCountryExact,
			common.StationQueryByCountryCodeExact,
			common.StationQueryByState,
			common.StationQueryByStateExact,
			common.StationQueryByLanguage,
			common.StationQueryByLanguageExact,
			common.StationQueryByTag,
			common.StationQueryByTagExact,
		},
		0,
	)

	return SearchModel{
		theme:         theme,
		inputModel:    i,
		querySelector: selector,
	}

}

// Commands

func updateCommandsForTextfieldFocus() tea.Msg {
	return bottomBarUpdateMsg{
		commands: []string{"q: quit", "tab: cycle focus", "enter: search"},
	}
}

func updateCommandsForSelectorFocus() tea.Msg {
	return bottomBarUpdateMsg{
		commands: []string{"q: quit", "tab: cycle focus", "↑/↓: change filter"},
	}
}

// Bubbletea

func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, updateCommandsForTextfieldFocus)
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.inputModel.Focused() {
				m.inputModel.Blur()
				m.querySelector.Focus()
				return m, updateCommandsForSelectorFocus
			} else {
				m.inputModel.Focus()
				m.querySelector.Blur()
				return m, updateCommandsForTextfieldFocus
			}
		case "q":
			if !m.inputModel.Focused() {
				return m, quitCmd
			}
		case "enter":
			if !m.inputModel.Focused() {
				return m, nil
			}
			return m, func() tea.Msg {
				return switchToLoadingModelMsg{
					query:     m.querySelector.Selection(),
					queryText: m.inputModel.Value(),
				}
			}
		}
	}

	var cmds []tea.Cmd

	newInputModel, inputCmd := m.inputModel.Update(msg)
	m.inputModel = newInputModel

	if inputCmd != nil {
		cmds = append(cmds, inputCmd)
	}

	newSelectorModel, selectorCmd := m.querySelector.Update(msg)
	m.querySelector = newSelectorModel

	if selectorCmd != nil {
		cmds = append(cmds, selectorCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m SearchModel) View() string {

	searchType := m.querySelector.Selection().Render()
	searchType = strings.ToLower(searchType)

	rightOfLogoStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	rightV := rightOfLogoStyle.Render(
		fmt.Sprintf("\n%s\n\n%s\n\n%s\n%s",
			m.theme.SecondaryText.Render(fmt.Sprint("Search radio ", searchType)),
			m.inputModel.View(),
			m.querySelector.View(),
			m.theme.TertiaryText.Render(m.querySelector.Selection().ExampleString()),
		))

	leftV := fmt.Sprintf(
		"\n%s\n\n",
		assets.Logo,
	)

	v := lipgloss.JoinHorizontal(lipgloss.Top, leftV, rightV)

	return v
}

func (m *SearchModel) SetWidthAndHeight(width int, height int) {
	m.width = width
	m.height = height
}
