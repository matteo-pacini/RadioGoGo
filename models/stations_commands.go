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
	"errors"
	"time"

	"github.com/zi0p4tch0/radiogogo/api"
	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/i18n"
	"github.com/zi0p4tch0/radiogogo/playback"
	"github.com/zi0p4tch0/radiogogo/storage"

	tea "github.com/charmbracelet/bubbletea"
)

// Playback messages

type playbackStartedMsg struct {
	station common.Station
}
type playbackStoppedMsg struct{}

// Error messages

type nonFatalError struct {
	stopPlayback bool
	err          error
}
type clearNonFatalError struct{}

// Cursor messages

type stationCursorMovedMsg struct {
	offset        int
	totalStations int
}

// Volume messages

type volumeDebounceExpiredMsg struct {
	changeID int64
}

type volumeRestartCompleteMsg struct {
	station common.Station
}

type volumeRestartFailedMsg struct {
	err error
}

// Recording messages

type recordingStartedMsg struct {
	filePath string
}

type recordingStoppedMsg struct {
	filePath string
}

type recordingErrorMsg struct {
	err error
}

// Bookmark and hidden station messages

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
type bookmarkToggleFailedMsg struct {
	err error
}
type hideStationFailedMsg struct {
	err error
}
type unhideStationFailedMsg struct {
	err error
}
type stationsRefetchedMsg struct {
	stations []common.Station
}
type stationsRefetchFailedMsg struct {
	err error
}

// Vote messages

type voteSucceededMsg struct {
	station common.Station
	cursor  int
}
type voteFailedMsg struct {
	err error
}
type voteCooldownMsg struct {
	remainingMinutes int
}

// Playback commands

// playStationCmd starts playback of a station at the given volume.
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

// stopStationCmd stops the currently playing station.
func stopStationCmd(playbackManager playback.PlaybackManagerService) tea.Cmd {
	return func() tea.Msg {
		err := playbackManager.StopStation()
		if err != nil {
			return nonFatalError{stopPlayback: false, err: err}
		}
		return playbackStoppedMsg{}
	}
}

// notifyRadioBrowserCmd notifies the RadioBrowser API that a station was played (click count).
func notifyRadioBrowserCmd(browser api.RadioBrowserService, station common.Station) tea.Cmd {
	return func() tea.Msg {
		_, err := browser.ClickStation(station)
		if err != nil {
			return nonFatalError{stopPlayback: false, err: err}
		}
		return nil
	}
}

// Volume commands

const volumeDebounceDelay = 300 * time.Millisecond

// startVolumeDebounceCmd starts a timer for volume change debouncing.
// When the timer expires, it sends a volumeDebounceExpiredMsg with the changeID.
func startVolumeDebounceCmd(changeID int64) tea.Cmd {
	return tea.Tick(volumeDebounceDelay, func(t time.Time) tea.Msg {
		return volumeDebounceExpiredMsg{changeID: changeID}
	})
}

// restartPlaybackWithVolumeCmd stops and restarts playback with a new volume level.
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

// Recording commands

// startRecordingCmd starts recording the current stream to the given output path.
func startRecordingCmd(pm playback.PlaybackManagerService, outputPath string) tea.Cmd {
	return func() tea.Msg {
		err := pm.StartRecording(outputPath)
		if err != nil {
			return recordingErrorMsg{err: err}
		}
		return recordingStartedMsg{filePath: outputPath}
	}
}

// stopRecordingCmd stops the current recording.
func stopRecordingCmd(pm playback.PlaybackManagerService) tea.Cmd {
	return func() tea.Msg {
		filePath, err := pm.StopRecording()
		if err != nil {
			return recordingErrorMsg{err: err}
		}
		return recordingStoppedMsg{filePath: filePath}
	}
}

// Bookmark and hidden station commands

