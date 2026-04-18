package ui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF79C6")).
			MarginBottom(1)

	artPanelStyle = lipgloss.NewStyle().
			Width(24).
			PaddingRight(2)

	infoPanelStyle = lipgloss.NewStyle().
			PaddingLeft(1)

	artistStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CBA6F7")).
			Bold(true)

	albumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89B4FA")).
			Italic(true)

	trackStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Bold(true).
			MarginBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")).
			Italic(true).
			MarginTop(1)

	progressBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CBA6F7")).
				Background(lipgloss.Color("#313244"))

	vizStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89DCEB")).
			Bold(true)

	vizSectionStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585B70")).
			Italic(true).
			MarginTop(1)

	playlistTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#CBA6F7")).
				MarginBottom(1)

	playlistNumStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#585B70")).
				Width(3)

	playlistItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9399B2"))

	currentNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")).
			Bold(true).
			Width(3)

	currentTrackStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A6E3A1")).
				Bold(true)
)
