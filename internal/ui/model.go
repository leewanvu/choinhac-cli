package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"choinhaccli/internal/audio"
)

type tickMsg time.Time
type trackFinishedMsg struct{}

func waitForTrackFinished(p *audio.Player) tea.Cmd {
	return func() tea.Msg {
		<-p.Done()
		return trackFinishedMsg{}
	}
}

// Model represents the application state
type Model struct {
	player       *audio.Player
	width        int
	err          error
	viz          visualizer
	art          string // cached rendered album art, refreshed on track change
	lastTrackIdx int
}

// NewModel creates a new UI model
func NewModel(p *audio.Player) Model {
	return Model{
		player: p,
	}
}

// Init initializes the tea application
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.tickCmd(), waitForTrackFinished(m.player))
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
		case "n", "right":
			m.player.Next()
		case "p", "left":
			m.player.Prev()
		case "r":
			m.player.Random()
		case "=", "+", "up":
			m.player.VolumeUp()
		case "-", "down":
			m.player.VolumeDown()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tickMsg:
		m.viz.update(m.player.GetAmplitude())
		if m.player.PlaylistIdx != m.lastTrackIdx || m.art == "" {
			m.art = renderArt(m.player.Metadata.CoverArt)
			m.lastTrackIdx = m.player.PlaylistIdx
		}
		return m, m.tickCmd()

	case trackFinishedMsg:
		m.player.Next()
		m.art = renderArt(m.player.Metadata.CoverArt)
		m.lastTrackIdx = m.player.PlaylistIdx
		return m, waitForTrackFinished(m.player)

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}
