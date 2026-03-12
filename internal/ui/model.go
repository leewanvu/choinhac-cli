package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"choinhaccli/internal/audio"
)

type tickMsg time.Time

// Model represents the applicaton state
type Model struct {
	player *audio.Player
	width  int
	err    error
}

// NewModel creates a new UI model
func NewModel(p *audio.Player) Model {
	return Model{
		player: p,
	}
}

// Init initializes the tea application
func (m Model) Init() tea.Cmd {
	return m.tickCmd()
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles incoming messages and events
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.player.Stop()
			return m, tea.Quit
		case " ":
			m.player.TogglePause()
		case "=", "+", "up":
			m.player.VolumeUp()
		case "-", "down":
			m.player.VolumeDown()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tickMsg:
		// When track finishes, audio player might set state to StateStopped.
		// Re-trigger tick.
		return m, m.tickCmd()

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// formatDuration helper
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	min := d / time.Minute
	sec := (d % time.Minute) / time.Second
	return fmt.Sprintf("%02d:%02d", min, sec)
}

func progressBar(width int, current, total time.Duration) string {
	if total <= 0 {
		bar := strings.Repeat("░", width)
		return progressBarStyle.Render(bar)
	}
	percent := float64(current) / float64(total)
	if percent > 1.0 {
		percent = 1.0
	}

	filled := int(float64(width) * percent)
	empty := width - filled

	if filled < 0 {
		filled = 0
	}
	if empty < 0 {
		empty = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return progressBarStyle.Render(bar)
}

// View Renders the TUI
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\nPress q to quit.", m.err)
	}

	meta := m.player.Metadata
	status := "Playing"
	if m.player.GetState() == audio.StatePaused {
		status = "Paused "
	} else if m.player.GetState() == audio.StateStopped {
		status = "Stopped"
	}

	// Title
	title := titleStyle.Render(fmt.Sprintf("\n🎵  %s", "CLI Music Player"))

	// Metadata
	artistRow := labelStyle.Render("Artist:") + valueStyle.Render(meta.Artist)
	albumRow := labelStyle.Render("Album:") + valueStyle.Render(meta.Album)
	trackRow := labelStyle.Render("Track:") + valueStyle.Render(meta.Title)

	metadata := metadataStyle.Render(
		artistRow + "\n" +
			albumRow + "\n" +
			trackRow,
	)

	// Stats
	volInfo := ""
	vol := m.player.GetVolume()
	if vol == 0 {
		volInfo = "(Vol: Normal)"
	} else if vol > 0 {
		volInfo = fmt.Sprintf("(Vol: +%.1f)", vol)
	} else {
		volInfo = fmt.Sprintf("(Vol: %.1f)", vol)
	}

	statsText := fmt.Sprintf("Sample Rate: %d Hz | Status: %s %s", meta.SampleRate, status, volInfo)
	stats := statsStyle.Render(statsText)

	// Progress
	pos := m.player.GetPosition()
	dur := meta.Duration

	timeStr := fmt.Sprintf("%s / %s", formatDuration(pos), formatDuration(dur))

	barWidth := m.width - 4 - len(timeStr) - 5 // a bit of padding
	if barWidth < 10 {
		barWidth = 30
	}
	if barWidth > 80 {
		barWidth = 80
	}

	progBar := progressBar(barWidth, pos, dur)
	progress := fmt.Sprintf("%s %s", progBar, timeStr)

	// Help
	help := helpStyle.Render("space: play/pause • ↑/+: vol up • ↓/-: vol down • q: quit")

	return appStyle.Render(
		title + "\n\n" +
			metadata + "\n\n" +
			stats + "\n" +
			progress + "\n\n" +
			help,
	)
}
