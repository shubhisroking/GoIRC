package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	logFile, err := tea.LogToFile("irc_debug.log", "debug")
	if err != nil {
		fmt.Println("could not create log file:", err)
	}
	defer logFile.Close()

	log.Println("Starting Bubble Tea IRC Client...")

	prog := tea.NewProgram(initialModel(), tea.WithAltScreen())
	p = prog

	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
	}
	log.Println("Bubble Tea IRC Client stopped.")
}
