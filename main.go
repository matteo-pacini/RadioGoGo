package main

import (
	"fmt"
	"os"
	"radiogogo/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	model := ui.NewModel()

	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting program: %v", err)
		os.Exit(1)
	}

}
