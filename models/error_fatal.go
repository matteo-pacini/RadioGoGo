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
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	// How many ticks (one tick per second) to wait before quitting
	quitTicks = 30
)

// Messages

type quitTickMsg struct{}

// Model

type ErrorModel struct {
	theme Theme

	message     string
	recoverable bool

	tickCount int
	width     int
	height    int
}

func NewErrorModel(theme Theme, err string, recoverable bool) ErrorModel {

	return ErrorModel{
		theme:       theme,
		message:     err,
		recoverable: recoverable,
	}

}

func (m ErrorModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return quitTickMsg{}
	})
}

func (m ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, quitCmd
		case "enter", "esc":
			if m.recoverable {
				return m, func() tea.Msg { return switchToSearchModelMsg{} }
			}
		}
	case quitTickMsg:
		if m.recoverable {
			// Don't auto-quit for recoverable errors
			return m, nil
		}
		m.tickCount++
		if m.tickCount >= quitTicks {
			return m, quitCmd
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return quitTickMsg{}
		})
	}

	return m, nil

}

func (m ErrorModel) View() string {

	var message string
	if m.recoverable {
		message = fmt.Sprintf("%s\n\nPress Enter to try again or \"q\" to quit.", m.message)
	} else {
		message = fmt.Sprintf("%s\n\nQuitting in %d seconds (or press \"q\" to exit now)...", m.message, quitTicks-m.tickCount)
	}

	return "\n" + m.theme.ErrorText.Render(message) + "\n\n"

}

func (m *ErrorModel) SetWidthAndHeight(width int, height int) {
	m.width = width
	m.height = height
}
