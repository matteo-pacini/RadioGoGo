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

package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zi0p4tch0/radiogogo/data"
	"github.com/zi0p4tch0/radiogogo/playback"
)

type HeaderModel struct {
	theme Theme

	width         int
	showOffset    bool
	stationOffset int
	totalStations int
}

func NewHeaderModel(theme Theme, playbackManager playback.PlaybackManagerService) HeaderModel {
	return HeaderModel{
		theme: theme,
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
	}
	return m, nil
}

func (m HeaderModel) View() string {

	header := m.theme.PrimaryBlock.Render("radiogogo")
	version := m.theme.SecondaryBlock.Render(fmt.Sprintf("v%s", data.Version))

	leftHeader := header + version

	if m.showOffset {

		rightHeader := m.theme.PrimaryBlock.Render(fmt.Sprintf("%d/%d", m.stationOffset+1, m.totalStations))

		fillerWidth := m.width - lipgloss.Width(leftHeader) - lipgloss.Width(rightHeader)
		filler := lipgloss.NewStyle().Width(fillerWidth).Render(" ")

		return leftHeader + filler + rightHeader + "\n"

	} else {

		return leftHeader + "\n"
	}

}