// toggleBookmarkCmd toggles the bookmark status of a station.
func toggleBookmarkCmd(storage storage.StationStorageService, station common.Station) tea.Cmd {
	return func() tea.Msg {
		var err error
		if storage.IsBookmarked(station.StationUuid) {
			err = storage.RemoveBookmark(station.StationUuid)
		} else {
			err = storage.AddBookmark(station.StationUuid)
		}
		if err != nil {
			return bookmarkToggleFailedMsg{err: err}
		}
		return bookmarkToggledMsg{station: station}
	}
}

// hideStationCmd hides a station from search results.
func hideStationCmd(storage storage.StationStorageService, station common.Station, cursor int) tea.Cmd {
	return func() tea.Msg {
		err := storage.AddHidden(station.StationUuid)
		if err != nil {
			return hideStationFailedMsg{err: err}
		}
		return stationHiddenMsg{station: station, cursor: cursor}
	}
}

// unhideStationCmd removes a station from the hidden list.
func unhideStationCmd(storage storage.StationStorageService, station common.Station) tea.Cmd {
	return func() tea.Msg {
		err := storage.RemoveHidden(station.StationUuid)
		if err != nil {
			return unhideStationFailedMsg{err: err}
		}
		return stationUnhiddenMsg{station: station}
	}
}

// fetchBookmarksCmd fetches all bookmarked stations from storage and the API.
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

// fetchBookmarksForSearchCmd fetches bookmarks and switches directly to bookmarks view.
// Used when accessing bookmarks from the search screen.
func fetchBookmarksForSearchCmd(browser api.RadioBrowserService, storage storage.StationStorageService) tea.Cmd {
	return func() tea.Msg {
		uuids, err := storage.GetBookmarks()
		if err != nil {
			return switchToErrorModelMsg{err: err.Error(), recoverable: true}
		}
		if len(uuids) == 0 {
			return switchToBookmarksMsg{stations: []common.Station{}}
		}
		stations, err := browser.GetStationsByUUIDs(uuids)
		if err != nil {
			return switchToErrorModelMsg{err: err.Error(), recoverable: true}
		}
		return switchToBookmarksMsg{stations: stations}
	}
}

// fetchHiddenStationsCmd fetches all hidden stations from storage and the API.
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

// refetchStationsCmd refetches search results from the API using the stored query.
func refetchStationsCmd(browser api.RadioBrowserService, query common.StationQuery, queryText string) tea.Cmd {
	return func() tea.Msg {
		stations, err := browser.GetStations(query, queryText, "votes", true, 0, 100, true)
		if err != nil {
			return stationsRefetchFailedMsg{err: err}
		}
		return stationsRefetchedMsg{stations: stations}
	}
}

// Vote commands

const voteCooldownDuration = 10 * time.Minute

// voteStationCmd votes for a station via the RadioBrowser API.
// It first checks the global cooldown, then makes the API call and records the timestamp.
// RadioBrowser enforces a 10-minute cooldown per IP for all votes (not per-station).
func voteStationCmd(
	browser api.RadioBrowserService,
	stor storage.StationStorageService,
	station common.Station,
	cursor int,
) tea.Cmd {
	return func() tea.Msg {
		// Check global cooldown (API enforces per-IP, not per-station)
		if lastVote, exists := stor.GetLastVoteTimestamp(); exists {
			elapsed := time.Since(lastVote)
			if elapsed < voteCooldownDuration {
				remaining := int((voteCooldownDuration - elapsed).Minutes()) + 1
				return voteCooldownMsg{remainingMinutes: remaining}
			}
		}

		// Call API
		resp, err := browser.VoteStation(station)
		if err != nil {
			return voteFailedMsg{err: err}
		}
		if !resp.Ok {
			return voteFailedMsg{err: errors.New(resp.Message)}
		}

		// Record global timestamp (ignore error - vote succeeded on API side)
		_ = stor.SetLastVoteTimestamp(time.Now())

		return voteSucceededMsg{station: station, cursor: cursor}
	}
}

// UI commands

// errorClearDelay is the duration to wait before clearing non-fatal errors.
const errorClearDelay = 3 * time.Second

