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
	"os/exec"
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
	currentFfplay         *exec.Cmd
	currentStation        api.Station
	currentStationSpinner spinner.Model
	volume                int
	err                   string

	browser *api.RadioBrowser
}

func NewStationsModel(browser *api.RadioBrowser, stations []api.Station) StationsModel {

	return StationsModel{
		stations:      stations,
		stationsTable: newStationsTableModel(stations),
		volume:        80,
		browser:       browser,
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

// Playback

type playbackStartedMsg struct {
	station api.Station
	cmd     *exec.Cmd
}
type playbackStoppedMsg struct{}

func runFfplay(station api.Station, volume int) tea.Cmd {
	return func() tea.Msg {
		cmd, err := playback.FFPlayPlayStation(station, volume)
		if err != nil {
			return nonFatalError{stopPlayback: false, err: err}
		}
		return playbackStartedMsg{station: station, cmd: cmd}
	}
}

func killFfplay(cmd *exec.Cmd) tea.Cmd {
	return func() tea.Msg {
		if cmd == nil {
			return nil
		}
		err := cmd.Process.Kill()
		if err != nil {
			return switchToErrorModelMsg{err: err.Error()}
		}
		_, err = cmd.Process.Wait()
		if err != nil {
			return switchToErrorModelMsg{err: err.Error()}
		}
		return playbackStoppedMsg{}
	}
}

func notifyRadioBrowser(browser *api.RadioBrowser, station api.Station) tea.Cmd {
	return func() tea.Msg {
		_, err := browser.ClickStation(station)
		if err != nil {
			return nonFatalError{stopPlayback: false, err: err}
		}
		return nil
	}
}

// Error messages

type nonFatalError struct {
	stopPlayback bool
	err          error
}
type clearNonFatalError struct{}

// Model

func (m StationsModel) Init() tea.Cmd {
	return nil
}

func (m StationsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case playbackStartedMsg:
		m.currentStation = msg.station
		m.currentFfplay = msg.cmd
		m.currentStationSpinner = spinner.New()
		m.currentStationSpinner.Spinner = spinner.Dot
		m.currentStationSpinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#5a4f9f"))
		return m, tea.Batch(
			m.currentStationSpinner.Tick,
			notifyRadioBrowser(m.browser, m.currentStation),
		)
	case playbackStoppedMsg:
		m.currentStation = api.Station{}
		m.currentFfplay = nil
		m.currentStationSpinner = spinner.Model{}
		return m, nil
	case nonFatalError:
		var cmds []tea.Cmd
		if msg.stopPlayback {
			cmds = append(cmds, killFfplay(m.currentFfplay))
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
			if m.currentFfplay != nil {
				return m, killFfplay(m.currentFfplay)
			}
		case "q":
			return m, tea.Sequence(killFfplay(m.currentFfplay), tea.Quit)
		case "s":
			return m, tea.Sequence(killFfplay(m.currentFfplay), func() tea.Msg {
				return switchToSearchModelMsg{}
			})
		case "9":
			if m.volume > 0 {
				m.volume -= 10
			}
			return m, nil
		case "0":
			if m.volume < 100 {
				m.volume += 10
			}
			return m, nil
		case "enter":
			station := m.stations[m.stationsTable.Cursor()]
			return m, tea.Sequence(killFfplay(m.currentFfplay), runFfplay(station, m.volume))
		}
	}

	var cmds []tea.Cmd

	if m.currentFfplay != nil {
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

	playingBar := ""

	if m.err != "" {
		playingBar += StyleSetError(m.err)
		playingBar += "\n"
	} else if m.currentFfplay != nil {
		playingBar +=
			m.currentStationSpinner.View() +
				StyleSetPlaying("now playing: "+m.currentStation.Name)
		playingBar += "\n"
	}

	commands := []string{"q: quit", "s: search", "enter: play", "↑/↓: navigate"}

	if m.currentFfplay != nil {
		commands = append(commands, "ctrl+k: stop playing")
	} else {
		commands = append(commands, "9/0: volume down/up", "volume: "+fmt.Sprintf("%d", m.volume))
	}

	return m.stationsTable.View() + "\n\n" + playingBar + StyleBottomBar(commands)
}
