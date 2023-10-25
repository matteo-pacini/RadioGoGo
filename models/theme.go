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

package models

import (
	"fmt"

	"github.com/zi0p4tch0/radiogogo/config"
	"github.com/zi0p4tch0/radiogogo/data"
	"github.com/zi0p4tch0/radiogogo/playback"

	"github.com/charmbracelet/lipgloss"
)

// Theme represents a style configuration for the application.
type Theme struct {
	config config.Config
}

// PrimaryColor returns the primary color of the given theme.
func (t Theme) PrimaryColor() string {
	return t.config.Theme.PrimaryColor
}

// SecondaryColor returns the secondary color of the given theme.
func (t Theme) SecondaryColor() string {
	return t.config.Theme.SecondaryColor
}

// TertiaryColor returns the tertiary color of the given theme.
func (t Theme) TertiaryColor() string {
	return t.config.Theme.TertiaryColor
}

// TextColor returns the text color for the given theme.
func (t Theme) TextColor() string {
	return t.config.Theme.TextColor
}

// ErrorColor returns the error color for the given theme.
func (t Theme) ErrorColor() string {
	return t.config.Theme.ErrorColor
}

// Header returns a string containing the styled header and version of the application.
func (t Theme) Header(playbackManager playback.PlaybackManagerService) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.TextColor())).
		Background(lipgloss.Color(t.PrimaryColor())).
		PaddingLeft(2).
		PaddingRight(2)

	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.TextColor())).
		Background(lipgloss.Color(t.SecondaryColor())).
		PaddingLeft(2).
		PaddingRight(2)

	header := headerStyle.Render("radiogogo")
	version := versionStyle.Render(fmt.Sprintf("v%s", data.Version))
	engine := headerStyle.Render("Playback engine: " + playbackManager.Name())

	return header + version + engine + "\n"
}

// StyleBottomBar returns a string representing the styled bottom bar of the given Theme.
// It takes a slice of strings representing the commands to be displayed in the bottom bar.
// The function iterates over the commands and applies a different style to each one based on its index.
// If the index is even, the command is styled with the primary color of the Theme as background.
// If the index is odd, the command is styled with the secondary color of the Theme as background.
// The styled commands are concatenated into a single string and returned.
func (t Theme) StyleBottomBar(commands []string) string {

	var bottomBar string
	for i, command := range commands {
		if i%2 == 0 {
			bottomBar += lipgloss.NewStyle().
				Foreground(lipgloss.Color(t.TextColor())).
				Background(lipgloss.Color(t.PrimaryColor())).
				PaddingLeft(2).
				PaddingRight(2).
				Render(command)
		} else {
			bottomBar += lipgloss.NewStyle().
				Foreground(lipgloss.Color(t.TextColor())).
				Background(lipgloss.Color(t.SecondaryColor())).
				PaddingLeft(2).
				PaddingRight(2).
				Render(command)
		}
	}
	return bottomBar

}

// StyleSetForegroundText returns a string with the input text styled with the text color of the Theme.
func (t Theme) StyleSetForegroundText(input string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.TextColor())).
		Render(input)
}

// StyleSetForegroundPrimary returns a string with the input text styled with the primary color of the Theme.
// If the bold parameter is true, the text is also styled as bold.
func (t Theme) StyleSetForegroundPrimary(input string, bold bool) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.PrimaryColor())).
		Bold(bold).
		Render(input)
}

// StyleSetForegroundSecondary returns a string with the input text styled with the secondary color of the Theme.
// If the bold parameter is true, the text is also styled as bold.
func (t Theme) StyleSetForegroundSecondary(input string, bold bool) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.SecondaryColor())).
		Bold(bold).
		Render(input)
}

// StyleSetForegroundTertiary returns a string with the input text styled with the tertiary color of the Theme.
func (t Theme) StyleSetForegroundTertiary(input string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.TertiaryColor())).
		Render(input)
}

// StyleSetForegroundError returns a string with the input text styled with the error color of the Theme.
func (t Theme) StyleSetForegroundError(input string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.ErrorColor())).
		Bold(true).
		Render(input)
}
