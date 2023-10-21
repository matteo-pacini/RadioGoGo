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

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SearchModel struct {
	inputModel textinput.Model
	width      int
	height     int
}

func NewSearchModel() SearchModel {
	i := textinput.New()
	i.Placeholder = "Name"
	i.Width = 30
	i.Focus()

	return SearchModel{
		inputModel: i,
	}

}

func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, func() tea.Msg {
		return bottomBarUpdateMsg{
			commands: []string{"q: quit", "enter: search"},
		}
	})
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, quitCmd
		case "enter":
			return m, func() tea.Msg {
				return switchToLoadingModelMsg{query: m.inputModel.Value()}
			}
		}
	}

	newInputModel, cmd := m.inputModel.Update(msg)
	m.inputModel = newInputModel

	return m, cmd
}

func (m SearchModel) View() string {

	v := fmt.Sprintf(
		"\nSearch a radio by name\n\n%s",
		m.inputModel.View(),
	)

	return v
}
