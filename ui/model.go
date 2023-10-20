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
	"radiogogo/api"
	"radiogogo/playback"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type modelState int

const (
	searchState modelState = iota
	errorState
	loadingState
	stationsState
)

// State switching messages

type switchToErrorModelMsg struct {
	err string
}
type switchToSearchModelMsg struct {
}
type switchToLoadingModelMsg struct {
	query string
}
type switchToStationsModelMsg struct {
	stations []api.Station
}

// Dependency injection messages

type radioBrowserReadyMsg struct {
	browser *api.RadioBrowser
}

// UI messages

type bottomBarUpdateMsg struct {
	commands []string
}

// Quit message

type quitMsg struct{}

func radiogogoQuit() tea.Msg {
	return quitMsg{}
}

// Model

type Model struct {
	// Models
	state         modelState
	searchModel   SearchModel
	errorModel    ErrorModel
	loadingModel  LoadingModel
	stationsModel StationsModel

	// State
	width             int
	height            int
	browser           *api.RadioBrowser
	bottomBarCommands []string
}

func NewModel() Model {
	return Model{
		state: searchState,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Sequence(func() tea.Msg {
		if !playback.IsFFplayAvailable() {
			var osSpecific string
			switch runtime.GOOS {
			case "darwin":
				osSpecific = "\n\nFFmpeg can be installed with Homebrew: 'brew install ffmpeg'."
			case "linux":
				osSpecific = "\n\nFFmpeg can be installed with your distro's package manager, such as `apt` for Debian/Ubuntu or `dnf` for Fedora."
			case "windows":
				osSpecific = "\n\nFFmpeg can be installed with Chocolatey: choco install ffmpeg."
			case "freebsd":
				osSpecific = "\n\nFFmpeg can be installed with pkg: pkg install ffmpeg."
			case "netbsd":
				osSpecific = "\n\nFFmpeg can be installed with pkgsrc: pkgin install ffmpeg."
			case "openbsd":
				osSpecific = "\n\nFFmpeg can be installed with pkg_add: pkg_add ffmpeg."
			default:
				osSpecific = "\n\n(Sorry, FFmpeg installation instructions are not available for your operating system)"
			}

			return switchToErrorModelMsg{
				err: `RadioGoGo requires "ffplay" (part of "ffmpeg") to be installed and available in your PATH.` + osSpecific,
			}
		}
		return nil
	}, func() tea.Msg {
		browser, err := api.NewRadioBrowser()
		if err != nil {
			return switchToErrorModelMsg{err: err.Error()}
		}
		return radioBrowserReadyMsg{browser: browser}
	}, func() tea.Msg {
		return switchToSearchModelMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	// Top-level messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case quitMsg:
		return m, tea.Quit
	case radioBrowserReadyMsg:
		m.browser = msg.browser
		return m, nil
	case bottomBarUpdateMsg:
		m.bottomBarCommands = msg.commands
		return m, nil
	}

	// State transitions

	switch msg := msg.(type) {
	case switchToSearchModelMsg:
		m.searchModel = NewSearchModel()
		m.state = searchState
		return m, m.searchModel.Init()
	case switchToLoadingModelMsg:
		m.loadingModel = NewLoadingModel(m.browser, msg.query)
		m.state = loadingState
		return m, m.loadingModel.Init()
	case switchToStationsModelMsg:
		m.stationsModel = NewStationsModel(m.browser, msg.stations)
		m.state = stationsState
		return m, m.stationsModel.Init()
	case switchToErrorModelMsg:
		m.errorModel = NewErrorModel(msg.err)
		m.state = errorState
		return m, m.errorModel.Init()
	}

	// State handling

	switch m.state {
	case searchState:
		newSearchModel, cmd := m.searchModel.Update(msg)
		m.searchModel = newSearchModel.(SearchModel)
		return m, cmd
	case loadingState:
		newLoadingModel, cmd := m.loadingModel.Update(msg)
		m.loadingModel = newLoadingModel.(LoadingModel)
		return m, cmd
	case stationsState:
		newStationsModel, cmd := m.stationsModel.Update(msg)
		m.stationsModel = newStationsModel.(StationsModel)
		return m, cmd
	case errorState:
		newErrorModel, cmd := m.errorModel.Update(msg)
		m.errorModel = newErrorModel.(ErrorModel)
		return m, cmd
	}

	return m, cmd
}

func (m Model) View() string {

	var view string

	view = Header()

	effectiveHeight := m.height - 2 // 2 = header height + bottom bar height

	var currentView string

	switch m.state {
	case searchState:
		m.searchModel.width = m.width
		m.searchModel.height = effectiveHeight
		currentView = m.searchModel.View()
	case loadingState:
		currentView = m.loadingModel.View()
	case stationsState:
		m.stationsModel.width = m.width
		m.stationsModel.height = effectiveHeight
		currentView = m.stationsModel.View()
	case errorState:
		currentView = m.errorModel.View()
	}

	currentViewHeight := lipgloss.Height(currentView)

	// Render the current view

	view += currentView

	// Push the view down to the bottom of the terminal

	view += lipgloss.NewStyle().
		Height(m.height - currentViewHeight).
		Render()

	// Render bottom bar

	view += StyleBottomBar(m.bottomBarCommands)

	return view
}
