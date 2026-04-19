package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "musiccli",
	Short: "A high-fidelity CLI music player",
	Long:  "A minimalist CLI music player supporting FLAC and WAV with a TUI and AI music appreciation.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(playCmd)
	rootCmd.AddCommand(feelCmd)
	rootCmd.AddCommand(serveCmd)
}
