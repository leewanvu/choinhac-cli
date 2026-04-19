---
phase: 05
title: Style Update
status: pending
priority: medium
effort: 20m
---

# Phase 05 — Style Update

## Overview

Replace `style.go` with a richer color scheme and add new styles needed by the 2-column layout, visualizer, and art panels.

## Related Files

- **Modify:** `internal/ui/style.go` — replace entirely

## New Style Definitions

```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
    appStyle = lipgloss.NewStyle().Padding(1, 2)

    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#FF79C6")).
        MarginBottom(1)

    // Art panel — fixed width to match artWidth + padding
    artPanelStyle = lipgloss.NewStyle().
        Width(24).
        PaddingRight(2)

    // Info panel — flex fills remaining width
    infoPanelStyle = lipgloss.NewStyle().
        PaddingLeft(1)

    artistStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#CBA6F7")).
        Bold(true).
        MarginBottom(0)

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
```

## Color Palette (Catppuccin Mocha inspired)

| Role | Color | Hex |
|------|-------|-----|
| Title / accent | Pink | `#FF79C6` |
| Artist | Mauve | `#CBA6F7` |
| Album | Blue | `#89B4FA` |
| Track name | Text | `#F8F8F2` |
| Progress bar fg | Mauve | `#CBA6F7` |
| Progress bar bg | Surface1 | `#313244` |
| Visualizer | Sky | `#89DCEB` |
| Current track | Green | `#A6E3A1` |
| Playlist items | Overlay0 | `#9399B2` |
| Track numbers | Surface2 | `#585B70` |
| Help text | Surface2 | `#585B70` |

## Success Criteria

- `go build ./...` passes
- All styles referenced in `view.go` and `visualizer.go` are defined here
- No undefined style variable errors

## Notes

- Catppuccin Mocha is a widely loved dark terminal palette — cohesive, readable, and modern
- Remove old `labelStyle`, `valueStyle`, `metadataStyle`, `statsStyle` — they're replaced by `artistStyle`, `albumStyle`, `trackStyle`, `statusStyle`
