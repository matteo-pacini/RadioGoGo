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
	"time"

	"github.com/google/uuid"
	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/i18n"
	"github.com/zi0p4tch0/radiogogo/playback"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// handlePlaybackMessages handles playback-related messages (start, stop).
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m StationsModel) handlePlaybackMessages(msg tea.Msg) (bool, StationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case playbackStartedMsg:
		m.currentStation = msg.station
		m.volumeChangePending = false
		m.currentStationSpinner = spinner.New()
		m.currentStationSpinner.Spinner = spinner.Dot
		m.currentStationSpinner.Style = m.theme.PrimaryText
		return true, m, tea.Batch(
			m.currentStationSpinner.Tick,
			notifyRadioBrowserCmd(m.browser, m.currentStation),
			updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording(), m.keybindings),
			func() tea.Msg { return playbackStatusMsg{status: PlaybackPlaying} },
			func() tea.Msg { return recordingStatusMsg{isRecording: false} },
		)
	case playbackStoppedMsg:
		m.currentStation = common.Station{}
		m.currentStationSpinner = spinner.Model{}
		return true, m, tea.Batch(
			updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false, m.keybindings),
			func() tea.Msg { return playbackStatusMsg{status: PlaybackIdle} },
			func() tea.Msg { return recordingStatusMsg{isRecording: false} },
		)
	}
	return false, m, nil
}

// handleErrorMessages handles error-related messages.
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m StationsModel) handleErrorMessages(msg tea.Msg) (bool, StationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case nonFatalError:
		var cmds []tea.Cmd
		if msg.stopPlayback {
			cmds = append(cmds, stopStationCmd(m.playbackManager))
		}
		m.err = msg.err.Error()
		cmds = append(cmds, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		}))
		return true, m, tea.Sequence(cmds...)
	case clearNonFatalError:
		m.err = ""
		return true, m, nil
	}
	return false, m, nil
}

// handleVolumeMessages handles volume change-related messages.
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m StationsModel) handleVolumeMessages(msg tea.Msg) (bool, StationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case volumeDebounceExpiredMsg:
		if msg.changeID == m.pendingVolumeChangeID && m.volumeChangePending {
			m.volumeChangePending = false
			station := m.playbackManager.CurrentStation()
			if station.StationUuid != uuid.Nil {
				return true, m, restartPlaybackWithVolumeCmd(m.playbackManager, station, m.volume)
			}
		}
		return true, m, nil
	case volumeRestartCompleteMsg:
		m.currentStation = msg.station
		return true, m, func() tea.Msg { return playbackStatusMsg{status: PlaybackPlaying} }
	case volumeRestartFailedMsg:
		m.err = i18n.Tf("error_volume_change", map[string]interface{}{"Error": msg.err})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})
	}
	return false, m, nil
}

// handleRecordingMessages handles recording-related messages.
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m StationsModel) handleRecordingMessages(msg tea.Msg) (bool, StationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case recordingStartedMsg:
		return true, m, tea.Batch(
			updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), true, m.keybindings),
			func() tea.Msg { return recordingStatusMsg{isRecording: true} },
		)
	case recordingStoppedMsg:
		return true, m, tea.Batch(
			updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), false, m.keybindings),
			func() tea.Msg { return recordingStatusMsg{isRecording: false} },
		)
	case recordingErrorMsg:
		m.err = i18n.Tf("error_recording", map[string]interface{}{"Error": msg.err})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})
	}
	return false, m, nil
}

