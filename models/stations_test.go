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
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/mocks"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

var defaultStationsKeybindings = config.Keybindings{
	Quit:           "q",
	Search:         "s",
	Record:         "r",
	BookmarkToggle: "b",
	BookmarksView:  "B",
	HideStation:    "h",
	ManageHidden:   "H",
	ChangeLanguage: "L",
	VolumeDown:     "9",
	VolumeUp:       "0",
	NavigateDown:   "j",
	NavigateUp:     "k",
	StopPlayback:   "ctrl+k",
}

func createTestStation(name string) common.Station {
	return common.Station{
		StationUuid:    uuid.New(),
		Name:           name,
		CountryCode:    "US",
		LanguagesCodes: "en",
		Codec:          "mp3",
		Votes:          100,
	}
}

func createTestStationsModel(stations []common.Station, keybindings config.Keybindings) StationsModel {
	mockPM := &mocks.MockPlaybackManagerService{
		NameResult:          "ffplay",
		IsAvailableResult:   true,
		VolumeDefaultResult: 50,
		VolumeMaxResult:     100,
	}
	mockStorage := &mocks.MockStationStorageService{}

	return NewStationsModel(
		Theme{},
		nil, // browser not needed for keybinding tests
		mockPM,
		mockStorage,
		stations,
		viewModeSearchResults,
		"",
		"",
		keybindings,
	)
}

func TestStationsModel_CustomQuitKey(t *testing.T) {
	stations := []common.Station{createTestStation("Test Radio")}

	t.Run("custom quit key triggers quit", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.Quit = "x"

		model := createTestStationsModel(stations, customKb)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
		_, cmd := model.Update(input)

		assert.NotNil(t, cmd)
		// The command is a tea.Sequence, we can verify it's not nil
	})

	t.Run("default quit key does not work with custom binding", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.Quit = "x"

		model := createTestStationsModel(stations, customKb)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
		_, cmd := model.Update(input)

		// When default key is pressed with custom binding, it should not trigger quit
		// cmd could be nil or a table update command
		if cmd != nil {
			// Execute the command and check it's not a quit-related message
			msg := cmd()
			_, isQuit := msg.(quitMsg)
			assert.False(t, isQuit, "default quit key should not work when custom binding is set")
		}
	})
}

func TestStationsModel_CustomSearchKey(t *testing.T) {
	stations := []common.Station{createTestStation("Test Radio")}

	t.Run("custom search key returns to search", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.Search = "/"

		model := createTestStationsModel(stations, customKb)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
		_, cmd := model.Update(input)

		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, switchToSearchModelMsg{}, msg)
	})

	t.Run("default search key does not work with custom binding", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.Search = "/"

		model := createTestStationsModel(stations, customKb)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
		_, cmd := model.Update(input)

		// Should not trigger search switch
		if cmd != nil {
			msg := cmd()
			_, isSearch := msg.(switchToSearchModelMsg)
			assert.False(t, isSearch, "default search key should not work when custom binding is set")
		}
	})
}

func TestStationsModel_CustomVolumeKeys(t *testing.T) {
	stations := []common.Station{createTestStation("Test Radio")}

	t.Run("custom volume up key works", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.VolumeUp = "+"

		model := createTestStationsModel(stations, customKb)
		initialVolume := model.volume

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")}
		newModel, _ := model.Update(input)

		assert.Greater(t, newModel.(StationsModel).volume, initialVolume)
	})

	t.Run("custom volume down key works", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.VolumeDown = "-"

		model := createTestStationsModel(stations, customKb)
		initialVolume := model.volume

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")}
		newModel, _ := model.Update(input)

		assert.Less(t, newModel.(StationsModel).volume, initialVolume)
	})

	t.Run("default volume keys do not work with custom bindings", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.VolumeUp = "+"
		customKb.VolumeDown = "-"

		model := createTestStationsModel(stations, customKb)
		initialVolume := model.volume

		// Try default keys
		inputUp := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")}
		newModel, _ := model.Update(inputUp)
		assert.Equal(t, initialVolume, newModel.(StationsModel).volume, "default volume up should not work")

		inputDown := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9")}
		newModel, _ = model.Update(inputDown)
		assert.Equal(t, initialVolume, newModel.(StationsModel).volume, "default volume down should not work")
	})
}

