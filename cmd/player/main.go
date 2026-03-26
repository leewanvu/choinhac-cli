package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"choinhaccli/internal/audio"
	"choinhaccli/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: player <path_to_audio_file_or_directory>")
		fmt.Println("Example: player track.flac")
		fmt.Println("Example: player ./music_folder")
		os.Exit(1)
	}

	path := os.Args[1]

	// Init audio subsystem
	if err := audio.InitSpeaker(); err != nil {
		fmt.Printf("Failed to initialize audio speaker: %v\n", err)
		os.Exit(1)
	}

	p := audio.NewPlayer()

	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error accessing path: %v\n", err)
		os.Exit(1)
	}

	var playlist []string
	if fileInfo.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			fmt.Printf("Error reading directory: %v\n", err)
			os.Exit(1)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				ext := filepath.Ext(entry.Name())
				if ext == ".wav" || ext == ".flac" {
					playlist = append(playlist, filepath.Join(path, entry.Name()))
				}
			}
		}
		if len(playlist) == 0 {
			fmt.Println("No supported audio files (.wav, .flac) found in directory.")
			os.Exit(1)
		}
		sort.Strings(playlist)
	} else {
		ext := filepath.Ext(path)
		if ext != ".wav" && ext != ".flac" {
			fmt.Printf("Unsupported file format: %s. Please provide a .wav or .flac file.\n", ext)
			os.Exit(1)
		}
		playlist = []string{path}
	}

	// Load the audio file(s) and start playback
	if err := p.LoadPlaylist(playlist, 0); err != nil {
		fmt.Printf("Error playing path: %v\n", err)
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
