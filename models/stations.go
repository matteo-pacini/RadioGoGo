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
	"time"

	"github.com/google/uuid"
	"github.com/zi0p4tch0/radiogogo/api"
	"github.com/zi0p4tch0/radiogogo/common"
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
	theme Theme

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

	// Hidden modal state
	showHiddenModal   bool
	hiddenStations    []common.Station
	hiddenModalCursor int

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
) StationsModel {

	return StationsModel{
		theme:           theme,
		stations:        stations,
		stationsTable:   newStationsTableModel(theme, stations, storage),
		volume:          playbackManager.VolumeDefault(),
		viewMode:        viewMode,
		storage:         storage,
		browser:         browser,
		playbackManager: playbackManager,
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
			{Title: "Name", Width: 30},
			{Title: "Country", Width: 10},
			{Title: "Language(s)", Width: 15},
			{Title: "Codec(s)", Width: 15},
			{Title: "Votes", Width: 10},
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
		updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false),
		func() tea.Msg {
			return stationCursorMovedMsg{
				offset:        m.stationsTable.Cursor(),
				totalStations: len(m.stations),
			}
		},
	)
}

// Update handles incoming messages and updates the model state accordingly.
func (m StationsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// Playback state messages
	case playbackStartedMsg:
		m.currentStation = msg.station
		m.volumeChangePending = false
		m.currentStationSpinner = spinner.New()
		m.currentStationSpinner.Spinner = spinner.Dot
		m.currentStationSpinner.Style = m.theme.PrimaryText
		return m, tea.Batch(
			m.currentStationSpinner.Tick,
			notifyRadioBrowserCmd(m.browser, m.currentStation),
			updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
			func() tea.Msg { return playbackStatusMsg{status: PlaybackPlaying} },
		)
	case playbackStoppedMsg:
		m.currentStation = common.Station{}
		m.currentStationSpinner = spinner.Model{}
		return m, tea.Batch(
			updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false),
			func() tea.Msg { return playbackStatusMsg{status: PlaybackIdle} },
			func() tea.Msg { return recordingStatusMsg{isRecording: false} },
		)

	// Error handling
	case nonFatalError:
		var cmds []tea.Cmd
		if msg.stopPlayback {
			cmds = append(cmds, stopStationCmd(m.playbackManager))
		}
		m.err = msg.err.Error()
		cmds = append(cmds, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		}))

		return m, tea.Sequence(cmds...)
	case clearNonFatalError:
		m.err = ""
		return m, nil

	// Volume change with debouncing
	case volumeDebounceExpiredMsg:
		if msg.changeID == m.pendingVolumeChangeID && m.volumeChangePending {
			m.volumeChangePending = false
			station := m.playbackManager.CurrentStation()
			if station.StationUuid != uuid.Nil {
				return m, restartPlaybackWithVolumeCmd(m.playbackManager, station, m.volume)
			}
		}
		return m, nil
	case volumeRestartCompleteMsg:
		m.currentStation = msg.station
		return m, func() tea.Msg { return playbackStatusMsg{status: PlaybackPlaying} }
	case volumeRestartFailedMsg:
		m.err = fmt.Sprintf("Volume change failed: %v", msg.err)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})

	// Recording state messages
	case recordingStartedMsg:
		return m, tea.Batch(
			updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), true),
			func() tea.Msg { return recordingStatusMsg{isRecording: true} },
		)
	case recordingStoppedMsg:
		return m, tea.Batch(
			updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), false),
			func() tea.Msg { return recordingStatusMsg{isRecording: false} },
		)
	case recordingErrorMsg:
		m.err = fmt.Sprintf("Recording error: %v", msg.err)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})

	// Bookmark and hidden station messages
	case bookmarkToggledMsg:
		// Refresh table to update star prefix
		m.stationsTable = newStationsTableModel(m.theme, m.stations, m.storage)
		return m, nil
	case stationHiddenMsg:
		// Remove hidden station from current view
		newStations := make([]common.Station, 0, len(m.stations)-1)
		for _, s := range m.stations {
			if s.StationUuid != msg.station.StationUuid {
				newStations = append(newStations, s)
			}
		}
		m.stations = newStations
		m.stationsTable = newStationsTableModel(m.theme, m.stations, m.storage)
		// Adjust cursor if needed
		if msg.cursor >= len(m.stations) && len(m.stations) > 0 {
			m.stationsTable.SetCursor(len(m.stations) - 1)
		}
		return m, func() tea.Msg {
			return stationCursorMovedMsg{
				offset:        m.stationsTable.Cursor(),
				totalStations: len(m.stations),
			}
		}
	case bookmarksFetchedMsg:
		// Switch to bookmarks view
		m.viewMode = viewModeBookmarks
		m.stations = msg.stations
		m.stationsTable = newStationsTableModel(m.theme, m.stations, m.storage)
		return m, tea.Batch(
			updateCommandsCmd(m.viewMode, m.playbackManager.IsPlaying(), m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
			func() tea.Msg {
				return stationCursorMovedMsg{
					offset:        m.stationsTable.Cursor(),
					totalStations: len(m.stations),
				}
			},
		)
	case bookmarksFetchFailedMsg:
		m.err = fmt.Sprintf("Failed to load bookmarks: %v", msg.err)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})
	case hiddenFetchedMsg:
		m.hiddenStations = msg.stations
		m.hiddenModalCursor = 0
		m.showHiddenModal = true
		return m, nil
	case hiddenFetchFailedMsg:
		m.err = fmt.Sprintf("Failed to load hidden stations: %v", msg.err)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})
	case stationUnhiddenMsg:
		// Remove from modal list
		newHidden := make([]common.Station, 0, len(m.hiddenStations)-1)
		for _, s := range m.hiddenStations {
			if s.StationUuid != msg.station.StationUuid {
				newHidden = append(newHidden, s)
			}
		}
		m.hiddenStations = newHidden
		if m.hiddenModalCursor >= len(m.hiddenStations) && len(m.hiddenStations) > 0 {
			m.hiddenModalCursor = len(m.hiddenStations) - 1
		}
		if len(m.hiddenStations) == 0 {
			m.showHiddenModal = false
		}
		return m, nil

	// Key event handling
	case tea.KeyMsg:
		// Handle modal input first (captures all keys when modal is open)
		if handled, cmd := m.handleHiddenModalInput(msg); handled {
			return m, cmd
		}

		switch msg.String() {
		case "up", "down", "j", "k":
			cmds = append(cmds, func() tea.Msg {
				return stationCursorMovedMsg{
					offset:        m.stationsTable.Cursor(),
					totalStations: len(m.stations),
				}
			})
		case "ctrl+k":
			return m, func() tea.Msg {
				err := m.playbackManager.StopStation()
				if err != nil {
					return nonFatalError{stopPlayback: false, err: err}
				}
				return playbackStoppedMsg{}
			}
		case "q":
			return m, tea.Sequence(stopStationCmd(m.playbackManager), quitCmd)
		case "s":
			return m, func() tea.Msg {
				return switchToSearchModelMsg{}
			}
		case "9":
			return m, m.handleVolumeDecrease()
		case "0":
			return m, m.handleVolumeIncrease()
		case "r":
			return m, m.handleRecordingToggle()
		case "b":
			// Toggle bookmark on selected station
			if len(m.stations) == 0 {
				return m, nil
			}
			station := m.stations[m.stationsTable.Cursor()]
			return m, toggleBookmarkCmd(m.storage, station)
		case "h":
			// Hide station (only in search results mode)
			if m.viewMode != viewModeSearchResults {
				return m, nil
			}
			if len(m.stations) == 0 {
				return m, nil
			}
			station := m.stations[m.stationsTable.Cursor()]
			return m, hideStationCmd(m.storage, station, m.stationsTable.Cursor())
		case "B":
			// Toggle view mode (search results <-> bookmarks)
			if m.viewMode == viewModeSearchResults {
				return m, fetchBookmarksCmd(m.browser, m.storage)
			} else {
				return m, func() tea.Msg { return switchToSearchModelMsg{} }
			}
		case "H":
			// Open hidden stations modal
			return m, fetchHiddenStationsCmd(m.browser, m.storage)
		case "enter":
			if len(m.stations) == 0 {
				return m, nil
			}
			station := m.stations[m.stationsTable.Cursor()]
			return m, playStationCmd(m.playbackManager, station, m.volume)
		}
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