// handleBookmarkMessages handles bookmark-related messages.
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m StationsModel) handleBookmarkMessages(msg tea.Msg) (bool, StationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case bookmarkToggledMsg:
		cursor := m.stationsTable.Cursor()
		if m.viewMode == viewModeBookmarks {
			m.savedCursor = cursor
			return true, m, fetchBookmarksCmd(m.browser, m.storage)
		}
		m.stationsTable = newStationsTableModel(m.theme, m.stations, m.storage)
		m.updateTableDimensions()
		m.stationsTable.SetCursor(cursor)
		return true, m, nil

	case bookmarksFetchedMsg:
		cursorToRestore := 0
		if m.viewMode == viewModeSearchResults {
			m.savedStations = m.stations
			m.savedCursor = m.stationsTable.Cursor()
		} else {
			cursorToRestore = m.savedCursor
		}
		m.viewMode = viewModeBookmarks
		m.stations = msg.stations
		m.stationsTable = newStationsTableModel(m.theme, m.stations, m.storage)
		m.updateTableDimensions()
		if cursorToRestore > 0 && cursorToRestore < len(m.stations) {
			m.stationsTable.SetCursor(cursorToRestore)
		} else if cursorToRestore >= len(m.stations) && len(m.stations) > 0 {
			m.stationsTable.SetCursor(len(m.stations) - 1)
		}
		return true, m, tea.Batch(
			updateCommandsCmd(m.viewMode, m.playbackManager.IsPlaying(), m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording(), m.keybindings),
			func() tea.Msg {
				return stationCursorMovedMsg{
					offset:        m.stationsTable.Cursor(),
					totalStations: len(m.stations),
				}
			},
		)

	case bookmarksFetchFailedMsg:
		m.err = i18n.Tf("error_load_bookmarks", map[string]interface{}{"Error": msg.err})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})

	case bookmarkToggleFailedMsg:
		m.err = i18n.Tf("error_bookmark_toggle", map[string]interface{}{"Error": msg.err})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})
	}
	return false, m, nil
}

// handleHiddenStationMessages handles hidden station-related messages.
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m StationsModel) handleHiddenStationMessages(msg tea.Msg) (bool, StationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case stationHiddenMsg:
		newStations := make([]common.Station, 0, len(m.stations)-1)
		for _, s := range m.stations {
			if s.StationUuid != msg.station.StationUuid {
				newStations = append(newStations, s)
			}
		}
		m.stations = newStations
		m.stationsTable = newStationsTableModel(m.theme, m.stations, m.storage)
		m.updateTableDimensions()
		if msg.cursor >= len(m.stations) && len(m.stations) > 0 {
			m.stationsTable.SetCursor(len(m.stations) - 1)
		} else {
			m.stationsTable.SetCursor(msg.cursor)
		}
		return true, m, func() tea.Msg {
			return stationCursorMovedMsg{
				offset:        m.stationsTable.Cursor(),
				totalStations: len(m.stations),
			}
		}

	case hiddenFetchedMsg:
		m.hiddenStations = msg.stations
		m.hiddenModalCursor = 0
		m.showHiddenModal = true
		return true, m, nil

	case hiddenFetchFailedMsg:
		m.err = i18n.Tf("error_load_hidden", map[string]interface{}{"Error": msg.err})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})

	case hideStationFailedMsg:
		m.err = i18n.Tf("error_hide_station", map[string]interface{}{"Error": msg.err})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})

	case unhideStationFailedMsg:
		m.err = i18n.Tf("error_unhide_station", map[string]interface{}{"Error": msg.err})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})

	case stationUnhiddenMsg:
		m.needsRefetch = true
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
			if m.needsRefetch {
				m.needsRefetch = false
				return true, m, refetchStationsCmd(m.browser, m.lastQuery, m.lastQueryText)
			}
		}
		return true, m, nil

	case stationsRefetchedMsg:
		filtered := make([]common.Station, 0, len(msg.stations))
		for _, s := range msg.stations {
			if !m.storage.IsHidden(s.StationUuid) {
				filtered = append(filtered, s)
			}
		}
		m.stations = filtered
		m.stationsTable = newStationsTableModel(m.theme, m.stations, m.storage)
		m.updateTableDimensions()
		// Restore cursor position (used by vote success and unhide)
		if m.savedCursor > 0 && m.savedCursor < len(m.stations) {
			m.stationsTable.SetCursor(m.savedCursor)
		} else if m.savedCursor >= len(m.stations) && len(m.stations) > 0 {
			m.stationsTable.SetCursor(len(m.stations) - 1)
		}
		m.savedCursor = 0
		return true, m, func() tea.Msg {
			return stationCursorMovedMsg{
				offset:        m.stationsTable.Cursor(),
				totalStations: len(m.stations),
			}
		}

	case stationsRefetchFailedMsg:
		m.err = i18n.Tf("error_refresh_stations", map[string]interface{}{"Error": msg.err})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})
	}
	return false, m, nil
}