// clearErrorAfterDelayCmd returns a tea.Cmd that clears non-fatal errors after a delay.
// This helper reduces duplication of the error-clearing pattern used throughout the handlers.
func clearErrorAfterDelayCmd() tea.Cmd {
	return tea.Tick(errorClearDelay, func(t time.Time) tea.Msg {
		return clearNonFatalError{}
	})
}

// updateCommandsCmd returns a command that updates the bottom bar with appropriate commands
// based on the current view mode and playback state.
func updateCommandsCmd(viewMode stationsViewMode, isPlaying bool, volume int, volumeIsPercentage bool, isRecording bool, kb config.Keybindings) tea.Cmd {
	return func() tea.Msg {

		// Row 1: Navigation and playback
		var commands []string
		if viewMode == viewModeSearchResults {
			commands = []string{
				i18n.Tf("cmd_quit", map[string]interface{}{"Key": kb.Quit}),
				i18n.Tf("cmd_search", map[string]interface{}{"Key": kb.Search}),
				i18n.T("cmd_enter_play"),
				i18n.T("cmd_move"),
			}
		} else {
			commands = []string{
				i18n.Tf("cmd_quit", map[string]interface{}{"Key": kb.Quit}),
				i18n.Tf("cmd_back", map[string]interface{}{"Key": kb.BookmarksView}),
				i18n.T("cmd_enter_play"),
				i18n.T("cmd_move"),
			}
		}

		var volumeDisplay string
		if volume == 0 {
			volumeDisplay = i18n.T("volume_mute")
		} else {
			if volumeIsPercentage {
				volumeDisplay = i18n.Tf("volume_display_percent", map[string]interface{}{"Volume": volume})
			} else {
				volumeDisplay = i18n.Tf("volume_display", map[string]interface{}{"Volume": volume})
			}
		}

		if isPlaying {
			if isRecording {
				commands = append(commands,
					i18n.Tf("cmd_stop_record", map[string]interface{}{"Key": kb.Record}),
					i18n.Tf("cmd_stop", map[string]interface{}{"Key": kb.StopPlayback}),
					i18n.Tf("cmd_volume", map[string]interface{}{"VolumeDown": kb.VolumeDown, "VolumeUp": kb.VolumeUp}),
					volumeDisplay,
				)
			} else {
				commands = append(commands,
					i18n.Tf("cmd_record", map[string]interface{}{"Key": kb.Record}),
					i18n.Tf("cmd_stop", map[string]interface{}{"Key": kb.StopPlayback}),
					i18n.Tf("cmd_volume", map[string]interface{}{"VolumeDown": kb.VolumeDown, "VolumeUp": kb.VolumeUp}),
					volumeDisplay,
				)
			}
		} else {
			commands = append(commands,
				i18n.Tf("cmd_volume", map[string]interface{}{"VolumeDown": kb.VolumeDown, "VolumeUp": kb.VolumeUp}),
				volumeDisplay,
			)
		}

		// Row 2: Bookmark/hide commands
		var secondaryCommands []string
		if viewMode == viewModeSearchResults {
			secondaryCommands = []string{
				i18n.Tf("cmd_bookmark", map[string]interface{}{"Key": kb.BookmarkToggle}),
				i18n.Tf("cmd_bookmarks", map[string]interface{}{"Key": kb.BookmarksView}),
				i18n.Tf("cmd_vote", map[string]interface{}{"Key": kb.Vote}),
				i18n.Tf("cmd_hide", map[string]interface{}{"Key": kb.HideStation}),
				i18n.Tf("cmd_manage_hidden", map[string]interface{}{"Key": kb.ManageHidden}),
			}
		} else {
			// "B: back" is already in primary row, no hide commands in bookmarks mode
			secondaryCommands = []string{i18n.Tf("cmd_bookmark", map[string]interface{}{"Key": kb.BookmarkToggle})}
		}

		return bottomBarUpdateMsg{
			commands:          commands,
			secondaryCommands: secondaryCommands,
		}
	}
}
