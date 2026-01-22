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

// Package models implements the TUI components and state machine for RadioGoGo.
// It uses BubbleTea's Elm-inspired architecture with the following states:
//   - bootState: Initialization, checks if playback (FFplay) is available
//   - searchState: User enters search criteria (name, country, codec, etc.)
//   - loadingState: Fetches stations from RadioBrowser API
//   - stationsState: Displays results in a table, allows selection and playback
//   - errorState: Shows error messages
//   - terminalTooSmallState: Displays when terminal is below minimum size
//
// State transitions are handled via typed messages (switchToXModelMsg) and
// the main Model coordinates between child models for each state.
package models

import (
	"fmt"

	"github.com/zi0p4tch0/radiogogo/api"
	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/playback"
	"github.com/zi0p4tch0/radiogogo/storage"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	minTerminalWidth  = 115
	minTerminalHeight = 29
)

type modelState int

const (
	bootState modelState = iota
	searchState
	errorState
	loadingState
	stationsState
	terminalTooSmallState
)

// State switching messages

type switchToErrorModelMsg struct {
	err         string
	recoverable bool
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
	commands          []string
	secondaryCommands []string
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
				err:         playbackManager.NotAvailableErrorString(),
				recoverable: false,
			}
		}
		return switchToSearchModelMsg{}
	}
}

// Model is the root BubbleTea model that coordinates the application state machine.
// It manages state transitions between search, loading, stations, and error views,
// and handles global messages like window resize and quit events.
type Model struct {

	// Theme
	theme Theme

	// Models
	headerModel                HeaderModel
	searchModel                SearchModel
	errorModel                 ErrorModel
	loadingModel               LoadingModel
	stationsModel              StationsModel
	bottomBarCommands          []string
	bottomBarSecondaryCommands []string

	// State
	state           modelState
	previousState   modelState
	width           int
	height          int
	browser         api.RadioBrowserService
	playbackManager playback.PlaybackManagerService
	storage         storage.StationStorageService
}

// NewDefaultModel creates a new Model with production dependencies (real API client,
// FFplay playback manager, and file-based storage). Returns an error if any
// dependency initialization fails.
func NewDefaultModel(config config.Config) (Model, error) {

	browser, err := api.NewRadioBrowser()
	if err != nil {
		return Model{}, err
	}

	playbackManager := playback.NewFFPlaybackManager()

	storageService, err := storage.NewFileStorage()
	if err != nil {
		return Model{}, err
	}

	return NewModel(config, browser, playbackManager, storageService), nil

}

// NewModel creates a new Model with the provided dependencies. This constructor
// is preferred for testing as it allows injecting mock implementations.
func NewModel(
	config config.Config,
	browser api.RadioBrowserService,
	playbackManager playback.PlaybackManagerService,
	storage storage.StationStorageService,
) Model {

	theme := NewTheme(config)

	return Model{
		theme:           theme,
		headerModel:     NewHeaderModel(theme, playbackManager),
		state:           bootState,
		browser:         browser,
		playbackManager: playbackManager,
		storage:         storage,
	}
}

// Init initializes the model by checking if playback is available.
// If FFplay is not found, transitions to error state; otherwise transitions to search state.
func (m Model) Init() tea.Cmd {
	return checkIfPlaybackIsPossibleCmd(m.playbackManager)
}

