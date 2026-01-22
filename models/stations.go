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
	"github.com/zi0p4tch0/radiogogo/assets"
	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/playback"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
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

	browser         api.RadioBrowserService
	playbackManager playback.PlaybackManagerService
	width           int
	height          int
}

func NewStationsModel(
	theme Theme,
	browser api.RadioBrowserService,
	playbackManager playback.PlaybackManagerService,
	stations []common.Station,
) StationsModel {

	return StationsModel{
		theme:           theme,
		stations:        stations,
		stationsTable:   newStationsTableModel(theme, stations),
		volume:          playbackManager.VolumeDefault(),
		browser:         browser,
		playbackManager: playbackManager,
	}
}

func newStationsTableModel(theme Theme, stations []common.Station) table.Model {

	rows := make([]table.Row, len(stations))
	for i, station := range stations {
		rows[i] = table.Row{
			station.Name,
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

func updateCommandsCmd(isPlaying bool, volume int, volumeIsPercentage bool, isRecording bool) tea.Cmd {
	return func() tea.Msg {

		commands := []string{"q: quit", "s: search", "enter: play", "↑/↓: move"}

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
				commands = append(commands, "r: stop rec", "ctrl+k: stop", "9/0: vol down/up", volumeDisplay)
			} else {
				commands = append(commands, "r: record", "ctrl+k: stop", "9/0: vol down/up", volumeDisplay)
			}
		} else {
			commands = append(commands, "9/0: vol down/up", volumeDisplay)
		}

		return bottomBarUpdateMsg{
			commands: commands,
		}
	}
}

// Model

func (m StationsModel) Init() tea.Cmd {
	return tea.Batch(
		updateCommandsCmd(false, m.volume, m.playbackManager.VolumeIsPercentage(), false),
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
			updateCommandsCmd(true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
			func() tea.Msg { return playbackStatusMsg{status: PlaybackPlaying} },
		)
	case playbackStoppedMsg:
		m.currentStation = common.Station{}
		m.currentStationSpinner = spinner.Model{}
		return m, tea.Batch(
			updateCommandsCmd(false, m.volume, m.playbackManager.VolumeIsPercentage(), false),
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
			updateCommandsCmd(true, m.volume, m.playbackManager.VolumeIsPercentage(), true),
			func() tea.Msg { return recordingStatusMsg{isRecording: true} },
		)
	case recordingStoppedMsg:
		return m, tea.Batch(
			updateCommandsCmd(true, m.volume, m.playbackManager.VolumeIsPercentage(), false),
			func() tea.Msg { return recordingStatusMsg{isRecording: false} },
		)
	case recordingErrorMsg:
		m.err = fmt.Sprintf("Recording error: %v", msg.err)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNonFatalError{}
		})
	case tea.KeyMsg:
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
			return m, tea.Sequence(
				stopStationCmd(m.playbackManager),
				func() tea.Msg {
					return switchToSearchModelMsg{}
				},
			)
		case "9":
			if m.volume > m.playbackManager.VolumeMin() {
				m.volume -= 10
				if m.playbackManager.IsPlaying() {
					changeID := time.Now().UnixNano()
					m.pendingVolumeChangeID = changeID
					m.volumeChangePending = true
					return m, tea.Batch(
						updateCommandsCmd(true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
						startVolumeDebounceCmd(changeID),
						func() tea.Msg { return playbackStatusMsg{status: PlaybackRestarting} },
					)
				}
				return m, updateCommandsCmd(false, m.volume, m.playbackManager.VolumeIsPercentage(), false)
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
						updateCommandsCmd(true, m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
						startVolumeDebounceCmd(changeID),
						func() tea.Msg { return playbackStatusMsg{status: PlaybackRestarting} },
					)
				}
				return m, updateCommandsCmd(false, m.volume, m.playbackManager.VolumeIsPercentage(), false)
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

	extraBar := ""

	if m.err != "" {
		extraBar += m.theme.ErrorText.Render(m.err)
	} else if m.playbackManager.IsPlaying() {
		extraBar +=
			m.currentStationSpinner.View() +
				m.theme.SecondaryText.Bold(true).Render("Listening to: "+m.currentStation.Name)
	} else {
		extraBar += m.theme.PrimaryText.Bold(true).Render("It's quiet here, time to play something!")
	}

	var v string
	if len(m.stations) == 0 {
		v = fmt.Sprintf(
			"\n%s\n\n%s\n",
			assets.NoStations,
			m.theme.SecondaryText.Bold(true).Render("No stations found, try another search!"),
		)
	} else {
		v = "\n" + m.stationsTable.View() + "\n"
		v += extraBar
	}

	return v
}

func (m *StationsModel) SetWidthAndHeight(width int, height int) {
	m.width = width
	m.height = height
	m.stationsTable.SetWidth(width)
	m.stationsTable.SetHeight(height - 4)
}
