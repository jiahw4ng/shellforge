package main

import (
	"fmt"
	"os"
	"shellforge/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	program := tea.NewProgram(tui.New(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Shellforge could not start:", err)
		os.Exit(1)
	}
}
