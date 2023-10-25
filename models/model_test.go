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
	"testing"

	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/mocks"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestCheckIfPlaybackIsPossibleCmd(t *testing.T) {

	t.Run("returns switchToErrorModelMsg if playback is not available", func(t *testing.T) {

		playbackManager := mocks.MockPlaybackManagerService{
			IsAvailableResult: false,
		}

		msg := checkIfPlaybackIsPossibleCmd(&playbackManager)()

		assert.IsType(t, switchToErrorModelMsg{}, msg)

	})

	t.Run("returns switchToSearchModelMsg if playback is available", func(t *testing.T) {

		playbackManager := mocks.MockPlaybackManagerService{
			IsAvailableResult: true,
		}

		msg := checkIfPlaybackIsPossibleCmd(&playbackManager)()

		assert.IsType(t, switchToSearchModelMsg{}, msg)

	})

}

func TestModel_Init(t *testing.T) {

	t.Run("starts the search model if playback is available", func(t *testing.T) {

		playbackManager := mocks.MockPlaybackManagerService{
			IsAvailableResult: true,
		}

		browser := mocks.MockRadioBrowserService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)

		cmd := model.Init()
		assert.NotNil(t, cmd)

		msg := cmd()

		assert.IsType(t, switchToSearchModelMsg{}, msg)

	})

	t.Run("starts the error model if playback is not available", func(t *testing.T) {

		playbackManager := mocks.MockPlaybackManagerService{
			IsAvailableResult: false,
		}

		browser := mocks.MockRadioBrowserService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)

		cmd := model.Init()
		assert.NotNil(t, cmd)

		msg := cmd()

		assert.IsType(t, switchToErrorModelMsg{}, msg)

	})

}

func TestModel_Update(t *testing.T) {

	t.Run("stores terminal size changes and returns a nil command", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)

		msg := tea.WindowSizeMsg{Width: 100, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 100, newModel.(Model).width)
		assert.Equal(t, 100, newModel.(Model).height)
		assert.Nil(t, cmd)

	})

	t.Run("propagates adjusted terminal size changes to SearchModel if active", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)
		model.state = searchState

		msg := tea.WindowSizeMsg{Width: 100, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 100, newModel.(Model).searchModel.width)
		assert.Equal(t, 98 /* -2 for top and bottom bars */, newModel.(Model).searchModel.height)

		assert.Nil(t, cmd)

	})

	t.Run("propagates adjusted terminal size changes to ErrorModel if active", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)
		model.state = errorState

		msg := tea.WindowSizeMsg{Width: 100, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 100, newModel.(Model).errorModel.width)
		assert.Equal(t, 98 /* -2 for top and bottom bars */, newModel.(Model).errorModel.height)

		assert.Nil(t, cmd)

	})

	t.Run("propagates adjusted terminal size changes to LoadingModel if active", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)
		model.state = loadingState

		msg := tea.WindowSizeMsg{Width: 100, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 100, newModel.(Model).loadingModel.width)
		assert.Equal(t, 98 /* -2 for top and bottom bars */, newModel.(Model).loadingModel.height)

		assert.Nil(t, cmd)

	})

	t.Run("propagates adjusted terminal size changes to StationsModel if active", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)
		model.state = stationsState

		msg := tea.WindowSizeMsg{Width: 100, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 100, newModel.(Model).stationsModel.width)
		assert.Equal(t, 98 /* -2 for top and bottom bars */, newModel.(Model).stationsModel.height)

		assert.Nil(t, cmd)

	})

	t.Run("broadcasts a tea.QuitMsg command if a quitMsg is received", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)

		msg := quitMsg{}

		_, cmd := model.Update(tea.Msg(msg))

		returnedMsg := cmd()

		assert.IsType(t, tea.QuitMsg{}, returnedMsg)

	})

	t.Run("stores bottom bar commands and returns a nil command", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)

		msg := bottomBarUpdateMsg{commands: []string{"test"}}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, []string{"test"}, newModel.(Model).bottomBarCommands)
		assert.Nil(t, cmd)

	})

	t.Run("recreates and switches to search model if switchToSearchModelMsg is received", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)
		model.searchModel.width = 111

		msg := switchToSearchModelMsg{}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, searchState, newModel.(Model).state)
		assert.Equal(t, 0, newModel.(Model).searchModel.width)
		assert.NotNil(t, cmd)

	})

	t.Run("recreates and switches to loading model if switchToLoadingModelMsg is received", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}

		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)
		model.loadingModel.queryText = "test"

		msg := switchToLoadingModelMsg{queryText: "test2"}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, loadingState, newModel.(Model).state)
		assert.Equal(t, "test2", newModel.(Model).loadingModel.queryText)
		assert.NotNil(t, cmd)

	})

	t.Run("recreates and switches to stations model if switchToStationsModelMsg is received", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)
		model.stationsModel.volume = 1

		msg := switchToStationsModelMsg{}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, stationsState, newModel.(Model).state)
		assert.NotEqual(t, 1, newModel.(Model).stationsModel.volume)
		assert.NotNil(t, cmd)

	})

	t.Run("recreates and switches to error model if switchToErrorModelMsg is received", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager)
		model.errorModel.message = "test"

		msg := switchToErrorModelMsg{err: "test2"}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, errorState, newModel.(Model).state)
		assert.Equal(t, "test2", newModel.(Model).errorModel.message)
		assert.NotNil(t, cmd)

	})

}
