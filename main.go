package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize audio (non-fatal — game works silently if audio fails)
	initAudio()

	// WithAltScreen() takes over the full terminal (like vim does).
	// When the program exits, the terminal restores to its previous state.
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
