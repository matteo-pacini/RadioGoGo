package models

import (
	"radiogogo/mocks"
	"testing"

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

		model := NewModel(&browser, &playbackManager)

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

		model := NewModel(&browser, &playbackManager)

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

		model := NewModel(&browser, &playbackManager)

		msg := tea.WindowSizeMsg{Width: 100, Height: 100}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, 100, newModel.(Model).width)
		assert.Equal(t, 100, newModel.(Model).height)
		assert.Nil(t, cmd)

	})

	t.Run("propagates adjusted terminal size changes to SearchModel if active", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(&browser, &playbackManager)
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

		model := NewModel(&browser, &playbackManager)
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

		model := NewModel(&browser, &playbackManager)
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

		model := NewModel(&browser, &playbackManager)
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

		model := NewModel(&browser, &playbackManager)

		msg := quitMsg{}

		_, cmd := model.Update(tea.Msg(msg))

		returnedMsg := cmd()

		assert.IsType(t, tea.QuitMsg{}, returnedMsg)

	})

	t.Run("stores bottom bar commands and returns a nil command", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(&browser, &playbackManager)

		msg := bottomBarUpdateMsg{commands: []string{"test"}}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, []string{"test"}, newModel.(Model).bottomBarCommands)
		assert.Nil(t, cmd)

	})

	t.Run("recreates and switches to search model if switchToSearchModelMsg is received", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(&browser, &playbackManager)
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

		model := NewModel(&browser, &playbackManager)
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

		model := NewModel(&browser, &playbackManager)
		model.stationsModel.volume = defaultVolume + 1

		msg := switchToStationsModelMsg{}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, stationsState, newModel.(Model).state)
		assert.Equal(t, defaultVolume, newModel.(Model).stationsModel.volume)
		assert.NotNil(t, cmd)

	})

	t.Run("recreates and switches to error model if switchToErrorModelMsg is received", func(t *testing.T) {

		browser := mocks.MockRadioBrowserService{}
		playbackManager := mocks.MockPlaybackManagerService{}

		model := NewModel(&browser, &playbackManager)
		model.errorModel.message = "test"

		msg := switchToErrorModelMsg{err: "test2"}

		newModel, cmd := model.Update(tea.Msg(msg))

		assert.Equal(t, errorState, newModel.(Model).state)
		assert.Equal(t, "test2", newModel.(Model).errorModel.message)
		assert.NotNil(t, cmd)

	})

}
