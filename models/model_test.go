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

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})

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

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})

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

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})

		msg := tea.WindowSizeMsg{Width: 100, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 100, newModel.(Model).width)
		assert.Equal(t, 100, newModel.(Model).height)
		assert.Nil(t, cmd)

	})

	t.Run("propagates adjusted terminal size changes to SearchModel if active", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = searchState

		msg := tea.WindowSizeMsg{Width: 120, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 120, newModel.(Model).searchModel.width)
		assert.Equal(t, 98 /* -2 for top and bottom bars */, newModel.(Model).searchModel.height)

		assert.Nil(t, cmd)

	})

	t.Run("propagates adjusted terminal size changes to ErrorModel if active", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = errorState

		msg := tea.WindowSizeMsg{Width: 120, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 120, newModel.(Model).errorModel.width)
		assert.Equal(t, 98 /* -2 for top and bottom bars */, newModel.(Model).errorModel.height)

		assert.Nil(t, cmd)

	})

	t.Run("propagates adjusted terminal size changes to LoadingModel if active", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = loadingState

		msg := tea.WindowSizeMsg{Width: 120, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 120, newModel.(Model).loadingModel.width)
		assert.Equal(t, 98 /* -2 for top and bottom bars */, newModel.(Model).loadingModel.height)

		assert.Nil(t, cmd)

	})

	t.Run("propagates adjusted terminal size changes to StationsModel if active", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = stationsState

		msg := tea.WindowSizeMsg{Width: 120, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 120, newModel.(Model).stationsModel.width)
		assert.Equal(t, 97 /* -3 for header and two-row bottom bar */, newModel.(Model).stationsModel.height)

		assert.Nil(t, cmd)

	})

	t.Run("broadcasts a tea.QuitMsg command if a quitMsg is received", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})

		msg := quitMsg{}

		_, cmd := model.Update(tea.Msg(msg))

		returnedMsg := cmd()

		assert.IsType(t, tea.QuitMsg{}, returnedMsg)

	})

	t.Run("stores bottom bar commands and returns a nil command", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})

		msg := bottomBarUpdateMsg{commands: []string{"test"}}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, []string{"test"}, newModel.(Model).bottomBarCommands)
		assert.Nil(t, cmd)

	})

	t.Run("recreates and switches to search model if switchToSearchModelMsg is received", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
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

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
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

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
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

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.errorModel.message = "test"

		msg := switchToErrorModelMsg{err: "test2"}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, errorState, newModel.(Model).state)
		assert.Equal(t, "test2", newModel.(Model).errorModel.message)
		assert.NotNil(t, cmd)

	})

	t.Run("switches to terminalTooSmallState when width is below minimum", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = searchState

		msg := tea.WindowSizeMsg{Width: 100, Height: 50}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, terminalTooSmallState, newModel.(Model).state)
		assert.Equal(t, searchState, newModel.(Model).previousState)
		assert.Nil(t, cmd)

	})

	t.Run("switches to terminalTooSmallState when height is below minimum", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = stationsState

		msg := tea.WindowSizeMsg{Width: 120, Height: 10}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, terminalTooSmallState, newModel.(Model).state)
		assert.Equal(t, stationsState, newModel.(Model).previousState)
		assert.Nil(t, cmd)

	})

	t.Run("restores previous state when terminal is resized back to adequate size", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = terminalTooSmallState
		model.previousState = searchState

		msg := tea.WindowSizeMsg{Width: 120, Height: 50}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, searchState, newModel.(Model).state)
		assert.Nil(t, cmd)

	})

	t.Run("stays in terminalTooSmallState if still too small", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = terminalTooSmallState
		model.previousState = searchState

		msg := tea.WindowSizeMsg{Width: 80, Height: 20}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, terminalTooSmallState, newModel.(Model).state)
		assert.Nil(t, cmd)

	})

	t.Run("stores cursor movement and returns nil command", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})

		msg := stationCursorMovedMsg{offset: 5, totalStations: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 5, newModel.(Model).headerModel.stationOffset)
		assert.Equal(t, 100, newModel.(Model).headerModel.totalStations)
		assert.Nil(t, cmd)

	})

}

