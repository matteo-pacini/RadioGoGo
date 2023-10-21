package ui

import (
	"fmt"
	"radiogogo/data"

	"github.com/charmbracelet/lipgloss"
)

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
