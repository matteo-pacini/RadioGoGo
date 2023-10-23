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
	"github.com/zi0p4tch0/radiogogo/api"
	"github.com/zi0p4tch0/radiogogo/common"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoadingModel struct {
	spinnerModel spinner.Model
	query        common.StationQuery
	queryText    string
	width        int
	height       int

	browser api.RadioBrowserService
}

func NewLoadingModel(browser api.RadioBrowserService, query common.StationQuery, queryText string) LoadingModel {

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(secondaryColor))

	return LoadingModel{
		spinnerModel: s,
		query:        query,
		queryText:    queryText,
		browser:      browser,
	}

}

func (m LoadingModel) Init() tea.Cmd {
	return tea.Batch(m.spinnerModel.Tick, searchStations(m.browser, m.query, m.queryText))
}

func (m LoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newSpinnerModel, cmd := m.spinnerModel.Update(msg)
	m.spinnerModel = newSpinnerModel
	return m, cmd
}

func (m LoadingModel) View() string {
	return "\n" + m.spinnerModel.View() + " Fetching radio stations..."
}

// Commands

func searchStations(browser api.RadioBrowserService, query common.StationQuery, queryText string) tea.Cmd {
	return func() tea.Msg {
		stations, err := browser.GetStations(query, queryText, "votes", true, 0, 100, true)
		if err != nil {
			return switchToErrorModelMsg{err: err.Error()}
		}
		return switchToStationsModelMsg{stations: stations}
	}
}

func (m *LoadingModel) SetWidthAndHeight(width int, height int) {
	m.width = width
	m.height = height
}
