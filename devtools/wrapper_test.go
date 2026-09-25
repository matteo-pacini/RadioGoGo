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

package devtools

import (
	"testing"

	"github.com/zi0p4tch0/radiogogo/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setViewMsg switches the fake model's reported view.
type setViewMsg string

// fakeModel echoes typed keys into its view and counts View calls.
type fakeModel struct {
	typed     string
	view      string
	received  *[]tea.Msg
	viewCalls *int
}

func newFakeModel() fakeModel {
	return fakeModel{view: "search", received: &[]tea.Msg{}, viewCalls: new(int)}
}

func (f fakeModel) Init() tea.Cmd { return nil }

func (f fakeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	*f.received = append(*f.received, msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		f.typed += msg.String()
	case setViewMsg:
		f.view = string(msg)
	}
	return f, nil
}

func (f fakeModel) View() string {
	*f.viewCalls++
	return "\x1b[1mtyped:" + f.typed + "\x1b[0m"
}

func (f fakeModel) Snapshot() models.Snapshot {
	return models.Snapshot{View: f.view, Width: 120, Height: 40}
}

func update(t *testing.T, m tea.Model, msg tea.Msg) tea.Model {
	t.Helper()
	next, _ := m.Update(msg)
	return next
}

func TestWrapper_ForwardsOrdinaryMessages(t *testing.T) {
	fake := newFakeModel()
	m := Wrap(fake)

	m = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	require.Len(t, *fake.received, 1)
	assert.Equal(t, "typed:a", m.(*remoteModel).screen())
}

func TestWrapper_ControlMessagesNotForwarded(t *testing.T) {
	fake := newFakeModel()
	m := Wrap(fake)

	reply := make(chan snapshotResult, 1)
	m = update(t, m, snapshotReq{reply: reply})
	m = update(t, m, waitReq{id: 1, text: "never", reply: make(chan waitResult, 1)})
	m = update(t, m, waitCancel{id: 1})
	m = update(t, m, watchReq{id: 2, ch: make(chan Event, 1)})
	update(t, m, watchDrop{id: 2})

	assert.Empty(t, *fake.received)
	res := <-reply
	assert.Equal(t, "typed:", res.screen, "ANSI must be stripped")
	assert.Equal(t, "search", res.state.View)
}

func TestWrapper_Wait(t *testing.T) {
	t.Run("matches immediately", func(t *testing.T) {
		reply := make(chan waitResult, 1)
		update(t, Wrap(newFakeModel()), waitReq{id: 1, text: "typed:", reply: reply})
		res := <-reply
		assert.True(t, res.matched)
		assert.Equal(t, "typed:", res.screen)
	})

	t.Run("matches after a later message", func(t *testing.T) {
		reply := make(chan waitResult, 1)
		m := update(t, Wrap(newFakeModel()), waitReq{id: 1, text: "typed:x", reply: reply})
		assert.Empty(t, reply)

		update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

		res := <-reply
		assert.True(t, res.matched)
		assert.Empty(t, m.(*remoteModel).waits)
	})

	t.Run("cancel returns last screen", func(t *testing.T) {
		reply := make(chan waitResult, 1)
		m := update(t, Wrap(newFakeModel()), waitReq{id: 1, text: "never", reply: reply})
		m = update(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
		update(t, m, waitCancel{id: 1})

		res := <-reply
		assert.False(t, res.matched)
		assert.Equal(t, "typed:y", res.screen)
	})

	t.Run("cancel after match sends nothing more", func(t *testing.T) {
		reply := make(chan waitResult, 1)
		m := update(t, Wrap(newFakeModel()), waitReq{id: 1, text: "typed:", reply: reply})
		update(t, m, waitCancel{id: 1})
		assert.Len(t, reply, 1)
	})
}

func TestWrapper_Watch(t *testing.T) {
	t.Run("emits message and view events", func(t *testing.T) {
		ch := make(chan Event, 10)
		m := update(t, Wrap(newFakeModel()), watchReq{id: 1, ch: ch})

		m = update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
		m = update(t, m, setViewMsg("loading"))

		e := <-ch
		assert.Equal(t, Event{Type: "msg", Name: "tea.KeyMsg", TS: e.TS}, e)
		e = <-ch
		assert.Equal(t, "devtools.setViewMsg", e.Name)
		e = <-ch
		assert.Equal(t, Event{Type: "view", From: "search", To: "loading", TS: e.TS}, e)
		assert.False(t, e.TS.IsZero())

		update(t, m, watchDrop{id: 1})
		update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
		assert.Empty(t, ch)
	})

	t.Run("full channel drops without blocking and reports count", func(t *testing.T) {
		ch := make(chan Event, 1)
		m := update(t, Wrap(newFakeModel()), watchReq{id: 1, ch: ch})

		for range 4 {
			m = update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
		}
		<-ch
		update(t, m, tea.KeyMsg{Type: tea.KeyEnter})

		e := <-ch
		assert.Equal(t, 3, e.Dropped)
	})
}

func TestWrapper_RendersOnlyWhenObserved(t *testing.T) {
	fake := newFakeModel()
	m := Wrap(fake)

	m = update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, 0, *fake.viewCalls)

	m = update(t, m, waitReq{id: 1, text: "never", reply: make(chan waitResult, 1)})
	calls := *fake.viewCalls
	update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, calls+1, *fake.viewCalls)
}