// handleVolumeDecrease handles the volume decrease key press with debouncing.
func (m *StationsModel) handleVolumeDecrease() tea.Cmd {
	if m.volume > m.playbackManager.VolumeMin() {
		m.volume -= 10
		if m.playbackManager.IsPlaying() {
			changeID := time.Now().UnixNano()
			m.pendingVolumeChangeID = changeID
			m.volumeChangePending = true
			return tea.Batch(
				updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
				startVolumeDebounceCmd(changeID),
				func() tea.Msg { return playbackStatusMsg{status: PlaybackRestarting} },
			)
		}
		return updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false)
	}
	return nil
}

// handleVolumeIncrease handles the volume increase key press with debouncing.
func (m *StationsModel) handleVolumeIncrease() tea.Cmd {
	if m.volume < m.playbackManager.VolumeMax() {
		m.volume += 10
		if m.playbackManager.IsPlaying() {
			changeID := time.Now().UnixNano()
			m.pendingVolumeChangeID = changeID
			m.volumeChangePending = true
			return tea.Batch(
				updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
				startVolumeDebounceCmd(changeID),
				func() tea.Msg { return playbackStatusMsg{status: PlaybackRestarting} },
			)
		}
		return updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false)
	}
	return nil
}

// handleRecordingToggle handles the recording toggle key press.
func (m *StationsModel) handleRecordingToggle() tea.Cmd {
	// Only allow recording if something is playing
	if !m.playbackManager.IsPlaying() {
		return nil
	}

	// Toggle recording
	if m.playbackManager.IsRecording() {
		return stopRecordingCmd(m.playbackManager)
	}

	// Check if ffmpeg is available
	if !m.playbackManager.IsRecordingAvailable() {
		m.err = m.playbackManager.RecordingNotAvailableErrorString()
		return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})
	}

	// Generate filename and start recording
	station := m.playbackManager.CurrentStation()
	filename := playback.GenerateRecordingFilename(station.Name, station.Codec)
	return startRecordingCmd(m.playbackManager, filename)
}

// View renders the stations view.
func (m StationsModel) View() string {

	var v string
	if len(m.stations) == 0 {
		var emptyMsg string
		if m.viewMode == viewModeBookmarks {
			emptyMsg = "No bookmarks yet! Press 'B' to go back and bookmark some stations."
		} else {
			emptyMsg = "No stations found, try another search!"
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
			m.theme.SecondaryText.Bold(true).Render("Now playing: "+m.currentStation.Name) +
			" " + m.currentStationSpinner.View()
	} else {
		statusBar = m.theme.TertiaryText.Render("Select a station and press enter to play")
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
