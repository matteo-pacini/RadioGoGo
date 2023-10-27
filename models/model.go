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
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/playback"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type modelState int

const (
	bootState modelState = iota
	searchState
	errorState
	loadingState
	stationsState
)

// State switching messages

type switchToErrorModelMsg struct {
	err string
}
type switchToSearchModelMsg struct {
}
type switchToLoadingModelMsg struct {
	query     common.StationQuery
	queryText string
}
type switchToStationsModelMsg struct {
	stations []common.Station
}

// UI messages

type bottomBarUpdateMsg struct {
	commands []string
}

// Quit message

type quitMsg struct{}

func quitCmd() tea.Msg {
	return quitMsg{}
}

// Commands

func checkIfPlaybackIsPossibleCmd(playbackManager playback.PlaybackManagerService) tea.Cmd {
	return func() tea.Msg {
		if !playbackManager.IsAvailable() {
			return switchToErrorModelMsg{
				err: playbackManager.NotAvailableErrorString(),
			}
		}
		return switchToSearchModelMsg{}
	}
}

// Model

type Model struct {

	// Theme
	theme Theme

	// Models
	headerModel       HeaderModel
	searchModel       SearchModel
	errorModel        ErrorModel
	loadingModel      LoadingModel
	stationsModel     StationsModel
	bottomBarCommands []string

	// State
	state           modelState
	width           int
	height          int
	browser         api.RadioBrowserService
	playbackManager playback.PlaybackManagerService
}

func NewDefaultModel(config config.Config) (Model, error) {

	browser, err := api.NewRadioBrowser()
	if err != nil {
		return Model{}, err
	}

	var playbackManager playback.PlaybackManagerService
	if config.PlaybackEngine == playback.FFPlay {
		playbackManager = playback.NewFFPlaybackManager()
	} else {
		playbackManager = playback.NewMPVbackManager()
	}

	return NewModel(config, browser, playbackManager), nil

}

func NewModel(
	config config.Config,
	browser api.RadioBrowserService,
	playbackManager playback.PlaybackManagerService,
) Model {

	theme := NewTheme(config)

	return Model{
		theme:           theme,
		headerModel:     NewHeaderModel(theme, playbackManager),
		state:           bootState,
		browser:         browser,
		playbackManager: playbackManager,
	}
}

func (m Model) Init() tea.Cmd {
	return checkIfPlaybackIsPossibleCmd(m.playbackManager)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Top-level messages
	switch msg := msg.(type) {
	case stationCursorMovedMsg:
		m.headerModel.totalStations = msg.totalStations
		m.headerModel.stationOffset = msg.offset
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.headerModel.width = msg.Width
		childHeight := m.height - 2 // 2 = header height + bottom bar height
		switch m.state {
		case searchState:
			m.searchModel.SetWidthAndHeight(m.width, childHeight)
		case loadingState:
			m.loadingModel.SetWidthAndHeight(m.width, childHeight)
		case stationsState:
			m.stationsModel.SetWidthAndHeight(m.width, childHeight)
		case errorState:
			m.errorModel.SetWidthAndHeight(m.width, childHeight)
		}
		return m, nil
	case quitMsg:
		return m, tea.Quit
	case bottomBarUpdateMsg:
		m.bottomBarCommands = msg.commands
		return m, nil
	}

	// State transitions

	childHeight := m.height - 2 // 2 = header height + bottom bar height

	switch msg := msg.(type) {
	case switchToSearchModelMsg:
		m.headerModel.showOffset = false
		m.searchModel = NewSearchModel(m.theme)
		m.searchModel.SetWidthAndHeight(m.width, childHeight)
		m.state = searchState
		return m, m.searchModel.Init()
	case switchToLoadingModelMsg:
		m.headerModel.showOffset = false
		m.loadingModel = NewLoadingModel(m.theme, m.browser, msg.query, msg.queryText)
		m.loadingModel.SetWidthAndHeight(m.width, childHeight)
		m.state = loadingState
		return m, m.loadingModel.Init()
	case switchToStationsModelMsg:
		m.headerModel.showOffset = true
		m.stationsModel = NewStationsModel(m.theme, m.browser, m.playbackManager, msg.stations)
		m.stationsModel.SetWidthAndHeight(m.width, childHeight)
		m.state = stationsState
		return m, m.stationsModel.Init()
	case switchToErrorModelMsg:
		m.headerModel.showOffset = false
		m.errorModel = NewErrorModel(m.theme, msg.err)
		m.errorModel.SetWidthAndHeight(m.width, childHeight)
		m.state = errorState
		return m, m.errorModel.Init()
	}

	// State handling

	switch m.state {
	case searchState:
		newSearchModel, cmd := m.searchModel.Update(msg)
		m.searchModel = newSearchModel.(SearchModel)
		return m, cmd
	case loadingState:
		newLoadingModel, cmd := m.loadingModel.Update(msg)
		m.loadingModel = newLoadingModel.(LoadingModel)
		return m, cmd
	case stationsState:
		newStationsModel, cmd := m.stationsModel.Update(msg)
		m.stationsModel = newStationsModel.(StationsModel)
		return m, cmd
	case errorState:
		newErrorModel, cmd := m.errorModel.Update(msg)
		m.errorModel = newErrorModel.(ErrorModel)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {

	var view string

	view = m.headerModel.View()

	var currentView string

	switch m.state {
	case bootState:
		currentView = "\nInitializing..."
	case searchState:
		currentView = m.searchModel.View()
	case loadingState:
		currentView = m.loadingModel.View()
	case stationsState:
		currentView = m.stationsModel.View()
	case errorState:
		currentView = m.errorModel.View()
	}

	currentViewHeight := lipgloss.Height(currentView)

	// Render the current view

	view += currentView

	// Push the bottom bar at the bottom of the terminal

	view += lipgloss.NewStyle().
		Height(m.height - currentViewHeight).
		Render()

	// Render bottom bar

	view += m.theme.StyleBottomBar(m.bottomBarCommands)

	return view
}
