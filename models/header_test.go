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

	"github.com/stretchr/testify/assert"
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/mocks"
)

func TestNewHeaderModel(t *testing.T) {
	theme := NewTheme(config.NewDefaultConfig())
	mockPM := &mocks.MockPlaybackManagerService{
		NameResult: "testplayer",
	}

	header := NewHeaderModel(theme, mockPM)

	assert.Equal(t, "testplayer", header.playerName)
	assert.Equal(t, PlaybackIdle, header.playbackStatus)
	assert.False(t, header.isRecording)
	assert.False(t, header.showOffset)
}

func TestHeaderModelInit(t *testing.T) {
	theme := NewTheme(config.NewDefaultConfig())
	mockPM := &mocks.MockPlaybackManagerService{
		NameResult: "ffplay",
	}

	header := NewHeaderModel(theme, mockPM)
	cmd := header.Init()

	assert.Nil(t, cmd)
}

func TestHeaderModelUpdate(t *testing.T) {
	theme := NewTheme(config.NewDefaultConfig())
	mockPM := &mocks.MockPlaybackManagerService{
		NameResult: "ffplay",
	}

	t.Run("handles stationCursorMovedMsg", func(t *testing.T) {
		header := NewHeaderModel(theme, mockPM)
		msg := stationCursorMovedMsg{offset: 5, totalStations: 100}

		newModel, cmd := header.Update(msg)
		updatedHeader := newModel.(HeaderModel)

		assert.Equal(t, 5, updatedHeader.stationOffset)
		assert.Equal(t, 100, updatedHeader.totalStations)
		assert.Nil(t, cmd)
	})

	t.Run("handles playbackStatusMsg", func(t *testing.T) {
		header := NewHeaderModel(theme, mockPM)
		msg := playbackStatusMsg{status: PlaybackPlaying}

		newModel, cmd := header.Update(msg)
		updatedHeader := newModel.(HeaderModel)

		assert.Equal(t, PlaybackPlaying, updatedHeader.playbackStatus)
		assert.Nil(t, cmd)
	})

	t.Run("handles recordingStatusMsg", func(t *testing.T) {
		header := NewHeaderModel(theme, mockPM)
		msg := recordingStatusMsg{isRecording: true}

		newModel, cmd := header.Update(msg)
		updatedHeader := newModel.(HeaderModel)

		assert.True(t, updatedHeader.isRecording)
		assert.Nil(t, cmd)
	})
}

func TestHeaderModelView(t *testing.T) {
	theme := NewTheme(config.NewDefaultConfig())
	mockPM := &mocks.MockPlaybackManagerService{
		NameResult: "ffplay",
	}

	t.Run("minimal header when showOffset is false", func(t *testing.T) {
		header := NewHeaderModel(theme, mockPM)
		header.showOffset = false

		view := header.View()

		assert.Contains(t, view, "radiogogo")
		assert.NotContains(t, view, "ffplay")
		assert.True(t, strings.HasSuffix(view, "\n"))
	})

	t.Run("full header when showOffset is true", func(t *testing.T) {
		header := NewHeaderModel(theme, mockPM)
		header.showOffset = true
		header.width = 120
		header.stationOffset = 0
		header.totalStations = 50

		view := header.View()

		assert.Contains(t, view, "radiogogo")
		assert.Contains(t, view, "ffplay")
		assert.Contains(t, view, "1/50")
		assert.True(t, strings.HasSuffix(view, "\n"))
	})

	t.Run("shows recording indicator", func(t *testing.T) {
		header := NewHeaderModel(theme, mockPM)
		header.showOffset = true
		header.width = 120
		header.isRecording = true

		view := header.View()

		assert.Contains(t, view, "rec")
	})
}
