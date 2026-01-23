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
	"github.com/zi0p4tch0/radiogogo/i18n"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// handleHiddenModalInput processes keyboard input when the hidden stations modal is open.
// Returns true if the input was handled (modal is showing), false otherwise.
func (m *StationsModel) handleHiddenModalInput(msg tea.KeyMsg) (bool, tea.Cmd) {
	if !m.showHiddenModal {
		return false, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.hiddenModalCursor > 0 {
			m.hiddenModalCursor--
		}
		return true, nil
	case "down", "j":
		if m.hiddenModalCursor < len(m.hiddenStations)-1 {
			m.hiddenModalCursor++
		}
		return true, nil
	case "enter":
		if len(m.hiddenStations) > 0 {
			station := m.hiddenStations[m.hiddenModalCursor]
			return true, unhideStationCmd(m.storage, station)
		}
		return true, nil
	case "esc", "H", "q":
		m.showHiddenModal = false
		// Trigger refetch if any stations were unhidden
		if m.needsRefetch {
			m.needsRefetch = false
			return true, tea.Batch(
				refetchStationsCmd(m.browser, m.lastQuery, m.lastQueryText),
				updateCommandsCmd(m.viewMode, m.playbackManager.IsPlaying(), m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording()),
			)
		}
		return true, updateCommandsCmd(m.viewMode, m.playbackManager.IsPlaying(), m.volume, m.playbackManager.VolumeIsPercentage(), m.playbackManager.IsRecording())
	}
	return true, nil
}

// renderWithModal renders the base view with the hidden stations modal overlay.
// The modal is centered on the screen and displays the list of hidden stations
// with navigation controls.
func (m StationsModel) renderWithModal(baseView string) string {
	modalContent := m.theme.SecondaryText.Bold(true).Render(i18n.T("hidden_stations_title")) + "\n\n"

	if len(m.hiddenStations) == 0 {
		modalContent += m.theme.TertiaryText.Render(i18n.T("no_hidden_stations"))
	} else {
		for i, station := range m.hiddenStations {
			cursor := "  "
			if i == m.hiddenModalCursor {
				cursor = "> "
			}
			// Truncate long station names to fit in modal
			name := station.Name
			if len(name) > 40 {
				name = name[:37] + "..."
			}
			modalContent += cursor + name + "\n"
		}
	}
	modalContent += "\n" + m.theme.TertiaryText.Render(i18n.T("hidden_modal_help"))

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.SecondaryColor)).
		Padding(1, 2).
		Width(50)

	modal := modalStyle.Render(modalContent)

	// Center the modal on the screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}
