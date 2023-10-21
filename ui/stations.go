// Copyright (c) 2023 Matteo Pacini
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

package ui

import (
	"fmt"
	"radiogogo/api"
	"radiogogo/playback"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StationsModel struct {
	stations              []api.Station
	stationsTable         table.Model
	currentStation        api.Station
	currentStationSpinner spinner.Model
	volume                int
	err                   string

	browser         *api.RadioBrowser
	playbackManager playback.PlaybackManager
	width           int
	height          int
}

func NewStationsModel(
	browser *api.RadioBrowser,
	playbackManager playback.PlaybackManager,
	stations []api.Station,
) StationsModel {

	return StationsModel{
		stations:        stations,
		stationsTable:   newStationsTableModel(stations),
		volume:          80,
		browser:         browser,
		playbackManager: playbackManager,
	}
}

func newStationsTableModel(stations []api.Station) table.Model {

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

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#5a4f9f")).
		Bold(false)
	t.SetStyles(s)

	return t

}

// Messages

type playbackStartedMsg struct {
	station api.Station
}
type playbackStoppedMsg struct{}

type nonFatalError struct {
	stopPlayback bool
	err          error
}
type clearNonFatalError struct{}

// Commands

func playStationCmd(
	playbackManager playback.PlaybackManager,
	station api.Station,
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

func stopStationCmd(playbackManager playback.PlaybackManager) tea.Cmd {
	return func() tea.Msg {
		err := playbackManager.StopStation()
		if err != nil {
			return nonFatalError{stopPlayback: false, err: err}
		}
		return playbackStoppedMsg{}
	}
}

func notifyRadioBrowserCmd(browser *api.RadioBrowser, station api.Station) tea.Cmd {
	return func() tea.Msg {
		_, err := browser.ClickStation(station)
		if err != nil {
			return nonFatalError{stopPlayback: false, err: err}
		}
		return nil
	}
}

func updateCommandsCmd(isPlaying bool, volume int) tea.Cmd {
	return func() tea.Msg {

		commands := []string{"q: quit", "s: search", "enter: play", "↑/↓: move"}

		if isPlaying {
			commands = append(commands, "ctrl+k: stop")
		} else {
			commands = append(commands, "9/0: vol down/up", "vol: "+fmt.Sprintf("%d", volume))
		}

		return bottomBarUpdateMsg{
			commands: commands,
		}
	}
}

// Model

func (m StationsModel) Init() tea.Cmd {
	return updateCommandsCmd(false, m.volume)
}

func (m StationsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case playbackStartedMsg:
		m.currentStation = msg.station
		m.currentStationSpinner = spinner.New()
		m.currentStationSpinner.Spinner = spinner.Dot
		m.currentStationSpinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#5a4f9f"))
		return m, tea.Batch(
			m.currentStationSpinner.Tick,
			notifyRadioBrowserCmd(m.browser, m.currentStation),
			updateCommandsCmd(true, m.volume),
		)
	case playbackStoppedMsg:
		m.currentStation = api.Station{}
		m.currentStationSpinner = spinner.Model{}
		return m, updateCommandsCmd(false, m.volume)
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
	case tea.KeyMsg:
		switch msg.String() {
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
			if m.volume > 0 && !m.playbackManager.IsPlaying() {
				m.volume -= 10
				return m, updateCommandsCmd(false, m.volume)
			}
			return m, nil
		case "0":
			if m.volume < 100 && !m.playbackManager.IsPlaying() {
				m.volume += 10
				return m, updateCommandsCmd(false, m.volume)
			}
			return m, nil
		case "enter":
			station := m.stations[m.stationsTable.Cursor()]
			return m, playStationCmd(m.playbackManager, station, m.volume)
		}
	}

	var cmds []tea.Cmd

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
		extraBar += StyleSetError(m.err)
	} else if m.playbackManager.IsPlaying() {
		extraBar +=
			m.currentStationSpinner.View() +
				StyleSetPlaying("Listening to: "+m.currentStation.Name)
	}

	v := "\n" + m.stationsTable.View() + "\n\n"

	vHeight := lipgloss.Height(v)

	v += lipgloss.NewStyle().
		PaddingTop(m.height - vHeight).
		Render(extraBar)

	return v
}
