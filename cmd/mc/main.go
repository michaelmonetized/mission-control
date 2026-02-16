package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/michaelmonetized/mission-control/pkg/ui"
)

func main() {
	// Check for subcommands first (fall back to shell scripts)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "tui", "ui", "":
			// Continue to TUI
		default:
			// Delegate to shell scripts
			fmt.Println("Use shell scripts for CLI commands: mc-discover, mc-git-status, etc.")
			fmt.Println("Or run without args for TUI: mc")
			os.Exit(0)
		}
	}

	// Start TUI
	p := tea.NewProgram(
		ui.NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
