package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"choinhaccli/internal/audio"
)

// View renders the full TUI.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\nPress q to quit.", m.err)
	}

	title := titleStyle.Render("♪  CLI Music Player")
	topPanel := renderTopPanel(m)
	vizRow := vizSectionStyle.Render(m.viz.render())
	playlist := renderPlaylist(m)
	help := helpStyle.Render("space: play/pause • n/→: next • p/←: prev • r: random • ↑/+: vol up • ↓/-: vol down • q: quit")

	return appStyle.Render(
		title + "\n\n" +
			topPanel + "\n" +
			vizRow + "\n\n" +
			playlist + "\n\n" +
			help,
	)
}

// renderTopPanel builds the 2-column art + info section.
func renderTopPanel(m Model) string {
	meta := m.player.Metadata

	playlistInfo := ""
	if len(m.player.Playlist) > 0 {
		playlistInfo = fmt.Sprintf("  [%d/%d]", m.player.PlaylistIdx+1, len(m.player.Playlist))
	}

	pos := m.player.GetPosition()
	dur := meta.Duration
	timeStr := fmt.Sprintf("%s / %s", formatDuration(pos), formatDuration(dur))
	bar := gradientProgressBar(32, pos, dur)

	infoCol := lipgloss.JoinVertical(lipgloss.Left,
		artistStyle.Render(meta.Artist),
		albumStyle.Render(meta.Album),
		trackStyle.Render(meta.Title+playlistInfo),
		"",
		bar+" "+timeStr,
		statusStyle.Render(statusText(m.player.GetState())+" "+volText(m.player.GetVolume())),
	)

	artPanel := artPanelStyle.Render(m.art)
	infoPanel := infoPanelStyle.Render(infoCol)

	return lipgloss.JoinHorizontal(lipgloss.Top, artPanel, infoPanel)
}

// renderPlaylist builds the scrollable playlist section.
func renderPlaylist(m Model) string {
	if len(m.player.Playlist) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, playlistTitleStyle.Render("Playlist"))

	startIdx := max(m.player.PlaylistIdx-3, 0)
	endIdx := min(startIdx+7, len(m.player.Playlist))

	if startIdx > 0 {
		lines = append(lines, playlistItemStyle.Render("  ···"))
	}

	for i := startIdx; i < endIdx; i++ {
		name := trackDisplayName(m.player.Playlist[i])
		num := fmt.Sprintf("%02d", i+1)
		prefix := "  "
		style := playlistItemStyle
		numStyle := playlistNumStyle

		if i == m.player.PlaylistIdx {
			prefix = "▶ "
			style = currentTrackStyle
			numStyle = currentNumStyle
		}

		line := numStyle.Render(num) + " " + style.Render(prefix+name)
		lines = append(lines, line)
	}

	if endIdx < len(m.player.Playlist) {
		lines = append(lines, playlistItemStyle.Render("  ···"))
	}

	return strings.Join(lines, "\n")
}

// gradientProgressBar renders a progress bar with sub-block precision.
func gradientProgressBar(width int, current, total time.Duration) string {
	subBlocks := []rune{'▏', '▎', '▍', '▌', '▋', '▊', '▉', '█'}
	if total <= 0 {
		return progressBarStyle.Render(strings.Repeat("░", width))
	}
	percent := float64(current) / float64(total)
	if percent > 1.0 {
		percent = 1.0
	}

	exactFilled := float64(width) * percent
	filled := int(exactFilled)
	frac := exactFilled - float64(filled)
	empty := width - filled - 1

	bar := strings.Repeat("█", filled)
	if empty >= 0 {
		subIdx := int(frac * float64(len(subBlocks)))
		if subIdx >= len(subBlocks) {
			subIdx = len(subBlocks) - 1
		}
		bar += string(subBlocks[subIdx])
		bar += strings.Repeat("░", empty)
	}

	return progressBarStyle.Render(bar)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	min := d / time.Minute
	sec := (d % time.Minute) / time.Second
	return fmt.Sprintf("%02d:%02d", min, sec)
}

func statusText(s audio.State) string {
	switch s {
	case audio.StatePlaying:
		return "▶ Playing"
	case audio.StatePaused:
		return "⏸ Paused"
	default:
		return "⏹ Stopped"
	}
}

func volText(vol float64) string {
	if vol >= 2.0 {
		return "🔊 Max"
	} else if vol > 0 {
		return fmt.Sprintf("🔊 +%.1f", vol)
	} else if vol > -2.0 {
		return fmt.Sprintf("🔉 %.1f", vol)
	}
	return fmt.Sprintf("🔈 %.1f", vol)
}

func trackDisplayName(path string) string {
	name := filepath.Base(path)
	return strings.TrimSuffix(name, filepath.Ext(name))
}
