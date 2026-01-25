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

	"github.com/google/uuid"
	"github.com/zi0p4tch0/radiogogo/api"
	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/i18n"
	"github.com/zi0p4tch0/radiogogo/playback"
	"github.com/zi0p4tch0/radiogogo/storage"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type stationsViewMode int

const (
	viewModeSearchResults stationsViewMode = iota
	viewModeBookmarks
)

// StationsModel handles the display and interaction with a list of radio stations.
// It manages playback, volume control, bookmarks, and hidden stations.
type StationsModel struct {
	theme       Theme
	keybindings config.Keybindings

	stations              []common.Station
	stationsTable         table.Model
	currentStation        common.Station
	currentStationSpinner spinner.Model
	volume                int
	err                   string

	// Volume change debouncing
	pendingVolumeChangeID int64
	volumeChangePending   bool

	// View mode and storage
	viewMode stationsViewMode
	storage  storage.StationStorageService

	// Saved state for returning from bookmarks view
	savedStations []common.Station
	savedCursor   int

	// Hidden modal state
	showHiddenModal   bool
	hiddenStations    []common.Station
	hiddenModalCursor int
	needsRefetch      bool

	// Last search query for refetching
	lastQuery     common.StationQuery
	lastQueryText string

	browser         api.RadioBrowserService
	playbackManager playback.PlaybackManagerService
	width           int
	height          int

	// Temporary success message (shown in green)
	successMsg string
}

// NewStationsModel creates a new StationsModel with the given dependencies and stations.
func NewStationsModel(
	theme Theme,
	browser api.RadioBrowserService,
	playbackManager playback.PlaybackManagerService,
	storage storage.StationStorageService,
	stations []common.Station,
	viewMode stationsViewMode,
	lastQuery common.StationQuery,
	lastQueryText string,
	keybindings config.Keybindings,
) StationsModel {

	return StationsModel{
		theme:           theme,
		keybindings:     keybindings,
		stations:        stations,
		stationsTable:   newStationsTableModel(theme, stations, storage),
		volume:          playbackManager.VolumeDefault(),
		viewMode:        viewMode,
		storage:         storage,
		browser:         browser,
		playbackManager: playbackManager,
		lastQuery:       lastQuery,
		lastQueryText:   lastQueryText,
	}
}

func newStationsTableModel(theme Theme, stations []common.Station, storage storage.StationStorageService) table.Model {

	rows := make([]table.Row, len(stations))
	for i, station := range stations {
		name := station.Name
		if storage != nil && storage.IsBookmarked(station.StationUuid) {
			name = "⭐ " + name
		}
		status := "✗"
		if station.LastCheckOk {
			status = "✓"
		}
		rows[i] = table.Row{
			name,
			station.CountryCode,
			fmt.Sprintf("%d", station.Bitrate),
			station.Codec,
			fmt.Sprintf("%d", station.ClickCount),
			fmt.Sprintf("%d", station.Votes),
			status,
		}
	}

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: i18n.T("header_name"), Width: 30},
			{Title: i18n.T("header_country"), Width: 10},
			{Title: i18n.T("header_bitrate"), Width: 8},
			{Title: i18n.T("header_codecs"), Width: 10},
			{Title: i18n.T("header_clicks"), Width: 8},
			{Title: i18n.T("header_votes"), Width: 8},
			{Title: i18n.T("header_status"), Width: 6},
		}),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	t.SetStyles(theme.StationsTableStyle)

	return t

}

// Init initializes the StationsModel and returns the initial command.
func (m StationsModel) Init() tea.Cmd {
	return tea.Batch(
		updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false, m.keybindings),
		func() tea.Msg {
			return stationCursorMovedMsg{
				offset:        m.stationsTable.Cursor(),
				totalStations: len(m.stations),
			}
		},
	)
}