// clearSuccessMsg clears the success message from the status bar.
type clearSuccessMsg struct{}

// handleVoteMessages handles voting-related messages.
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m StationsModel) handleVoteMessages(msg tea.Msg) (bool, StationsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case voteSucceededMsg:
		m.successMsg = i18n.T("vote_success")
		m.savedCursor = msg.cursor
		return true, m, tea.Batch(
			refetchStationsCmd(m.browser, m.lastQuery, m.lastQueryText),
			tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return clearSuccessMsg{}
			}),
		)

	case voteFailedMsg:
		m.err = i18n.Tf("error_vote", map[string]interface{}{"Error": msg.err})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})

	case voteCooldownMsg:
		m.err = i18n.Tfn("error_vote_cooldown", msg.remainingMinutes, map[string]interface{}{"Minutes": msg.remainingMinutes})
		return true, m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})

	case clearSuccessMsg:
		m.successMsg = ""
		return true, m, nil
	}
	return false, m, nil
}

// handleKeyMessage handles keyboard input.
// Returns (handled, model, cmd) where handled indicates if the message was processed.
func (m StationsModel) handleKeyMessage(msg tea.KeyMsg) (bool, StationsModel, tea.Cmd) {
	// Handle modal input first (captures all keys when modal is open)
	if handled, cmd := m.handleHiddenModalInput(msg); handled {
		return true, m, cmd
	}

	key := msg.String()

	// Navigation keys - just track cursor movement
	if key == "up" || key == "down" || key == m.keybindings.NavigateDown || key == m.keybindings.NavigateUp {
		return false, m, func() tea.Msg {
			return stationCursorMovedMsg{
				offset:        m.stationsTable.Cursor(),
				totalStations: len(m.stations),
			}
		}
	}

	switch {
	case key == m.keybindings.StopPlayback:
		return true, m, func() tea.Msg {
			err := m.playbackManager.StopStation()
			if err != nil {
				return nonFatalError{stopPlayback: false, err: err}
			}
			return playbackStoppedMsg{}
		}

	case key == m.keybindings.Quit:
		return true, m, tea.Sequence(stopStationCmd(m.playbackManager), quitCmd)

	case key == m.keybindings.Search:
		return true, m, func() tea.Msg { return switchToSearchModelMsg{} }

	case key == m.keybindings.VolumeDown:
		return true, m, m.handleVolumeChange(-1)

	case key == m.keybindings.VolumeUp:
		return true, m, m.handleVolumeChange(1)

	case key == m.keybindings.Record:
		return true, m, m.handleRecordingToggle()

	case key == m.keybindings.BookmarkToggle:
		if len(m.stations) == 0 {
			return true, m, nil
		}
		station := m.stations[m.stationsTable.Cursor()]
		return true, m, toggleBookmarkCmd(m.storage, station)

	case key == m.keybindings.HideStation:
		if m.viewMode != viewModeSearchResults || len(m.stations) == 0 {
			return true, m, nil
		}
		station := m.stations[m.stationsTable.Cursor()]
		// If hiding the currently playing station, stop playback (and recording if active)
		if m.currentStation.StationUuid == station.StationUuid {
			if m.playbackManager.IsRecording() {
				return true, m, tea.Sequence(
					stopRecordingCmd(m.playbackManager),
					stopStationCmd(m.playbackManager),
					hideStationCmd(m.storage, station, m.stationsTable.Cursor()),
				)
			}
			return true, m, tea.Sequence(
				stopStationCmd(m.playbackManager),
				hideStationCmd(m.storage, station, m.stationsTable.Cursor()),
			)
		}
		return true, m, hideStationCmd(m.storage, station, m.stationsTable.Cursor())

	case key == m.keybindings.BookmarksView:
		return m.handleBookmarksViewToggle()

	case key == m.keybindings.ManageHidden:
		if m.viewMode != viewModeSearchResults {
			return true, m, nil
		}
		return true, m, fetchHiddenStationsCmd(m.browser, m.storage)

	case key == m.keybindings.Vote:
		if m.viewMode != viewModeSearchResults || len(m.stations) == 0 {
			return true, m, nil
		}
		station := m.stations[m.stationsTable.Cursor()]
		return true, m, voteStationCmd(m.browser, m.storage, station, m.stationsTable.Cursor())

	case key == "enter":
		if len(m.stations) == 0 {
			return true, m, nil
		}
		station := m.stations[m.stationsTable.Cursor()]
		return true, m, playStationCmd(m.playbackManager, station, m.volume)
	}

	return false, m, nil
}

