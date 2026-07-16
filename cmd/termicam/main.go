package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/Megge06/TermiCam/internal/tui"
)

// Main entry point
func main() {
	p := tea.NewProgram(tui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
