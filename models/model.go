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
	"github.com/zi0p4tch0/radiogogo/api"
	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/i18n"
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
	stations  []common.Station
	query     common.StationQuery
	queryText string
}
type switchToBookmarksMsg struct {
	stations []common.Station
}

// UI messages

type bottomBarUpdateMsg struct {
	commands          []string
	secondaryCommands []string
}

// Language change message
type languageChangedMsg struct {
	lang string
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

	// Config
	config config.Config

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
	cfg config.Config,
	browser api.RadioBrowserService,
	playbackManager playback.PlaybackManagerService,
	storage storage.StationStorageService,
) Model {

	theme := NewTheme(cfg)

	return Model{
		config:          cfg,
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
	// Handle global messages (cursor, playback status, window resize, etc.)
	if handled, newM, cmd := m.handleGlobalMessages(msg); handled {
		return newM, cmd
	}

	// Handle state transitions
	if handled, newM, cmd := m.handleStateTransitions(msg); handled {
		return newM, cmd
	}

	// Delegate to current state
	return m.delegateToCurrentState(msg)
}

// View renders the current UI based on the active state.
// Composes the header, current state's view, and bottom bar.
func (m Model) View() string {

	// Handle terminal too small state separately
	if m.state == terminalTooSmallState {
		message := m.theme.ErrorText.Bold(true).Render(i18n.T("terminal_too_small")) + "\n\n" +
			i18n.Tf("terminal_min_size", map[string]interface{}{"MinWidth": minTerminalWidth, "MinHeight": minTerminalHeight}) + "\n" +
			i18n.Tf("terminal_current_size", map[string]interface{}{"Width": m.width, "Height": m.height})

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
		currentView = "\n" + i18n.T("initializing")
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

	// Render bottom bar (one or two rows) - skip when modal is showing
	if m.state == stationsState && m.stationsModel.IsModalShowing() {
		// Don't render bottom bar when modal is open
	} else if len(m.bottomBarSecondaryCommands) > 0 {
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
