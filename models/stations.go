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
		rows[i] = table.Row{
			name,
			station.CountryCode,
			station.LanguagesCodes,
			station.Codec,
			fmt.Sprintf("%d", station.Votes),
		}
	}

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: i18n.T("header_name"), Width: 30},
			{Title: i18n.T("header_country"), Width: 10},
			{Title: i18n.T("header_languages"), Width: 15},
			{Title: i18n.T("header_codecs"), Width: 15},
			{Title: i18n.T("header_votes"), Width: 10},
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

// View renders the stations view.
func (m StationsModel) View() string {

	var v string
	if len(m.stations) == 0 {
		var emptyMsg string
		if m.viewMode == viewModeBookmarks {
			emptyMsg = i18n.Tf("no_bookmarks", map[string]interface{}{"BookmarksKey": m.keybindings.BookmarksView})
		} else {
			emptyMsg = i18n.T("no_stations")
		}
		v = fmt.Sprintf(
			"\n%s\n",
			m.theme.SecondaryText.Bold(true).Render(emptyMsg),
		)
	} else {
		v = "\n" + m.stationsTable.View() + "\n"
	}

	// Show status bar (with blank line above)
	var statusBar string
	if m.currentStation.StationUuid != uuid.Nil {
		statusBar = m.currentStationSpinner.View() + " " +
			m.theme.SecondaryText.Bold(true).Render(i18n.Tf("now_playing", map[string]interface{}{"Name": m.currentStation.Name})) +
			" " + m.currentStationSpinner.View()
	} else {
		statusBar = m.theme.TertiaryText.Render(i18n.T("select_station"))
	}
	v += "\n" + statusBar

	// Show error if any
	if m.err != "" {
		v += m.theme.ErrorText.Render(m.err) + "\n"
	}

	// Add filler to push bottom bar to the bottom
	// m.height is the space allocated for stations view (terminal height - header - bottom bar)
	// But due to trailing newline merge effects when this view is concatenated with bottom bar,
	// we need to add 1 extra line: actual needed = m.height + 1
	contentHeight := lipgloss.Height(v)
	fillerHeight := m.height - contentHeight + 1
	v += RenderFiller(fillerHeight)

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
	// Table height calculation:
	// - "\n" before table (1 line)
	// - "\n" after table (1 line)
	// - "\n" blank line + status bar (2 lines)
	// So: tableHeight = height - 4
	m.stationsTable.SetHeight(height - 4)
}

// IsModalShowing returns true if a modal dialog is currently displayed.
func (m StationsModel) IsModalShowing() bool {
	return m.showHiddenModal
}
