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
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/i18n"

	tea "github.com/charmbracelet/bubbletea"
)

// handleGlobalMessages handles messages that apply regardless of current state.
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m Model) handleGlobalMessages(msg tea.Msg) (bool, Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stationCursorMovedMsg:
		m.headerModel.totalStations = msg.totalStations
		m.headerModel.stationOffset = msg.offset
		return true, m, nil

	case playbackStatusMsg:
		newHeaderModel, cmd := m.headerModel.Update(msg)
		m.headerModel = newHeaderModel.(HeaderModel)
		return true, m, cmd

	case recordingStatusMsg:
		newHeaderModel, cmd := m.headerModel.Update(msg)
		m.headerModel = newHeaderModel.(HeaderModel)
		return true, m, cmd

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)

	case quitMsg:
		return true, m, tea.Quit

	case bottomBarUpdateMsg:
		m.bottomBarCommands = msg.commands
		m.bottomBarSecondaryCommands = msg.secondaryCommands
		return true, m, nil

	case languageChangedMsg:
		return m.handleLanguageChange(msg)
	}
	return false, m, nil
}

// handleWindowResize handles terminal resize events.
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (bool, Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.headerModel.width = msg.Width

	tooSmall := m.width < minTerminalWidth || m.height < minTerminalHeight

	if tooSmall && m.state != terminalTooSmallState {
		m.previousState = m.state
		m.state = terminalTooSmallState
		return true, m, nil
	}

	if !tooSmall && m.state == terminalTooSmallState {
		m.state = m.previousState
	}

	// Update child model dimensions based on current state
	switch m.state {
	case searchState:
		childHeight := m.height - 2 // 1 header + 1 bottom bar row
		m.searchModel.SetWidthAndHeight(m.width, childHeight)
	case loadingState:
		childHeight := m.height - 2 // 1 header + 1 bottom bar row
		m.loadingModel.SetWidthAndHeight(m.width, childHeight)
	case stationsState:
		childHeight := m.height - 3 // 1 header + 2 bottom bar rows
		m.stationsModel.SetWidthAndHeight(m.width, childHeight)
	case errorState:
		childHeight := m.height - 2 // 1 header + 1 bottom bar row
		m.errorModel.SetWidthAndHeight(m.width, childHeight)
	}
	return true, m, nil
}

// handleLanguageChange handles language change events.
func (m Model) handleLanguageChange(msg languageChangedMsg) (bool, Model, tea.Cmd) {
	m.config.Language = msg.lang
	_ = m.config.Save(config.ConfigFile())
	_ = i18n.SetLanguage(msg.lang)

	// Recreate search model to refresh all strings
	m.searchModel = NewSearchModel(m.theme, m.browser, m.storage, m.config.Keybindings)
	m.searchModel.SetWidthAndHeight(m.width, m.height-2)
	return true, m, m.searchModel.Init()
}

// handleStateTransitions handles messages that trigger state changes.
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m Model) handleStateTransitions(msg tea.Msg) (bool, Model, tea.Cmd) {
	switch msg := msg.(type) {
	case switchToSearchModelMsg:
		if m.playbackManager != nil {
			m.playbackManager.StopStation()
		}
		m.headerModel.showOffset = false
		m.headerModel.playbackStatus = PlaybackIdle
		m.headerModel.isRecording = false
		m.bottomBarSecondaryCommands = nil
		m.searchModel = NewSearchModel(m.theme, m.browser, m.storage, m.config.Keybindings)
		m.searchModel.SetWidthAndHeight(m.width, m.height-2)
		m.state = searchState
		return true, m, m.searchModel.Init()

	case switchToLoadingModelMsg:
		m.headerModel.showOffset = false
		m.bottomBarSecondaryCommands = nil
		m.loadingModel = NewLoadingModel(m.theme, m.browser, msg.query, msg.queryText)
		m.loadingModel.SetWidthAndHeight(m.width, m.height-2)
		m.state = loadingState
		return true, m, m.loadingModel.Init()

	case switchToStationsModelMsg:
		m.headerModel.showOffset = true
		filteredStations := filterHiddenStations(msg.stations, m.storage)
		m.stationsModel = NewStationsModel(m.theme, m.browser, m.playbackManager, m.storage, filteredStations, viewModeSearchResults, msg.query, msg.queryText, m.config.Keybindings)
		m.stationsModel.SetWidthAndHeight(m.width, m.height-3)
		m.state = stationsState
		return true, m, m.stationsModel.Init()

	case switchToBookmarksMsg:
		m.headerModel.showOffset = true
		m.stationsModel = NewStationsModel(m.theme, m.browser, m.playbackManager, m.storage, msg.stations, viewModeBookmarks, "", "", m.config.Keybindings)
		m.stationsModel.SetWidthAndHeight(m.width, m.height-3)
		m.state = stationsState
		return true, m, m.stationsModel.Init()

	case switchToErrorModelMsg:
		m.headerModel.showOffset = false
		m.bottomBarSecondaryCommands = nil
		m.errorModel = NewErrorModel(m.theme, msg.err, msg.recoverable, m.config.Keybindings)
		m.errorModel.SetWidthAndHeight(m.width, m.height-2)
		m.state = errorState
		return true, m, m.errorModel.Init()
	}
	return false, m, nil
}

// delegateToCurrentState forwards messages to the currently active state's model.
// Returns (model, cmd).
func (m Model) delegateToCurrentState(msg tea.Msg) (Model, tea.Cmd) {
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
