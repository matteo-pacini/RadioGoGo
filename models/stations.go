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
	"strconv"
	"strings"

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

// formatNumber formats large numbers with K (thousands) or M (millions) suffix.
// Examples: 1234567 → "1.2M", 5678 → "5.7K", 999 → "999"
func formatNumber(n uint64) string {
	if n >= 1000000 {
		millions := float64(n) / 1000000.0
		return fmt.Sprintf("%.1fM", millions)
	} else if n >= 1000 {
		thousands := float64(n) / 1000.0
		return fmt.Sprintf("%.1fK", thousands)
	}
	return strconv.FormatUint(n, 10)
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// removeStationByUUID returns a new slice with the station matching the given UUID removed.
// If no station matches, the original slice contents are returned in a new slice.
func removeStationByUUID(stations []common.Station, stationUUID uuid.UUID) []common.Station {
	result := make([]common.Station, 0, len(stations))
	for _, s := range stations {
		if s.StationUuid != stationUUID {
			result = append(result, s)
		}
	}
	return result
}

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

	// Get the currently playing station (if any)
	currentStation := playbackManager.CurrentStation()

	return StationsModel{
		theme:           theme,
		keybindings:     keybindings,
		stations:        stations,
		stationsTable:   newStationsTableModel(theme, stations, storage, currentStation),
		volume:          playbackManager.VolumeDefault(),
		viewMode:        viewMode,
		storage:         storage,
		browser:         browser,
		playbackManager: playbackManager,
		lastQuery:       lastQuery,
		lastQueryText:   lastQueryText,
	}
}

func newStationsTableModel(theme Theme, stations []common.Station, storage storage.StationStorageService, currentStation common.Station) table.Model {

	rows := make([]table.Row, len(stations))
	for i, station := range stations {
		name := station.Name

		// Add now-playing indicator if this is the currently playing station
		if currentStation.StationUuid != uuid.Nil && station.StationUuid == currentStation.StationUuid {
			name = "▶ " + name
		}

		// Add bookmark star if bookmarked
		if storage != nil && storage.IsBookmarked(station.StationUuid) {
			name = "⭐ " + name
		}

		// Build quality string (bitrate + codec) with star indicators for quality tier
		quality := ""

		// Build the text content
		if station.Bitrate > 0 {
			quality = strconv.FormatUint(uint64(station.Bitrate), 10) + "k"
		}
		if station.Codec != "" {
			if quality != "" {
				quality += " "
			}
			quality += station.Codec
		}

		// Add quality tier stars as visual indicator
		if quality != "" {
			if station.Bitrate >= 256 {
				// High quality: 256+ kbps
				quality += " ★★"
			} else if station.Bitrate >= 128 {
				// Medium quality: 128-255 kbps
				quality += " ★"
			}
			// Low quality (<128 kbps): no stars
		} else {
			// No bitrate or codec available
			quality = "—"
		}

		// Format status
		status := "✗"
		if station.LastCheckOk {
			status = "✓"
		}

		rows[i] = table.Row{
			name,
			station.CountryCode,
			quality,
			formatNumber(uint64(station.ClickCount)),
			formatNumber(uint64(station.Votes)),
			status,
		}
	}

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: i18n.T("header_name"), Width: 35},
			{Title: i18n.T("header_country"), Width: 10},
			{Title: i18n.T("header_quality"), Width: 12},
			{Title: i18n.T("header_clicks"), Width: 10},
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

