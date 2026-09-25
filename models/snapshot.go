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
	"github.com/zi0p4tch0/radiogogo/common"
	"uuid"
)

// Snapshot is a read-only description of what the UI currently shows.
// It is intended for developer tooling that observes a running instance.
type Snapshot struct {
	// View is one of "boot", "search", "loading", "stations", "error" or
	// "terminal_too_small".
	View   string
	Width  int
	Height int
	// Stations is nil unless View is "stations".
	Stations *StationsSnapshot
}

// StationsSnapshot describes the stations view.
type StationsSnapshot struct {
	// Cursor is the zero-based index of the highlighted row.
	Cursor int
	// Selected is the station under the cursor, or nil if the list is empty.
	Selected *StationRef
	// Playing is the station being played, or nil if nothing is playing.
	Playing   *StationRef
	Volume    int
	Recording bool
	// ModalOpen reports whether the hidden-stations modal is shown.
	ModalOpen bool
}

// StationRef identifies a station.
type StationRef struct {
	Name string
	UUID string
}

// Snapshot returns the current UI state. It has no side effects.
func (m Model) Snapshot() Snapshot {
	s := Snapshot{
		View:   m.state.name(),
		Width:  m.width,
		Height: m.height,
	}
	if m.state != stationsState {
		return s
	}

	sm := m.stationsModel
	cursor := sm.stationsTable.Cursor()
	ss := &StationsSnapshot{
		Cursor:    cursor,
		Volume:    sm.volume,
		Recording: m.headerModel.isRecording,
		ModalOpen: sm.showHiddenModal,
	}
	if cursor >= 0 && cursor < len(sm.stations) {
		ss.Selected = newStationRef(sm.stations[cursor])
	}
	if sm.currentStation.StationUuid != uuid.Nil() {
		ss.Playing = newStationRef(sm.currentStation)
	}
	s.Stations = ss
	return s
}

func newStationRef(station common.Station) *StationRef {
	return &StationRef{Name: station.Name, UUID: station.StationUuid.String()}
}

func (s modelState) name() string {
	switch s {
	case bootState:
		return "boot"
	case searchState:
		return "search"
	case errorState:
		return "error"
	case loadingState:
		return "loading"
	case stationsState:
		return "stations"
	case terminalTooSmallState:
		return "terminal_too_small"
	}
	return "unknown"
}
