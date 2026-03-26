package ui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF5F87")).
		MarginBottom(1)

	metadataStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E1ED")).
		MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Width(12)

	valueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EEEEEE")).
		Bold(true)

	statsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		MarginBottom(1)

	progressBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F1F1")).
		Background(lipgloss.Color("#2A2A2D"))

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		MarginTop(1)

	playlistTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginTop(1).
		MarginBottom(0)

	playlistItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0"))

	currentTrackStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)
)
