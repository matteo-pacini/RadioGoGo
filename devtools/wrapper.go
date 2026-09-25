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
	"fmt"
	"strings"
	"time"

	"github.com/zi0p4tch0/radiogogo/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// Inspectable is a tea.Model that can describe its state.
type Inspectable interface {
	tea.Model
	Snapshot() models.Snapshot
}

// Control messages. They are handled by the wrapper and never reach the inner
// model. Every reply channel must be buffered so Update never blocks.
type (
	snapshotReq struct{ reply chan snapshotResult }
	waitReq     struct {
		id    int64
		text  string
		reply chan waitResult
	}
	waitCancel struct{ id int64 }
	watchReq   struct {
		id int64
		ch chan Event
	}
	watchDrop struct{ id int64 }
)

type snapshotResult struct {
	screen string
	state  models.Snapshot
}

type waitResult struct {
	matched bool
	screen  string
}

type watcher struct {
	ch      chan Event
	dropped int
}

// remoteModel wraps the application model. All of its fields are touched only
// from Update, which BubbleTea runs on a single goroutine.
type remoteModel struct {
	inner    Inspectable
	waits    map[int64]waitReq
	watchers map[int64]*watcher
}

// Wrap returns a model that behaves like inner and additionally answers the
// control messages sent by a Server.
func Wrap(inner Inspectable) tea.Model {
	return &remoteModel{
		inner:    inner,
		waits:    map[int64]waitReq{},
		watchers: map[int64]*watcher{},
	}
}

func (m *remoteModel) Init() tea.Cmd {
	return m.inner.Init()
}

func (m *remoteModel) View() string {
	return m.inner.View()
}

func (m *remoteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case snapshotReq:
		msg.reply <- snapshotResult{screen: m.screen(), state: m.inner.Snapshot()}
		return m, nil
	case waitReq:
		if screen := m.screen(); strings.Contains(screen, msg.text) {
			msg.reply <- waitResult{matched: true, screen: screen}
		} else {
			m.waits[msg.id] = msg
		}
		return m, nil
	case waitCancel:
		if w, ok := m.waits[msg.id]; ok {
			delete(m.waits, msg.id)
			w.reply <- waitResult{screen: m.screen()}
		}
		return m, nil
	case watchReq:
		m.watchers[msg.id] = &watcher{ch: msg.ch}
		return m, nil
	case watchDrop:
		delete(m.watchers, msg.id)
		return m, nil
	}

	watching := len(m.watchers) > 0
	var viewBefore string
	if watching {
		viewBefore = m.inner.Snapshot().View
	}

	next, cmd := m.inner.Update(msg)
	m.inner = next.(Inspectable)

	if watching {
		m.emit(Event{Type: "msg", Name: fmt.Sprintf("%T", msg)})
		if viewAfter := m.inner.Snapshot().View; viewAfter != viewBefore {
			m.emit(Event{Type: "view", From: viewBefore, To: viewAfter})
		}
	}

	// The screen can only change after a message is processed, so checking
	// here is enough to resolve every wait without polling.
	if len(m.waits) > 0 {
		screen := m.screen()
		for id, w := range m.waits {
			if strings.Contains(screen, w.text) {
				delete(m.waits, id)
				w.reply <- waitResult{matched: true, screen: screen}
			}
		}
	}

	return m, cmd
}

func (m *remoteModel) screen() string {
	return ansi.Strip(m.inner.View())
}

// emit delivers e to every watcher without blocking; a full channel drops the
// event and the count is reported on the next event that gets through.
func (m *remoteModel) emit(e Event) {
	e.TS = time.Now()
	for _, w := range m.watchers {
		e.Dropped = w.dropped
		select {
		case w.ch <- e:
			w.dropped = 0
		default:
			w.dropped++
		}
	}
}
