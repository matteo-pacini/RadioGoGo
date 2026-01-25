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
	"github.com/zi0p4tch0/radiogogo/config"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a style configuration for the application.
type Theme struct {
	PrimaryBlock   lipgloss.Style
	SecondaryBlock lipgloss.Style

	Text          lipgloss.Style
	PrimaryText   lipgloss.Style
	SecondaryText lipgloss.Style
	TertiaryText  lipgloss.Style
	ErrorText     lipgloss.Style
	SuccessText   lipgloss.Style

	StationsTableStyle table.Styles
	ModalStyle         lipgloss.Style

	QualityHighStyle   lipgloss.Style
	QualityMediumStyle lipgloss.Style
	QualityLowStyle    lipgloss.Style

	StatusBoxStyle lipgloss.Style

	// Color values for dynamic styling
	SecondaryColor string
}

func NewTheme(config config.Config) Theme {

	primaryBlock := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.TextColor)).
		Background(lipgloss.Color(config.Theme.PrimaryColor)).
		PaddingLeft(2).
		PaddingRight(2)

	secondaryBlock := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.TextColor)).
		Background(lipgloss.Color(config.Theme.SecondaryColor)).
		PaddingLeft(2).
		PaddingRight(2)

	text := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.TextColor))

	primaryText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.PrimaryColor))

	secondaryText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.SecondaryColor))

	tertiaryText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.TertiaryColor))

	errorText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.ErrorColor))

	successText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B"))

	stationsTableStyles := table.DefaultStyles()
	stationsTableStyles.Header = stationsTableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(config.Theme.TextColor)).
		BorderBottom(true).
		Bold(true)
	stationsTableStyles.Cell = stationsTableStyles.Cell.
		Foreground(lipgloss.Color(config.Theme.TextColor)).
		Padding(0, 1)
	stationsTableStyles.Selected = stationsTableStyles.Selected.
		Foreground(lipgloss.Color(config.Theme.TextColor)).
		Background(lipgloss.Color(config.Theme.PrimaryColor)).
		Bold(false)

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.Theme.SecondaryColor)).
		Padding(1, 2).
		Width(50)

	qualityHighStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Bold(true)

	qualityMediumStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB86C"))

	qualityLowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.TertiaryColor))

	statusBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.Theme.PrimaryColor)).
		Padding(0, 1)

	return Theme{
		PrimaryBlock:       primaryBlock,
		SecondaryBlock:     secondaryBlock,
		Text:               text,
		PrimaryText:        primaryText,
		SecondaryText:      secondaryText,
		TertiaryText:       tertiaryText,
		ErrorText:          errorText,
		SuccessText:        successText,
		StationsTableStyle: stationsTableStyles,
		ModalStyle:         modalStyle,
		QualityHighStyle:   qualityHighStyle,
		QualityMediumStyle: qualityMediumStyle,
		QualityLowStyle:    qualityLowStyle,
		StatusBoxStyle:     statusBoxStyle,
		SecondaryColor:     config.Theme.SecondaryColor,
	}
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
			bottomBar += t.PrimaryBlock.Render(command)
		} else {
			bottomBar += t.SecondaryBlock.Render(command)
		}
	}
	return bottomBar

}

// StyleTwoRowBottomBar returns a two-row bottom bar with primary commands on top
// and secondary commands below.
func (t Theme) StyleTwoRowBottomBar(primary, secondary []string) string {
	row1 := t.StyleBottomBar(primary)
	row2 := t.StyleBottomBar(secondary)
	return row1 + "\n" + row2
}
