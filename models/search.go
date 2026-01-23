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
	"strings"

	"github.com/zi0p4tch0/radiogogo/api"
	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/i18n"
	"github.com/zi0p4tch0/radiogogo/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SearchModel struct {
	theme         Theme
	browser       api.RadioBrowserService
	storage       storage.StationStorageService
	keybindings   config.Keybindings
	inputModel    textinput.Model
	querySelector SelectorModel[common.StationQuery]
	width         int
	height        int
}

func NewSearchModel(theme Theme, browser api.RadioBrowserService, storage storage.StationStorageService, keybindings config.Keybindings) SearchModel {
	i := textinput.New()
	i.Placeholder = i18n.T("search_placeholder")
	i.Width = 30
	i.TextStyle = theme.Text
	i.PlaceholderStyle = theme.TertiaryText
	i.Focus()

	selector := NewSelectorModel[common.StationQuery](
		theme,
		i18n.T("filter_label"),
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
		browser:       browser,
		storage:       storage,
		keybindings:   keybindings,
		inputModel:    i,
		querySelector: selector,
	}

}

// Commands

func updateCommandsForTextfieldFocus(kb config.Keybindings) tea.Cmd {
	return func() tea.Msg {
		return bottomBarUpdateMsg{
			commands: []string{
				i18n.Tf("cmd_quit", map[string]interface{}{"Key": kb.Quit}),
				i18n.T("cmd_cycle_focus"),
				i18n.T("cmd_enter_search"),
				i18n.Tf("cmd_bookmarks", map[string]interface{}{"Key": kb.BookmarksView}),
				i18n.Tf("cmd_change_language", map[string]interface{}{"Key": kb.ChangeLanguage}),
				i18n.T("current_language"),
			},
		}
	}
}

func updateCommandsForSelectorFocus(kb config.Keybindings) tea.Cmd {
	return func() tea.Msg {
		return bottomBarUpdateMsg{
			commands: []string{
				i18n.Tf("cmd_quit", map[string]interface{}{"Key": kb.Quit}),
				i18n.T("cmd_cycle_focus"),
				i18n.T("cmd_change_filter"),
				i18n.Tf("cmd_bookmarks", map[string]interface{}{"Key": kb.BookmarksView}),
				i18n.Tf("cmd_change_language", map[string]interface{}{"Key": kb.ChangeLanguage}),
				i18n.T("current_language"),
			},
		}
	}
}

// getNextLanguage returns the next language in the available languages cycle.
func getNextLanguage() string {
	langs := i18n.AvailableLanguages()
	current := i18n.CurrentLanguage()

	for idx, lang := range langs {
		if lang == current {
			return langs[(idx+1)%len(langs)]
		}
	}
	return langs[0]
}

// Bubbletea

func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, updateCommandsForTextfieldFocus(m.keybindings))
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.inputModel.Focused() {
				m.inputModel.Blur()
				m.querySelector.Focus()
				return m, updateCommandsForSelectorFocus(m.keybindings)
			} else {
				m.inputModel.Focus()
				m.querySelector.Blur()
				return m, updateCommandsForTextfieldFocus(m.keybindings)
			}
		case m.keybindings.Quit:
			if !m.inputModel.Focused() {
				return m, quitCmd
			}
		case m.keybindings.BookmarksView:
			return m, fetchBookmarksForSearchCmd(m.browser, m.storage)
		case m.keybindings.ChangeLanguage:
			nextLang := getNextLanguage()
			return m, func() tea.Msg {
				return languageChangedMsg{lang: nextLang}
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

	v := fmt.Sprintf("\n%s\n\n%s\n\n%s\n%s\n",
		m.theme.SecondaryText.Render(i18n.Tf("search_title", map[string]interface{}{"Type": searchType})),
		m.inputModel.View(),
		m.querySelector.View(),
		m.theme.TertiaryText.Render(m.querySelector.Selection().ExampleString()),
	)

	return v
}

func (m *SearchModel) SetWidthAndHeight(width int, height int) {
	m.width = width
	m.height = height
}
