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

package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	quitTicks = 30
)

type ErrorModel struct {
	message string

	tickCount int
}

func NewErrorModel(err string) ErrorModel {

	return ErrorModel{
		message: err,
	}

}

func (m ErrorModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return quitTickMsg{}
	})
}

type quitTickMsg struct{}

func (m ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, radiogogoQuit
		}
	case quitTickMsg:
		m.tickCount++
		if m.tickCount >= quitTicks {
			return m, radiogogoQuit
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return quitTickMsg{}
		})
	}

	return m, nil

}

func (m ErrorModel) View() string {

	message := fmt.Sprintf("%s\n\nQuitting in %d seconds (or press \"q\" to exit now)...", m.message, quitTicks-m.tickCount)

	errorRedStyle :=
		lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

	return "\n" + errorRedStyle.Render(message) + "\n\n"

}
