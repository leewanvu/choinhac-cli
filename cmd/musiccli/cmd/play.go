package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"choinhaccli/internal/audio"
	"choinhaccli/internal/ui"
)

var playCmd = &cobra.Command{
	Use:   "play <path>",
	Short: "Play an audio file or directory",
	Long:  "Play a .flac or .wav file, or all supported files in a directory.",
	Args:  cobra.ExactArgs(1),
	RunE:  runPlay,
}

func runPlay(cmd *cobra.Command, args []string) error {
	path := args[0]

	if err := audio.InitSpeaker(); err != nil {
		return fmt.Errorf("failed to initialize audio speaker: %w", err)
	}

	p := audio.NewPlayer()

	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error accessing path: %w", err)
	}

	var playlist []string
	if fileInfo.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("error reading directory: %w", err)
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
			return fmt.Errorf("no supported audio files (.wav, .flac) found in directory")
		}
		sort.Strings(playlist)
	} else {
		ext := filepath.Ext(path)
		if ext != ".wav" && ext != ".flac" {
			return fmt.Errorf("unsupported file format: %s (only .wav, .flac)", ext)
		}
		playlist = []string{path}
	}

	if err := p.LoadPlaylist(playlist, 0); err != nil {
		return fmt.Errorf("error loading playlist: %w", err)
	}

	m := ui.NewModel(p)
	program := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("error starting UI: %w", err)
	}

	p.Stop()
	return nil
}
