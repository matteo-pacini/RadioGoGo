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

func (m HeaderModel) View() string {

	header := m.theme.PrimaryBlock.Render("radiogogo")
	version := m.theme.SecondaryBlock.Render(fmt.Sprintf("v%s", data.Version))

	// Only show playback/rec indicators in stations view (when showOffset is true)
	if !m.showOffset {
		return header + version + "\n"
	}

	// Base style with background but no padding (we'll add padding to outer parts only)
	baseStyle := m.theme.PrimaryBlock.Copy().PaddingLeft(0).PaddingRight(0)

	// Playback status indicator
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
	playbackIndicator := baseStyle.Copy().PaddingLeft(2).Render("(") +
		playbackDotStyle.Render("●") +
		baseStyle.Copy().PaddingRight(2).Render(") "+m.playerName)

	// Recording status indicator
	var recDotColor lipgloss.Color
	if m.isRecording {
		recDotColor = lipgloss.Color("196") // red
	} else {
		recDotColor = lipgloss.Color("252") // white/gray
	}

	recDotStyle := baseStyle.Copy().Foreground(recDotColor)
	recIndicator := baseStyle.Render("(") +
		recDotStyle.Render("●") +
		baseStyle.Copy().PaddingRight(2).Render(") rec")

	leftHeader := header + version + playbackIndicator + recIndicator

	rightHeader := m.theme.PrimaryBlock.Render(fmt.Sprintf("%d/%d", m.stationOffset+1, m.totalStations))

	fillerWidth := m.width - lipgloss.Width(leftHeader) - lipgloss.Width(rightHeader)
	if fillerWidth < 0 {
		fillerWidth = 0
	}
	filler := lipgloss.NewStyle().Width(fillerWidth).Render(" ")

	return leftHeader + filler + rightHeader + "\n"

}