// Update handles incoming messages and manages state transitions.
// It processes global events (window resize, quit) and delegates state-specific
// messages to the appropriate child model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Top-level messages
	switch msg := msg.(type) {
	case stationCursorMovedMsg:
		m.headerModel.totalStations = msg.totalStations
		m.headerModel.stationOffset = msg.offset
		return m, nil
	case playbackStatusMsg:
		newHeaderModel, cmd := m.headerModel.Update(msg)
		m.headerModel = newHeaderModel.(HeaderModel)
		return m, cmd
	case recordingStatusMsg:
		newHeaderModel, cmd := m.headerModel.Update(msg)
		m.headerModel = newHeaderModel.(HeaderModel)
		return m, cmd
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.headerModel.width = msg.Width

		tooSmall := m.width < minTerminalWidth || m.height < minTerminalHeight

		if tooSmall && m.state != terminalTooSmallState {
			m.previousState = m.state
			m.state = terminalTooSmallState
			return m, nil
		}

		if !tooSmall && m.state == terminalTooSmallState {
			m.state = m.previousState
		}

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
		return m, nil
	case quitMsg:
		return m, tea.Quit
	case bottomBarUpdateMsg:
		m.bottomBarCommands = msg.commands
		m.bottomBarSecondaryCommands = msg.secondaryCommands
		return m, nil
	}

	// State transitions

	switch msg := msg.(type) {
	case switchToSearchModelMsg:
		// Stop any playing audio when returning to search
		if m.playbackManager != nil {
			m.playbackManager.StopStation()
		}
		m.headerModel.showOffset = false
		m.headerModel.playbackStatus = PlaybackIdle // Reset playback indicator
		m.headerModel.isRecording = false           // Reset recording indicator
		m.bottomBarSecondaryCommands = nil          // Clear two-row bar
		m.searchModel = NewSearchModel(m.theme)
		m.searchModel.SetWidthAndHeight(m.width, m.height-2) // 1 header + 1 bottom bar row
		m.state = searchState
		return m, m.searchModel.Init()
	case switchToLoadingModelMsg:
		m.headerModel.showOffset = false
		m.bottomBarSecondaryCommands = nil // Clear two-row bar
		m.loadingModel = NewLoadingModel(m.theme, m.browser, msg.query, msg.queryText)
		m.loadingModel.SetWidthAndHeight(m.width, m.height-2) // 1 header + 1 bottom bar row
		m.state = loadingState
		return m, m.loadingModel.Init()
	case switchToStationsModelMsg:
		m.headerModel.showOffset = true
		// Filter out hidden stations before displaying
		filteredStations := filterHiddenStations(msg.stations, m.storage)
		m.stationsModel = NewStationsModel(m.theme, m.browser, m.playbackManager, m.storage, filteredStations, viewModeSearchResults)
		m.stationsModel.SetWidthAndHeight(m.width, m.height-3) // 1 header + 2 bottom bar rows
		m.state = stationsState
		return m, m.stationsModel.Init()
	case switchToErrorModelMsg:
		m.headerModel.showOffset = false
		m.bottomBarSecondaryCommands = nil // Clear two-row bar
		m.errorModel = NewErrorModel(m.theme, msg.err, msg.recoverable)
		m.errorModel.SetWidthAndHeight(m.width, m.height-2) // 1 header + 1 bottom bar row
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

// View renders the current UI based on the active state.
// Composes the header, current state's view, and bottom bar.
func (m Model) View() string {

	// Handle terminal too small state separately
	if m.state == terminalTooSmallState {
		message := fmt.Sprintf(
			"%s\n\nMinimum size: %dx%d\nCurrent size: %dx%d",
			m.theme.ErrorText.Bold(true).Render("Terminal too small!"),
			minTerminalWidth, minTerminalHeight,
			m.width, m.height,
		)

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			message,
		)
	}

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

	// Render the current view
	view += currentView

	// Push the bottom bar at the bottom of the terminal (skip for stations - it handles its own height)
	if m.state != stationsState {
		// Measure the actual height of header + content combined
		// This accounts for trailing newline merge effects
		headerContentHeight := lipgloss.Height(view)
		bottomBarHeight := 1
		fillerHeight := CalculateFillerHeight(m.height, headerContentHeight, bottomBarHeight)
		view += RenderFiller(fillerHeight)
	}

	// Render bottom bar (one or two rows)
	if len(m.bottomBarSecondaryCommands) > 0 {
		view += m.theme.StyleTwoRowBottomBar(m.bottomBarCommands, m.bottomBarSecondaryCommands)
	} else {
		view += m.theme.StyleBottomBar(m.bottomBarCommands)
	}

	return view
}

// filterHiddenStations removes hidden stations from the list
func filterHiddenStations(stations []common.Station, storage storage.StationStorageService) []common.Station {
	if storage == nil {
		return stations
	}
	result := make([]common.Station, 0, len(stations))
	for _, s := range stations {
		if !storage.IsHidden(s.StationUuid) {
			result = append(result, s)
		}
	}
	return result
}