// TestModel_StateTransitionWorkflows tests complete navigation paths through the state machine.
func TestModel_StateTransitionWorkflows(t *testing.T) {

	t.Run("boot -> search -> loading -> stations workflow", func(t *testing.T) {
		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{
			IsAvailableResult: true,
		}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		assert.Equal(t, bootState, model.state)

		// Init triggers playback check
		cmd := model.Init()
		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, switchToSearchModelMsg{}, msg)

		// Transition to search state
		newModel, _ := model.Update(msg)
		model = newModel.(Model)
		assert.Equal(t, searchState, model.state)

		// Simulate search submission -> loading
		loadingMsg := switchToLoadingModelMsg{queryText: "jazz"}
		newModel, _ = model.Update(loadingMsg)
		model = newModel.(Model)
		assert.Equal(t, loadingState, model.state)
		assert.Equal(t, "jazz", model.loadingModel.queryText)

		// Simulate successful search -> stations
		stationsMsg := switchToStationsModelMsg{}
		newModel, _ = model.Update(stationsMsg)
		model = newModel.(Model)
		assert.Equal(t, stationsState, model.state)
	})

	t.Run("boot -> search -> loading -> error -> search workflow (error recovery)", func(t *testing.T) {
		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{
			IsAvailableResult: true,
		}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})

		// Boot -> Search
		cmd := model.Init()
		msg := cmd()
		newModel, _ := model.Update(msg)
		model = newModel.(Model)
		assert.Equal(t, searchState, model.state)

		// Search -> Loading
		loadingMsg := switchToLoadingModelMsg{queryText: "test"}
		newModel, _ = model.Update(loadingMsg)
		model = newModel.(Model)
		assert.Equal(t, loadingState, model.state)

		// Loading -> Error (API failure)
		errorMsg := switchToErrorModelMsg{err: "Network error", recoverable: true}
		newModel, _ = model.Update(errorMsg)
		model = newModel.(Model)
		assert.Equal(t, errorState, model.state)
		assert.Equal(t, "Network error", model.errorModel.message)
		assert.True(t, model.errorModel.recoverable)

		// Error -> Search (recovery)
		searchMsg := switchToSearchModelMsg{}
		newModel, _ = model.Update(searchMsg)
		model = newModel.(Model)
		assert.Equal(t, searchState, model.state)
	})

	t.Run("stations -> search -> loading -> stations workflow (new search)", func(t *testing.T) {
		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{
			IsAvailableResult: true,
		}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = stationsState

		// Return to search
		searchMsg := switchToSearchModelMsg{}
		newModel, _ := model.Update(searchMsg)
		model = newModel.(Model)
		assert.Equal(t, searchState, model.state)

		// New search
		loadingMsg := switchToLoadingModelMsg{queryText: "rock"}
		newModel, _ = model.Update(loadingMsg)
		model = newModel.(Model)
		assert.Equal(t, loadingState, model.state)
		assert.Equal(t, "rock", model.loadingModel.queryText)

		// Results loaded
		stationsMsg := switchToStationsModelMsg{}
		newModel, _ = model.Update(stationsMsg)
		model = newModel.(Model)
		assert.Equal(t, stationsState, model.state)
	})

	t.Run("terminal resize during workflow preserves state", func(t *testing.T) {
		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{
			IsAvailableResult: true,
		}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})
		model.state = stationsState

		// Resize to too small
		smallMsg := tea.WindowSizeMsg{Width: 50, Height: 20}
		newModel, _ := model.Update(smallMsg)
		model = newModel.(Model)
		assert.Equal(t, terminalTooSmallState, model.state)
		assert.Equal(t, stationsState, model.previousState)

		// Continue working during small state (should be ignored)
		loadingMsg := switchToLoadingModelMsg{queryText: "ignored"}
		newModel, _ = model.Update(loadingMsg)
		model = newModel.(Model)
		// State transition should still happen even in terminalTooSmallState
		assert.Equal(t, loadingState, model.state)

		// Resize back to adequate
		adequateMsg := tea.WindowSizeMsg{Width: 120, Height: 50}
		newModel, _ = model.Update(adequateMsg)
		model = newModel.(Model)
		// Should restore to loadingState (current state after the transition)
		assert.Equal(t, loadingState, model.state)
	})

	t.Run("fatal error prevents recovery", func(t *testing.T) {
		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{
			IsAvailableResult: false,
		}

		model := NewModel(config.Config{}, &browser, &playbackManager, &mocks.MockStationStorageService{})

		// Init fails due to no playback
		cmd := model.Init()
		msg := cmd()
		assert.IsType(t, switchToErrorModelMsg{}, msg)

		errorMsg := msg.(switchToErrorModelMsg)
		assert.False(t, errorMsg.recoverable)
	})

}
