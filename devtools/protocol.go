//go:build devtools

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

// Package devtools lets external tools (typically an LLM assistant) attach to a
// running RadioGoGo over a Unix socket to read the screen and state, inject
// keypresses and observe message flow. It is compiled only with the devtools
// build tag and never ships in release binaries.
package devtools

import (
	"time"

	"github.com/zi0p4tch0/radiogogo/models"
)

// Request is one line sent by a client.
type Request struct {
	Op        string   `json:"op"`
	Keys      []string `json:"keys,omitempty"`
	Text      string   `json:"text,omitempty"`
	TimeoutMS int      `json:"timeout_ms,omitempty"`
}

// Response is the single line written back for every request except watch.
type Response struct {
	OK     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
	Screen string `json:"screen,omitempty"`
	State  *State `json:"state,omitempty"`
}

// State is the structured description of the UI returned by the state op.
type State struct {
	View     string         `json:"view"`
	Width    int            `json:"width"`
	Height   int            `json:"height"`
	Stations *StationsState `json:"stations,omitempty"`
}

// StationsState is present only when View is "stations".
type StationsState struct {
	Cursor    int         `json:"cursor"`
	Selected  *StationRef `json:"selected"`
	Playing   *StationRef `json:"playing"`
	Volume    int         `json:"volume"`
	Recording bool        `json:"recording"`
	ModalOpen bool        `json:"modal_open"`
}

// StationRef identifies a station.
type StationRef struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

// Event is one line of a watch stream. Type is "msg" for a processed message
// (Name is its Go type) or "view" for a view change (From and To are view names).
// Dropped counts events discarded before this one because the client was slow.
type Event struct {
	Type    string    `json:"type"`
	Name    string    `json:"name,omitempty"`
	From    string    `json:"from,omitempty"`
	To      string    `json:"to,omitempty"`
	TS      time.Time `json:"ts"`
	Dropped int       `json:"dropped,omitempty"`
}

func newState(s models.Snapshot) *State {
	state := &State{View: s.View, Width: s.Width, Height: s.Height}
	if st := s.Stations; st != nil {
		state.Stations = &StationsState{
			Cursor:    st.Cursor,
			Selected:  newStationRef(st.Selected),
			Playing:   newStationRef(st.Playing),
			Volume:    st.Volume,
			Recording: st.Recording,
			ModalOpen: st.ModalOpen,
		}
	}
	return state
}

func newStationRef(r *models.StationRef) *StationRef {
	if r == nil {
		return nil
	}
	return &StationRef{Name: r.Name, UUID: r.UUID}
}