func TestStationsModel_CustomNavigationKeys(t *testing.T) {
	stations := []common.Station{
		createTestStation("Radio 1"),
		createTestStation("Radio 2"),
		createTestStation("Radio 3"),
	}

	t.Run("custom navigate down key works", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.NavigateDown = "n"

		model := createTestStationsModel(stations, customKb)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
		_, cmd := model.Update(input)

		// Navigation keys trigger stationCursorMovedMsg
		assert.NotNil(t, cmd)
	})

	t.Run("custom navigate up key works", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.NavigateUp = "p"

		model := createTestStationsModel(stations, customKb)
		// Move cursor down first
		model.stationsTable.SetCursor(1)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")}
		_, cmd := model.Update(input)

		assert.NotNil(t, cmd)
	})
}

func TestStationsModel_CustomBookmarkKey(t *testing.T) {
	stations := []common.Station{createTestStation("Test Radio")}

	t.Run("custom bookmark key toggles bookmark", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.BookmarkToggle = "m"

		model := createTestStationsModel(stations, customKb)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")}
		_, cmd := model.Update(input)

		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, bookmarkToggledMsg{}, msg)
	})

	t.Run("default bookmark key does not work with custom binding", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.BookmarkToggle = "m"

		model := createTestStationsModel(stations, customKb)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")}
		_, cmd := model.Update(input)

		// Should not trigger bookmark toggle
		if cmd != nil {
			msg := cmd()
			_, isBookmark := msg.(bookmarkToggledMsg)
			assert.False(t, isBookmark, "default bookmark key should not work when custom binding is set")
		}
	})
}

func TestStationsModel_CustomHideKey(t *testing.T) {
	stations := []common.Station{createTestStation("Test Radio")}

	t.Run("custom hide key hides station", func(t *testing.T) {
		customKb := defaultStationsKeybindings
		customKb.HideStation = "x"

		model := createTestStationsModel(stations, customKb)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
		_, cmd := model.Update(input)

		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, stationHiddenMsg{}, msg)
	})
}

// executeSequenceCommands extracts and executes commands from a tea.Sequence message
// using reflection since sequenceMsg is an unexported type (it's a []Cmd slice).
func executeSequenceCommands(msg tea.Msg) {
	seqValue := reflect.ValueOf(msg)
	if seqValue.Kind() == reflect.Slice {
		for i := 0; i < seqValue.Len(); i++ {
			cmdVal := seqValue.Index(i)
			if cmd, ok := cmdVal.Interface().(tea.Cmd); ok && cmd != nil {
				cmd()
			}
		}
	}
}

func TestHideStation_StopsPlaybackWhenHidingPlayingStation(t *testing.T) {
	station := createTestStation("Test Radio")
	stations := []common.Station{station}

	t.Run("hiding playing station stops playback", func(t *testing.T) {
		stopStationCalled := false

		mockPM := &mocks.MockPlaybackManagerService{
			NameResult:          "ffplay",
			IsAvailableResult:   true,
			VolumeDefaultResult: 50,
			VolumeMaxResult:     100,
			IsPlayingResult:     true,
			IsRecordingResult:   false,
			StopStationFunc: func() error {
				stopStationCalled = true
				return nil
			},
		}
		mockStorage := &mocks.MockStationStorageService{}

		model := NewStationsModel(
			Theme{},
			nil,
			mockPM,
			mockStorage,
			stations,
			viewModeSearchResults,
			"",
			"",
			defaultStationsKeybindings,
		)
		// Set current station to simulate it's playing
		model.currentStation = station

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
		_, cmd := model.Update(input)

		assert.NotNil(t, cmd)

		// Execute the returned command to get the sequence message
		msg := cmd()
		// Execute all commands in the sequence using reflection
		executeSequenceCommands(msg)

		assert.True(t, stopStationCalled, "StopStation should be called when hiding playing station")
	})

	t.Run("hiding non-playing station does not stop playback", func(t *testing.T) {
		stopStationCalled := false

		mockPM := &mocks.MockPlaybackManagerService{
			NameResult:          "ffplay",
			IsAvailableResult:   true,
			VolumeDefaultResult: 50,
			VolumeMaxResult:     100,
			IsPlayingResult:     false,
			StopStationFunc: func() error {
				stopStationCalled = true
				return nil
			},
		}
		mockStorage := &mocks.MockStationStorageService{}

		model := NewStationsModel(
			Theme{},
			nil,
			mockPM,
			mockStorage,
			stations,
			viewModeSearchResults,
			"",
			"",
			defaultStationsKeybindings,
		)
		// currentStation is zero value (not playing)

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
		_, cmd := model.Update(input)

		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, stationHiddenMsg{}, msg)
		assert.False(t, stopStationCalled, "StopStation should not be called when hiding non-playing station")
	})
}

