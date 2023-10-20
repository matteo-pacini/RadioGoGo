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
	"radiogogo/data"

	"github.com/charmbracelet/lipgloss"
)

func StyleBottomBar(commands []string) string {

	var bottomBar string
	for i, command := range commands {
		if i%2 == 0 {
			bottomBar += lipgloss.NewStyle().
				Foreground(lipgloss.Color("white")).
				Background(lipgloss.Color("#5a4f9f")).
				PaddingLeft(2).
				PaddingRight(2).
				Render(command)
		} else {
			bottomBar += lipgloss.NewStyle().
				Foreground(lipgloss.Color("white")).
				Background(lipgloss.Color("#8b77db")).
				PaddingLeft(2).
				PaddingRight(2).
				Render(command)
		}
	}
	return bottomBar

}

func StyleSetPlaying(input string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8b77db")).
		Bold(true).
		Render(input)
}

func StyleSetError(input string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("red")).
		Render(input)
}

func Header() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("white")).
		Background(lipgloss.Color("#5a4f9f")).
		PaddingLeft(2).
		PaddingRight(2)

	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("white")).
		Background(lipgloss.Color("#8b77db")).
		PaddingLeft(2).
		PaddingRight(2)

	header := headerStyle.Render("radiogogo")
	version := versionStyle.Render(fmt.Sprintf("v%s", data.Version))

	return header + version + "\n"
}
