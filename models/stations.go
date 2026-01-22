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

// Messages

type playbackStartedMsg struct {
	station common.Station
}
type playbackStoppedMsg struct{}

type nonFatalError struct {
	stopPlayback bool
	err          error
}
type clearNonFatalError struct{}

type stationCursorMovedMsg struct {
	offset        int
	totalStations int
}

type volumeDebounceExpiredMsg struct {
	changeID int64
}

type volumeRestartCompleteMsg struct {
	station common.Station
}

type volumeRestartFailedMsg struct {
	err error
}

type recordingStartedMsg struct {
	filePath string
}

type recordingStoppedMsg struct {
	filePath string
}

type recordingErrorMsg struct {
	err error
}

// Bookmark/Hidden messages
type bookmarkToggledMsg struct {
	station common.Station
}
type stationHiddenMsg struct {
	station common.Station
	cursor  int
}
type bookmarksFetchedMsg struct {
	stations []common.Station
}
type bookmarksFetchFailedMsg struct {
	err error
}
type hiddenFetchedMsg struct {
	stations []common.Station
}
type hiddenFetchFailedMsg struct {
	err error
}
type stationUnhiddenMsg struct {
	station common.Station
}

// Commands

func playStationCmd(
	playbackManager playback.PlaybackManagerService,
	station common.Station,
	volume int,
) tea.Cmd {
	return func() tea.Msg {
		err := playbackManager.PlayStation(station, volume)
		if err != nil {
			return nonFatalError{stopPlayback: false, err: err}
		}
		return playbackStartedMsg{station: station}
	}
}

func stopStationCmd(playbackManager playback.PlaybackManagerService) tea.Cmd {
	return func() tea.Msg {
		err := playbackManager.StopStation()
		if err != nil {
			return nonFatalError{stopPlayback: false, err: err}
		}
		return playbackStoppedMsg{}
	}
}

func notifyRadioBrowserCmd(browser api.RadioBrowserService, station common.Station) tea.Cmd {
	return func() tea.Msg {
		_, err := browser.ClickStation(station)
		if err != nil {
			return nonFatalError{stopPlayback: false, err: err}
		}
		return nil
	}
}

const volumeDebounceDelay = 300 * time.Millisecond

func startVolumeDebounceCmd(changeID int64) tea.Cmd {
	return tea.Tick(volumeDebounceDelay, func(t time.Time) tea.Msg {
		return volumeDebounceExpiredMsg{changeID: changeID}
	})
}

func restartPlaybackWithVolumeCmd(
	pm playback.PlaybackManagerService,
	station common.Station,
	volume int,
) tea.Cmd {
	return func() tea.Msg {
		if err := pm.StopStation(); err != nil {
			return volumeRestartFailedMsg{err: err}
		}
		if err := pm.PlayStation(station, volume); err != nil {
			return volumeRestartFailedMsg{err: err}
		}
		return volumeRestartCompleteMsg{station: station}
	}
}

func startRecordingCmd(pm playback.PlaybackManagerService, outputPath string) tea.Cmd {
	return func() tea.Msg {
		err := pm.StartRecording(outputPath)
		if err != nil {
			return recordingErrorMsg{err: err}
		}
		return recordingStartedMsg{filePath: outputPath}
	}
}

func stopRecordingCmd(pm playback.PlaybackManagerService) tea.Cmd {
	return func() tea.Msg {
		filePath, err := pm.StopRecording()
		if err != nil {
			return recordingErrorMsg{err: err}
		}
		return recordingStoppedMsg{filePath: filePath}
	}
}

func toggleBookmarkCmd(storage storage.StationStorageService, station common.Station) tea.Cmd {
	return func() tea.Msg {
		if storage.IsBookmarked(station.StationUuid) {
			storage.RemoveBookmark(station.StationUuid)
		} else {
			storage.AddBookmark(station.StationUuid)
		}
		return bookmarkToggledMsg{station: station}
	}
}

func hideStationCmd(storage storage.StationStorageService, station common.Station, cursor int) tea.Cmd {
	return func() tea.Msg {
		storage.AddHidden(station.StationUuid)
		return stationHiddenMsg{station: station, cursor: cursor}
	}
}

func unhideStationCmd(storage storage.StationStorageService, station common.Station) tea.Cmd {
	return func() tea.Msg {
		storage.RemoveHidden(station.StationUuid)
		return stationUnhiddenMsg{station: station}
	}
}

func fetchBookmarksCmd(browser api.RadioBrowserService, storage storage.StationStorageService) tea.Cmd {
	return func() tea.Msg {
		uuids, err := storage.GetBookmarks()
		if err != nil {
			return bookmarksFetchFailedMsg{err: err}
		}
		if len(uuids) == 0 {
			return bookmarksFetchedMsg{stations: []common.Station{}}
		}
		stations, err := browser.GetStationsByUUIDs(uuids)
		if err != nil {
			return bookmarksFetchFailedMsg{err: err}
		}
		return bookmarksFetchedMsg{stations: stations}
	}
}

func fetchHiddenStationsCmd(browser api.RadioBrowserService, storage storage.StationStorageService) tea.Cmd {
	return func() tea.Msg {
		uuids, err := storage.GetHidden()
		if err != nil {
			return hiddenFetchFailedMsg{err: err}
		}
		if len(uuids) == 0 {
			return hiddenFetchedMsg{stations: []common.Station{}}
		}
		stations, err := browser.GetStationsByUUIDs(uuids)
		if err != nil {
			return hiddenFetchFailedMsg{err: err}
		}
		return hiddenFetchedMsg{stations: stations}
	}
}