func TestHideStation_StopsRecordingAndPlaybackWhenHidingRecordingStation(t *testing.T) {
	station := createTestStation("Test Radio")
	stations := []common.Station{station}

	t.Run("hiding recording station stops both recording and playback", func(t *testing.T) {
		stopStationCalled := false
		stopRecordingCalled := false

		mockPM := &mocks.MockPlaybackManagerService{
			NameResult:          "ffplay",
			IsAvailableResult:   true,
			VolumeDefaultResult: 50,
			VolumeMaxResult:     100,
			IsPlayingResult:     true,
			IsRecordingResult:   true,
			StopStationFunc: func() error {
				stopStationCalled = true
				return nil
			},
			StopRecordingFunc: func() (string, error) {
				stopRecordingCalled = true
				return "/tmp/recording.mp3", nil
			},
		}
		mockStorage := &mocks.MockStationStorageService{}

		model := NewStationsModel(
			Theme{},
			nil,
			mockPM,
			mockStorage,
			stations,
			viewModeSearchResults,
			"",
			"",
			defaultStationsKeybindings,
		)
		// Set current station to simulate it's playing and recording
		model.currentStation = station

		input := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
		_, cmd := model.Update(input)

		assert.NotNil(t, cmd)

		// Execute the returned command to get the sequence message
		msg := cmd()
		// Execute all commands in the sequence using reflection
		executeSequenceCommands(msg)

		assert.True(t, stopRecordingCalled, "StopRecording should be called when hiding recording station")
		assert.True(t, stopStationCalled, "StopStation should be called when hiding recording station")
	})
}

func TestStationsModel_AllDefaultKeybindingsWork(t *testing.T) {
	stations := []common.Station{createTestStation("Test Radio")}

	t.Run("all default keybindings work correctly", func(t *testing.T) {
		model := createTestStationsModel(stations, defaultStationsKeybindings)

		// Test quit key
		_, quitCmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
		assert.NotNil(t, quitCmd, "quit key should work")

		// Test search key
		_, searchCmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
		assert.NotNil(t, searchCmd, "search key should work")
		searchMsg := searchCmd()
		assert.IsType(t, switchToSearchModelMsg{}, searchMsg)

		// Test bookmark key
		_, bookmarkCmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
		assert.NotNil(t, bookmarkCmd, "bookmark key should work")
		bookmarkMsg := bookmarkCmd()
		assert.IsType(t, bookmarkToggledMsg{}, bookmarkMsg)

		// Test hide key
		_, hideCmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
		assert.NotNil(t, hideCmd, "hide key should work")
		hideMsg := hideCmd()
		assert.IsType(t, stationHiddenMsg{}, hideMsg)

		// Test volume keys
		initialVolume := model.volume
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")})
		assert.Greater(t, newModel.(StationsModel).volume, initialVolume, "volume up should work")

		newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9")})
		assert.Less(t, newModel.(StationsModel).volume, initialVolume, "volume down should work")
	})
}
