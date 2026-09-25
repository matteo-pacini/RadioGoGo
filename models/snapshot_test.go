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
	"testing"

	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/mocks"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSnapshotTestModel() Model {
	playbackManager := mocks.MockPlaybackManagerService{IsAvailableResult: true, VolumeDefaultResult: 80}
	model := NewModel(config.Config{}, &mocks.MockRadioBrowserService{}, &playbackManager, &mocks.MockStationStorageService{})
	return updateModel(model, tea.WindowSizeMsg{Width: 120, Height: 40})
}

func updateModel(m Model, msg tea.Msg) Model {
	newModel, _ := m.Update(msg)
	return newModel.(Model)
}

func TestModel_Snapshot_ViewNames(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.Msg
		want string
	}{
		{"boot", nil, "boot"},
		{"search", switchToSearchModelMsg{}, "search"},
		{"loading", switchToLoadingModelMsg{queryText: "jazz"}, "loading"},
		{"stations", switchToStationsModelMsg{}, "stations"},
		{"error", switchToErrorModelMsg{err: "boom"}, "error"},
		{"terminal too small", tea.WindowSizeMsg{Width: 10, Height: 10}, "terminal_too_small"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newSnapshotTestModel()
			if tt.msg != nil {
				model = updateModel(model, tt.msg)
			}
			snap := model.Snapshot()
			assert.Equal(t, tt.want, snap.View)
			if tt.want == "stations" {
				assert.NotNil(t, snap.Stations)
			} else {
				assert.Nil(t, snap.Stations)
			}
		})
	}
}

func TestModel_Snapshot_Size(t *testing.T) {
	snap := newSnapshotTestModel().Snapshot()
	assert.Equal(t, 120, snap.Width)
	assert.Equal(t, 40, snap.Height)
}

func TestModel_Snapshot_Stations(t *testing.T) {
	stations := []common.Station{
		createTestStation("Zero"),
		createTestStation("One"),
		createTestStation("Two"),
	}

	t.Run("cursor and selection follow navigation", func(t *testing.T) {
		model := updateModel(newSnapshotTestModel(), switchToStationsModelMsg{stations: stations})

		snap := model.Snapshot()
		require.NotNil(t, snap.Stations)
		assert.Equal(t, 0, snap.Stations.Cursor)
		require.NotNil(t, snap.Stations.Selected)
		assert.Equal(t, "Zero", snap.Stations.Selected.Name)
		assert.Equal(t, stations[0].StationUuid.String(), snap.Stations.Selected.UUID)

		model = updateModel(model, tea.KeyMsg{Type: tea.KeyDown})
		model = updateModel(model, tea.KeyMsg{Type: tea.KeyDown})

		snap = model.Snapshot()
		assert.Equal(t, 2, snap.Stations.Cursor)
		assert.Equal(t, "Two", snap.Stations.Selected.Name)
	})

	t.Run("empty list has no selection", func(t *testing.T) {
		model := updateModel(newSnapshotTestModel(), switchToStationsModelMsg{})

		snap := model.Snapshot()
		require.NotNil(t, snap.Stations)
		assert.Nil(t, snap.Stations.Selected)
	})

	t.Run("playing, volume, recording and modal", func(t *testing.T) {
		model := updateModel(newSnapshotTestModel(), switchToStationsModelMsg{stations: stations})

		snap := model.Snapshot()
		assert.Nil(t, snap.Stations.Playing)
		assert.False(t, snap.Stations.Recording)
		assert.False(t, snap.Stations.ModalOpen)
		assert.Equal(t, 80, snap.Stations.Volume)

		model = updateModel(model, playbackStartedMsg{station: stations[1]})
		model = updateModel(model, recordingStatusMsg{isRecording: true})
		model.stationsModel.showHiddenModal = true

		snap = model.Snapshot()
		require.NotNil(t, snap.Stations.Playing)
		assert.Equal(t, "One", snap.Stations.Playing.Name)
		assert.Equal(t, stations[1].StationUuid.String(), snap.Stations.Playing.UUID)
		assert.True(t, snap.Stations.Recording)
		assert.True(t, snap.Stations.ModalOpen)

		model = updateModel(model, playbackStoppedMsg{})
		assert.Nil(t, model.Snapshot().Stations.Playing)
	})
}