// Update handles incoming messages and updates the model state accordingly.
// It delegates to specialized handlers for different message categories.
func (m StationsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Try each handler in turn - they return (handled, model, cmd)
	if handled, newM, cmd := m.handlePlaybackMessages(msg); handled {
		return newM, cmd
	}

	if handled, newM, cmd := m.handleErrorMessages(msg); handled {
		return newM, cmd
	}

	if handled, newM, cmd := m.handleVolumeMessages(msg); handled {
		return newM, cmd
	}

	if handled, newM, cmd := m.handleRecordingMessages(msg); handled {
		return newM, cmd
	}

	if handled, newM, cmd := m.handleBookmarkMessages(msg); handled {
		return newM, cmd
	}

	if handled, newM, cmd := m.handleHiddenStationMessages(msg); handled {
		return newM, cmd
	}

	if handled, newM, cmd := m.handleVoteMessages(msg); handled {
		return newM, cmd
	}

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		handled, newM, cmd := m.handleKeyMessage(keyMsg)
		if handled {
			return newM, cmd
		}
		// Navigation keys return cmd but don't fully handle the message
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m = newM
	}

	// Update spinner if playing
	if m.playbackManager.IsPlaying() {
		newSpinner, cmd := m.currentStationSpinner.Update(msg)
		m.currentStationSpinner = newSpinner
		cmds = append(cmds, cmd)
	}

	// Update table
	newStationsTable, cmd := m.stationsTable.Update(msg)
	m.stationsTable = newStationsTable
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// buildStatusBar returns the styled status bar string.
// Priority: success message > error message > now playing > default.
func (m StationsModel) buildStatusBar() string {
	if m.successMsg != "" {
		return m.theme.SuccessText.Render(m.successMsg)
	} else if m.err != "" {
		return m.theme.ErrorText.Render(m.err)
	} else if m.currentStation.StationUuid != uuid.Nil {
		return m.currentStationSpinner.View() + " " +
			m.theme.SecondaryText.Bold(true).Render(i18n.Tf("now_playing", map[string]interface{}{"Name": m.currentStation.Name})) +
			" " + m.currentStationSpinner.View()
	}
	return m.theme.TertiaryText.Render(i18n.T("select_station"))
}

// View renders the stations view.
func (m StationsModel) View() string {

	var v string
	var tableHeight int
	if len(m.stations) == 0 {
		var emptyMsg string
		if m.viewMode == viewModeBookmarks {
			emptyMsg = i18n.Tf("no_bookmarks", map[string]interface{}{"BookmarksKey": m.keybindings.BookmarksView})
		} else {
			emptyMsg = i18n.T("no_stations")
		}
		emptyContent := m.theme.SecondaryText.Bold(true).Render(emptyMsg)
		v = "\n" + emptyContent
		tableHeight = lipgloss.Height(emptyContent)
	} else {
		tableView := m.stationsTable.View()
		v = "\n" + tableView
		tableHeight = lipgloss.Height(tableView)
	}

	// Build status bar and calculate its height
	statusBar := m.buildStatusBar()
	statusHeight := lipgloss.Height(statusBar)

	// Calculate table area height (space for table + filler)
	// m.height = terminal - 3 (1 header + 2 bottom bars)
	// Layout: 1 (leading \n) + tableArea + 1 (blank before status) + statusHeight + 1 (space before bottom bar)
	// tableArea = m.height - 3 - statusHeight
	tableAreaHeight := m.height - 3 - statusHeight
	if tableAreaHeight < 1 {
		tableAreaHeight = 1
	}

	// Filler = tableArea - actual table content
	fillerHeight := tableAreaHeight - tableHeight
	if fillerHeight < 0 {
		fillerHeight = 0
	}

	// Add filler, then blank line before status, then status bar, then blank line after
	v += RenderFiller(fillerHeight)
	v += "\n\n" + statusBar // First \n terminates filler/table, second \n creates blank line
	v += "\n\n" // First \n terminates status, second \n creates blank line before bottom bar

	// Render hidden modal if showing
	if m.showHiddenModal {
		return m.renderWithModal(v)
	}

	return v
}

// SetWidthAndHeight updates the dimensions of the stations view.
func (m *StationsModel) SetWidthAndHeight(width int, height int) {
	m.width = width
	m.height = height
	m.stationsTable.SetWidth(width)

	// Calculate actual status bar height (can wrap to multiple lines)
	statusBar := m.buildStatusBar()
	statusHeight := lipgloss.Height(statusBar)
	if statusHeight < 1 {
		statusHeight = 1
	}

	// Table height calculation for viewport (max visible rows):
	// m.height = terminal - 3 (from model_handlers: 1 header + 2 bottom bars)
	// Layout: 1 (space after header) + table + filler + 1 (space before status) + statusHeight + 1 (space before bottom bar)
	// The table gets the maximum available space. View() adds filler between table and status bar
	// to keep status bar at a consistent position from the bottom.
	tableHeight := height - 3 - statusHeight
	if tableHeight < 1 {
		tableHeight = 1
	}
	m.stationsTable.SetHeight(tableHeight)
}

// IsModalShowing returns true if a modal dialog is currently displayed.
func (m StationsModel) IsModalShowing() bool {
	return m.showHiddenModal
}