// renderNowPlayingBox creates a styled multi-line "Now Playing" box with station details.
// The box displays station name, bitrate, codec, listener count on line 1,
// and country, tags, and bookmark status on line 2.
func (m StationsModel) renderNowPlayingBox() string {
	station := m.currentStation

	// Line 1: ▶ Station Name • 128 kbps MP3 • 🎧 45.2K listeners
	line1Parts := []string{
		"▶ " + station.Name,
	}

	// Add bitrate and codec if available
	if station.Bitrate > 0 && station.Codec != "" {
		line1Parts = append(line1Parts, fmt.Sprintf("%d kbps %s", station.Bitrate, strings.ToUpper(station.Codec)))
	} else if station.Bitrate > 0 {
		line1Parts = append(line1Parts, fmt.Sprintf("%d kbps", station.Bitrate))
	} else if station.Codec != "" {
		line1Parts = append(line1Parts, strings.ToUpper(station.Codec))
	}

	// Add click count (listeners)
	if station.ClickCount > 0 {
		line1Parts = append(line1Parts, "🎧 "+formatNumber(station.ClickCount)+" listeners")
	}

	line1 := strings.Join(line1Parts, " • ")

	// Line 2: 📍 Country • tag1, tag2, tag3 • ⭐ Bookmarked
	line2Parts := []string{}

	// Add country
	if station.CountryCode != "" {
		line2Parts = append(line2Parts, "📍 "+station.CountryCode)
	}

	// Add tags (first 3)
	if station.Tags != "" {
		tags := strings.Split(station.Tags, ",")
		// Limit to first 3 tags
		displayTags := tags
		if len(tags) > 3 {
			displayTags = tags[:3]
		}
		// Trim whitespace from each tag
		for i := range displayTags {
			displayTags[i] = strings.TrimSpace(displayTags[i])
		}
		if len(displayTags) > 0 {
			line2Parts = append(line2Parts, strings.Join(displayTags, ", "))
		}
	}

	// Add bookmark status
	if m.storage != nil && m.storage.IsBookmarked(station.StationUuid) {
		line2Parts = append(line2Parts, "⭐ Bookmarked")
	}

	line2 := strings.Join(line2Parts, " • ")

	// Build the box content
	boxContent := m.theme.PrimaryText.Render(line1)
	if line2 != "" {
		boxContent += "\n" + m.theme.SecondaryText.Render(line2)
	}

	// Create the box with rounded border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.SecondaryColor)).
		Padding(0, 1).
		Width(m.width - 4)

	return boxStyle.Render(boxContent)
}

// buildStatusBar returns the styled status bar string.
// Priority: success message > error message > now playing > default.
func (m StationsModel) buildStatusBar() string {
	if m.successMsg != "" {
		return m.theme.SuccessText.Render(m.successMsg)
	} else if m.err != "" {
		return m.theme.ErrorText.Render(m.err)
	} else if m.currentStation.StationUuid != uuid.Nil && m.playbackManager.IsPlaying() {
		return m.renderNowPlayingBox()
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
	m.updateTableDimensions()
}

// updateTableDimensions recalculates and sets the table dimensions.
// Call this after recreating the table with newStationsTableModel().
func (m *StationsModel) updateTableDimensions() {
	m.stationsTable.SetWidth(m.width)

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
	tableHeight := m.height - 3 - statusHeight
	if tableHeight < 1 {
		tableHeight = 1
	}
	m.stationsTable.SetHeight(tableHeight)
}

// IsModalShowing returns true if a modal dialog is currently displayed.
func (m StationsModel) IsModalShowing() bool {
	return m.showHiddenModal
}

// rebuildTablePreservingCursor rebuilds the stations table and restores the cursor position.
// If cursorOverride is >= 0, it uses that value; otherwise it preserves the current cursor.
// The cursor is bounds-checked to ensure it doesn't exceed the number of stations.
func (m *StationsModel) rebuildTablePreservingCursor(cursorOverride int) {
	cursor := m.stationsTable.Cursor()
	if cursorOverride >= 0 {
		cursor = cursorOverride
	}
	m.stationsTable = newStationsTableModel(m.theme, m.stations, m.storage, m.currentStation)
	m.updateTableDimensions()
	m.setCursorSafely(cursor)
}

// setCursorSafely sets the table cursor with bounds checking.
// If cursor is out of bounds, it clamps to the last valid index.
// If there are no stations, the cursor is set to 0.
func (m *StationsModel) setCursorSafely(cursor int) {
	if len(m.stations) == 0 {
		m.stationsTable.SetCursor(0)
		return
	}
	if cursor >= len(m.stations) {
		m.stationsTable.SetCursor(len(m.stations) - 1)
	} else if cursor < 0 {
		m.stationsTable.SetCursor(0)
	} else {
		m.stationsTable.SetCursor(cursor)
	}
}
