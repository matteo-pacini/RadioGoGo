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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zi0p4tch0/radiogogo/data"
	"github.com/zi0p4tch0/radiogogo/i18n"
	"github.com/zi0p4tch0/radiogogo/playback"
)

// PlaybackStatus represents the current state of the audio player
type PlaybackStatus int

const (
	PlaybackIdle PlaybackStatus = iota
	PlaybackPlaying
	PlaybackRestarting
)

// playbackStatusMsg is sent to update the header's playback status indicator
type playbackStatusMsg struct {
	status PlaybackStatus
}

// recordingStatusMsg is sent to update the header's recording indicator
type recordingStatusMsg struct {
	isRecording bool
}

type HeaderModel struct {
	theme Theme

	width          int
	showOffset     bool
	stationOffset  int
	totalStations  int
	playbackStatus PlaybackStatus
	playerName     string
	isRecording    bool
}

func NewHeaderModel(theme Theme, playbackManager playback.PlaybackManagerService) HeaderModel {
	return HeaderModel{
		theme:      theme,
		playerName: playbackManager.Name(),
	}
}

func (m HeaderModel) Init() tea.Cmd {
	return nil
}

func (m HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stationCursorMovedMsg:
		m.stationOffset = msg.offset
		m.totalStations = msg.totalStations
	case playbackStatusMsg:
		m.playbackStatus = msg.status
	case recordingStatusMsg:
		m.isRecording = msg.isRecording
	}
	return m, nil
}

// View renders the header bar with app name, version, and status indicators.
// Layout: [radiogogo][v0.x.x][(●) ffplay][(●) rec] ... [1/100]
//
// The header adapts based on context:
//   - In search/loading views: Shows only app name and version
//   - In stations view: Shows full header with playback/recording indicators
//
// Status indicator colors:
//   - Playback dot: white (idle), green (playing), yellow (restarting)
//   - Recording dot: white (not recording), red (recording)
func (m HeaderModel) View() string {

	header := m.theme.PrimaryBlock.Render("radiogogo")
	versionStr := data.Version
	if versionStr != "dev" {
		versionStr = "v" + versionStr
	}
	version := m.theme.SecondaryBlock.Render(versionStr)

	// In non-stations views (search, loading, error), show minimal header
	if !m.showOffset {
		return header + version + "\n"
	}

	// Create a base style with matching background but no padding.
	// We apply padding selectively to create proper spacing between elements
	// while maintaining a continuous colored background across all indicators.
	baseStyle := m.theme.PrimaryBlock.Copy().PaddingLeft(0).PaddingRight(0)

	// Playback status indicator: (●) ffplay
	// Color indicates: idle (white), playing (green), restarting (yellow)
	var playbackDotColor lipgloss.Color
	switch m.playbackStatus {
	case PlaybackIdle:
		playbackDotColor = lipgloss.Color("252") // white/gray
	case PlaybackPlaying:
		playbackDotColor = lipgloss.Color("42") // green
	case PlaybackRestarting:
		playbackDotColor = lipgloss.Color("226") // yellow
	}

	playbackDotStyle := baseStyle.Copy().Foreground(playbackDotColor)
	// Build indicator piece by piece to maintain consistent background
	playbackIndicator := baseStyle.Copy().PaddingLeft(2).Render("(") +
		playbackDotStyle.Render("●") +
		baseStyle.Copy().PaddingRight(2).Render(") "+i18n.T("header_play"))

	// Recording status indicator: (●) rec
	// Color indicates: not recording (white), recording (red)
	var recDotColor lipgloss.Color
	if m.isRecording {
		recDotColor = lipgloss.Color("196") // red
	} else {
		recDotColor = lipgloss.Color("252") // white/gray
	}

	recDotStyle := baseStyle.Copy().Foreground(recDotColor)
	recIndicator := baseStyle.Render("(") +
		recDotStyle.Render("●") +
		baseStyle.Copy().PaddingRight(2).Render(") "+i18n.T("header_recording"))

	// Compose left and right sections
	leftHeader := header + version + playbackIndicator + recIndicator
	rightHeader := m.theme.PrimaryBlock.Render(fmt.Sprintf("%d/%d", m.stationOffset+1, m.totalStations))

	// Fill remaining space to push station counter to the right edge
	fillerWidth := m.width - lipgloss.Width(leftHeader) - lipgloss.Width(rightHeader)
	if fillerWidth < 0 {
		fillerWidth = 0
	}
	filler := lipgloss.NewStyle().Width(fillerWidth).Render(" ")

	return leftHeader + filler + rightHeader + "\n"

}
