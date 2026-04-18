---
phase: 04
title: UI Layout Refactor — New Layout & view.go
status: pending
priority: high
effort: 60m
---

# Phase 04 — UI Layout Refactor

## Overview

Extract `View()` from `model.go` into `view.go` (model.go already at 234 LOC). Implement the new 2-column layout: album art left + track info right, visualizer full-width row, improved playlist with track numbers and duration column.

## Related Files

- **Create:** `internal/ui/view.go`
- **Modify:** `internal/ui/model.go` — remove `View()` + `progressBar()`, update Model struct, update tickMsg handler

## Target Layout

```
╔═══════════════════════════════════════════════════════════╗
║  ♪ CLI Music Player                                       ║
╠════════════════════╦══════════════════════════════════════╣
║  ┌──────────────┐  ║  Daft Punk                           ║
║  │▓▓▓▓  ♪  ▓▓▓▓│  ║  Random Access Memories              ║
║  │▓▓▓▓▓▓▓▓▓▓▓▓▓│  ║  Get Lucky                [3 / 12]  ║
║  │▓▓▓▓▓▓▓▓▓▓▓▓▓│  ║                                      ║
║  └──────────────┘  ║  ████████████░░░░  02:34 / 04:12    ║
╠════════════════════╩══════════════════════════════════════╣
║  ▁ ▃ ▅ ▇ █ ▆ ▄ ▂ ▁ ▃ ▆ █ ▇ ▅ ▃ ▁ ▂ ▄ ▆ ▇ █ ▆ ▄ ▂      ║
╠═══════════════════════════════════════════════════════════╣
║  Playlist                                                 ║
║   01  Track One                                  3:24    ║
║  ▶02  Get Lucky                                  4:12    ║
║   03  Instant Crush                              5:37    ║
╠═══════════════════════════════════════════════════════════╣
║  [Space] Play/Pause  [N/P] Next/Prev  [↑↓] Volume [Q] Quit ║
╚═══════════════════════════════════════════════════════════╝
```

## Implementation Steps

### 1. Update `Model` in `model.go`

```go
type Model struct {
    player      *audio.Player
    width       int
    err         error
    viz         visualizer
    art         string // cached rendered album art
    lastTrackIdx int   // detect track change for art refresh
}
```

Remove `View()`, `progressBar()`, `formatDuration()` from `model.go` — they move to `view.go`.

Update `Update()` tickMsg case:

```go
case tickMsg:
    m.viz.update(m.player.GetAmplitude())
    // Refresh art only on track change
    if m.player.PlaylistIdx != m.lastTrackIdx || m.art == "" {
        m.art = renderArt(m.player.Metadata.CoverArt)
        m.lastTrackIdx = m.player.PlaylistIdx
    }
    return m, m.tickCmd()
```

Also refresh art on `trackFinishedMsg`:
```go
case trackFinishedMsg:
    m.player.Next()
    m.art = renderArt(m.player.Metadata.CoverArt)
    m.lastTrackIdx = m.player.PlaylistIdx
    return m, waitForTrackFinished(m.player)
```

### 2. Create `internal/ui/view.go`

```go
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

    // Right column: track info + progress
    status := statusText(m.player.GetState())
    playlistInfo := ""
    if len(m.player.Playlist) > 0 {
        playlistInfo = fmt.Sprintf("  [%d/%d]", m.player.PlaylistIdx+1, len(m.player.Playlist))
    }

    pos := m.player.GetPosition()
    dur := meta.Duration
    timeStr := fmt.Sprintf("%s / %s", formatDuration(pos), formatDuration(dur))
    barWidth := 32 // fixed width inside right column
    bar := gradientProgressBar(barWidth, pos, dur)

    vol := volText(m.player.GetVolume())

    infoCol := lipgloss.JoinVertical(lipgloss.Left,
        artistStyle.Render(meta.Artist),
        albumStyle.Render(meta.Album),
        trackStyle.Render(meta.Title+playlistInfo),
        "",
        bar+" "+timeStr,
        statusStyle.Render(status+" "+vol),
    )
    infoPanel := infoPanelStyle.Render(infoCol)

    // Left column: album art (already rendered string)
    artPanel := artPanelStyle.Render(m.art)

    return lipgloss.JoinHorizontal(lipgloss.Top, artPanel, infoPanel)
}

// renderPlaylist builds the scrollable playlist section.
func renderPlaylist(m Model) string {
    if len(m.player.Playlist) == 0 {
        return ""
    }

    var lines []string
    lines = append(lines, playlistTitleStyle.Render("Playlist"))

    startIdx := m.player.PlaylistIdx - 3
    if startIdx < 0 {
        startIdx = 0
    }
    endIdx := startIdx + 7
    if endIdx > len(m.player.Playlist) {
        endIdx = len(m.player.Playlist)
    }

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

        // Track number + name (duration not available per-track without preloading)
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
    ext := filepath.Ext(name)
    return strings.TrimSuffix(name, ext)
}
```

### 3. Verify compile

```bash
cd /Users/vule/Work/musiccli && go build ./...
```

## Success Criteria

- `go build ./...` passes
- `model.go` drops below 150 LOC after extraction
- TUI shows 2-column layout with art panel on left
- Playlist shows track numbers with `▶` on current track

## Notes

- `lipgloss.JoinHorizontal` handles side-by-side panels natively — no manual padding needed
- Art panel width = `artWidth` cols (20) + 2 padding; info panel takes remaining space
- `gradientProgressBar` uses sub-block chars for smooth fractional fill — looks much nicer than `█░` only