// handleBookmarksViewToggle handles toggling between search results and bookmarks view.
func (m StationsModel) handleBookmarksViewToggle() (bool, StationsModel, tea.Cmd) {
	if m.viewMode == viewModeSearchResults {
		return true, m, tea.Sequence(
			stopStationCmd(m.playbackManager),
			fetchBookmarksCmd(m.browser, m.storage),
		)
	}

	// Return from bookmarks to previous stations (or search if none saved)
	if len(m.savedStations) > 0 {
		if err := m.playbackManager.StopStation(); err != nil {
			m.err = err.Error()
			// Continue anyway - we're transitioning views
		}
		m.currentStation = common.Station{}
		m.currentStationSpinner = spinner.Model{}

		m.viewMode = viewModeSearchResults
		m.stations = m.savedStations
		m.stationsTable = newStationsTableModel(m.theme, m.stations, m.storage)
		m.updateTableDimensions()
		m.stationsTable.SetCursor(m.savedCursor)
		m.savedStations = nil
		m.savedCursor = 0
		return true, m, tea.Batch(
			updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false, m.keybindings),
			func() tea.Msg { return playbackStatusMsg{status: PlaybackIdle} },
			func() tea.Msg {
				return stationCursorMovedMsg{
					offset:        m.stationsTable.Cursor(),
					totalStations: len(m.stations),
				}
			},
		)
	}
	return true, m, func() tea.Msg { return switchToSearchModelMsg{} }
}

// handleVolumeChange handles volume increase or decrease with debouncing.
// direction should be positive for increase, negative for decrease.
func (m *StationsModel) handleVolumeChange(direction int) tea.Cmd {
	step := 10
	if direction < 0 {
		step = -10
	}

	newVolume := m.volume + step
	if newVolume < m.playbackManager.VolumeMin() || newVolume > m.playbackManager.VolumeMax() {
		return nil
	}

	m.volume = newVolume
	if m.playbackManager.IsPlaying() {
		changeID := time.Now().UnixNano()
		m.pendingVolumeChangeID = changeID
		m.volumeChangePending = true
		return tea.Batch(
			updateCommandsCmd(m.viewMode, true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording(), m.keybindings),
			startVolumeDebounceCmd(changeID),
			func() tea.Msg { return playbackStatusMsg{status: PlaybackRestarting} },
		)
	}
	return updateCommandsCmd(m.viewMode, false, m.volume, m.playbackManager.VolumeIsPercentage(), false, m.keybindings)
}

// handleRecordingToggle handles the recording toggle key press.
func (m *StationsModel) handleRecordingToggle() tea.Cmd {
	if !m.playbackManager.IsPlaying() {
		return nil
	}

	if m.playbackManager.IsRecording() {
		return stopRecordingCmd(m.playbackManager)
	}

	if !m.playbackManager.IsRecordingAvailable() {
		m.err = m.playbackManager.RecordingNotAvailableErrorString()
		return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})
	}

	station := m.playbackManager.CurrentStation()
	filename := playback.GenerateRecordingFilename(station.Name, station.Codec)
	return startRecordingCmd(m.playbackManager, filename)
}
