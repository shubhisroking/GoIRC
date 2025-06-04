package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "1.0.0"

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		fmt.Println("Using default configuration...")
		config = DefaultConfig()
	}

	// Initialize logger
	logger, err := NewLogger(config)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Log startup
	logger.LogStartup(version)

	logger.Log("Starting GoIRC Client...")

	// Create Bubble Tea program with appropriate options
	var progOptions []tea.ProgramOption
	progOptions = append(progOptions, tea.WithAltScreen())

	// Only enable Bubble Tea logging if debug mode is explicitly enabled
	if config.Logging.Enabled && config.Logging.DebugMode {
		if logFile, err := tea.LogToFile(config.GetDebugLogFilePath(), "debug"); err != nil {
			logger.LogError("could not create debug log file: %v", err)
		} else {
			defer logFile.Close()
		}
	}

	prog := tea.NewProgram(initialModel(), progOptions...)
	p = prog

	if _, err := p.Run(); err != nil {
		logger.LogError("Error running program: %v", err)
		log.Fatal("Error running program:", err)
	}

	logger.LogShutdown()
	logger.Log("GoIRC Client stopped.")
}
