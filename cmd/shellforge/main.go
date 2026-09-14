package main

import (
	"fmt"
	"os"
	"shellforge/internal/tui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	program := tea.NewProgram(tui.New())
	finalModel, err := program.Run()
	if model, ok := finalModel.(tui.Model); ok {
		model.Close()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Shellforge could not start:", err)
		os.Exit(1)
	}
}