func updateCommandsCmd(viewMode stationsViewMode, isPlaying bool, volume int, volumeIsPercentage bool, isRecording bool) tea.Cmd {
	return func() tea.Msg {

		// Row 1: Navigation and playback
		var commands []string
		if viewMode == viewModeSearchResults {
			commands = []string{"q: quit", "s: search", "enter: play", "↑/↓: move"}
		} else {
			commands = []string{"q: quit", "B: back", "enter: play", "↑/↓: move"}
		}

		var volumeDisplay string
		if volume == 0 {
			volumeDisplay = "mute"
		} else {
			volumeDisplay = fmt.Sprintf("vol: %d", volume)
			if volumeIsPercentage {
				volumeDisplay += "%"
			}
		}

		if isPlaying {
			if isRecording {
				commands = append(commands, "r: stop rec", "ctrl+k: stop", "9/0: vol", volumeDisplay)
			} else {
				commands = append(commands, "r: record", "ctrl+k: stop", "9/0: vol", volumeDisplay)
			}
		} else {
			commands = append(commands, "9/0: vol", volumeDisplay)
		}

		// Row 2: Bookmark/hide commands
		var secondaryCommands []string
		if viewMode == viewModeSearchResults {
			secondaryCommands = []string{"b: bookmark", "B: bookmarks", "h: hide", "H: manage hidden"}
		} else {
			secondaryCommands = []string{"b: bookmark", "B: back", "H: manage hidden"}
		}

		return bottomBarUpdateMsg{
			commands:          commands,
			secondaryCommands: secondaryCommands,
		}
	}
}

// Model

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

func (m StationsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmds []tea.Cmd

	switch msg := msg.(type) {
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
	case tea.KeyMsg:
		// Modal input handling - capture all keys when modal is showing
		if m.showHiddenModal {
			switch msg.String() {
			case "up", "k":
				if m.hiddenModalCursor > 0 {
					m.hiddenModalCursor--
				}
				return m, nil
			case "down", "j":
				if m.hiddenModalCursor < len(m.hiddenStations)-1 {
					m.hiddenModalCursor++
				}
				return m, nil
			case "enter":
				if len(m.hiddenStations) > 0 {
					station := m.hiddenStations[m.hiddenModalCursor]
					return m, unhideStationCmd(m.storage, station)
				}
				return m, nil
			case "esc", "H", "q":
				m.showHiddenModal = false
				return m, updateCommandsCmd(m.viewMode, m.playbackManager.IsPlaying(), m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording())
			}
			return m, nil
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
			if m.volume > m.playbackManager.VolumeMin() {
				m.volume -= 10
				if m.playbackManager.IsPlaying() {
					changeID := time.Now().UnixNano()
					m.pendingVolumeChangeID = changeID
					m.volumeChangePending = true
					return m, tea.Batch(
						updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
						startVolumeDebounceCmd(changeID),
						func() tea.Msg { return playbackStatusMsg{status: PlaybackRestarting} },
					)
				}
				return m, updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false)
			}
			return m, nil
		case "0":
			if m.volume < m.playbackManager.VolumeMax() {
				m.volume += 10
				if m.playbackManager.IsPlaying() {
					changeID := time.Now().UnixNano()
					m.pendingVolumeChangeID = changeID
					m.volumeChangePending = true
					return m, tea.Batch(
						updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
						startVolumeDebounceCmd(changeID),
						func() tea.Msg { return playbackStatusMsg{status: PlaybackRestarting} },
					)
				}
				return m, updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false)
			}
			return m, nil
		case "r":
			// Only allow recording if something is playing
			if !m.playbackManager.IsPlaying() {
				return m, nil
			}

			// Toggle recording
			if m.playbackManager.IsRecording() {
				return m, stopRecordingCmd(m.playbackManager)
			}

			// Check if ffmpeg is available
			if !m.playbackManager.IsRecordingAvailable() {
				m.err = m.playbackManager.RecordingNotAvailableErrorString()
				return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
					return clearNonFatalError{}
				})
			}

			// Generate filename and start recording
			station := m.playbackManager.CurrentStation()
			filename := playback.GenerateRecordingFilename(station.Name, station.Codec)
			return m, startRecordingCmd(m.playbackManager, filename)
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

	if m.playbackManager.IsPlaying() {
		newSpinner, cmd := m.currentStationSpinner.Update(msg)
		m.currentStationSpinner = newSpinner
		cmds = append(cmds, cmd)
	}

	newStationsTable, cmd := m.stationsTable.Update(msg)
	m.stationsTable = newStationsTable

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

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

func (m StationsModel) renderWithModal(baseView string) string {
	modalContent := m.theme.SecondaryText.Bold(true).Render("Hidden Stations") + "\n\n"

	if len(m.hiddenStations) == 0 {
		modalContent += m.theme.TertiaryText.Render("No hidden stations")
	} else {
		for i, station := range m.hiddenStations {
			cursor := "  "
			if i == m.hiddenModalCursor {
				cursor = "> "
			}
			name := station.Name
			if len(name) > 40 {
				name = name[:37] + "..."
			}
			modalContent += cursor + name + "\n"
		}
	}
	modalContent += "\n" + m.theme.TertiaryText.Render("Enter: unhide | Esc/H: close")

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.SecondaryColor)).
		Padding(1, 2).
		Width(50)

	modal := modalStyle.Render(modalContent)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func (m *StationsModel) SetWidthAndHeight(width int, height int) {
	m.width = width
	m.height = height
	m.stationsTable.SetWidth(width)
	// Table height calculation:
	// - View adds "\n" before and "\n" after the table (2 lines)
	// - We want minimal filler for visual separation
	// So: tableHeight = height - 2
	m.stationsTable.SetHeight(height - 2)
}
