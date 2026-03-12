package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"choinhaccli/internal/audio"
	"choinhaccli/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: player <path_to_audio_file>")
		fmt.Println("Example: player track.flac")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Init audio subsystem
	if err := audio.InitSpeaker(); err != nil {
		fmt.Printf("Failed to initialize audio speaker: %v\n", err)
		os.Exit(1)
	}

	p := audio.NewPlayer()

	// Load the audio file and extract metadata
	if err := p.LoadAndPlay(filename); err != nil {
		fmt.Printf("Error playing file: %v\n", err)
		os.Exit(1)
	}

	// Initialize the BubbleTea UI Model
	m := ui.NewModel(p)
	
	// Create the program and run it using the alternate screen buffer
	program := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error starting UI: %v\n", err)
		os.Exit(1)
	}
	
	// Ensure player stops properly upon exit
	p.Stop()
}
