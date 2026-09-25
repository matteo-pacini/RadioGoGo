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
	"strings"
	"testing"

	"github.com/zi0p4tch0/radiogogo/i18n"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestModel_RemoteClientsChangedMsg(t *testing.T) {
	model := updateModel(newSnapshotTestModel(), switchToSearchModelMsg{})
	searchBefore := model.searchModel

	newModel, cmd := model.Update(RemoteClientsChangedMsg{Count: 2})
	model = newModel.(Model)

	assert.Equal(t, 2, model.remoteClients)
	assert.Nil(t, cmd)
	assert.Equal(t, searchBefore, model.searchModel, "message must not reach the child model")
}

func TestModel_RemoteIndicator(t *testing.T) {
	indicator := i18n.T("remote_attached")

	for _, state := range []tea.Msg{switchToSearchModelMsg{}, switchToStationsModelMsg{}} {
		model := updateModel(newSnapshotTestModel(), state)
		model = updateModel(model, bottomBarUpdateMsg{commands: []string{"q: quit"}})
		assert.NotContains(t, model.View(), indicator)

		model = updateModel(model, RemoteClientsChangedMsg{Count: 1})
		assert.Contains(t, model.View(), indicator)

		model = updateModel(model, RemoteClientsChangedMsg{Count: 0})
		assert.NotContains(t, model.View(), indicator)
	}
}

func TestModel_RemoteIndicator_HeightUnchanged(t *testing.T) {
	model := updateModel(newSnapshotTestModel(), tea.WindowSizeMsg{Width: minTerminalWidth, Height: minTerminalHeight})
	model = updateModel(model, switchToSearchModelMsg{})
	model = updateModel(model, bottomBarUpdateMsg{commands: []string{"q: quit"}})
	before := lipgloss.Height(model.View())

	model = updateModel(model, RemoteClientsChangedMsg{Count: 1})

	assert.Equal(t, before, lipgloss.Height(model.View()))
}

func TestModel_RemoteIndicator_DroppedWhenRowOverflows(t *testing.T) {
	model := updateModel(newSnapshotTestModel(), tea.WindowSizeMsg{Width: minTerminalWidth, Height: minTerminalHeight})
	model = updateModel(model, switchToSearchModelMsg{})
	wide := strings.Repeat("x", minTerminalWidth-6)
	model = updateModel(model, bottomBarUpdateMsg{commands: []string{wide}})
	model = updateModel(model, RemoteClientsChangedMsg{Count: 1})

	view := model.View()
	assert.NotContains(t, view, i18n.T("remote_attached"))
	assert.Contains(t, view, wide)
}
