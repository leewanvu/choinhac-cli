---
phase: 03
title: Visualizer Component
status: pending
priority: high
effort: 30m
---

# Phase 03 — Visualizer Component

## Overview

Simulated bar visualizer driven by real amplitude from Phase 01. Each 100ms tick, poll `Player.GetAmplitude()`, distribute amplitude across N bars with randomized spread, apply exponential decay so bars fall smoothly.

## Related Files

- **Create:** `internal/ui/visualizer.go`
- **Modify:** `internal/ui/model.go` — add `viz *visualizer` to Model

## Architecture

```
100ms tickMsg
    → viz.update(player.GetAmplitude())
        → spread amplitude to bar heights with random variation
        → apply decay to all bars (bars[i] *= decayFactor)
    → viz.render(width) → string of block chars
```

## Implementation Steps

### 1. Create `internal/ui/visualizer.go`

```go
package ui

import (
    "math/rand/v2"
    "strings"
)

const (
    vizBars      = 24    // number of bars
    vizDecay     = 0.75  // exponential decay per tick (higher = slower fall)
    vizSensitivity = 8.0 // multiplier to make bars more reactive
)

// block chars ordered low→high amplitude
var blocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

type visualizer struct {
    bars [vizBars]float64
}

// update feeds a new amplitude sample into the visualizer state.
func (v *visualizer) update(amplitude float64) {
    scaled := amplitude * vizSensitivity
    if scaled > 1.0 {
        scaled = 1.0
    }

    for i := range v.bars {
        // Each bar gets amplitude ± random spread
        spread := (rand.Float64()*0.4 - 0.2) * scaled
        target := scaled + spread
        if target < 0 {
            target = 0
        }
        if target > 1.0 {
            target = 1.0
        }
        // Take max of target and decayed current (bars rise fast, fall slow)
        decayed := v.bars[i] * vizDecay
        if target > decayed {
            v.bars[i] = target
        } else {
            v.bars[i] = decayed
        }
    }
}

// render returns a single-line string of block characters for the visualizer.
func (v *visualizer) render() string {
    var sb strings.Builder
    for _, h := range v.bars {
        idx := int(h * float64(len(blocks)-1))
        if idx >= len(blocks) {
            idx = len(blocks) - 1
        }
        sb.WriteRune(blocks[idx])
        sb.WriteRune(' ') // space between bars for breathing room
    }
    return vizStyle.Render(strings.TrimRight(sb.String(), " "))
}
```

### 2. Add `vizStyle` to `internal/ui/style.go`

```go
vizStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#C792EA")).
    Bold(true)
```

### 3. Add `viz visualizer` to `Model` in `model.go`

```go
type Model struct {
    player *audio.Player
    width  int
    err    error
    viz    visualizer
    art    string // cached rendered album art, refreshed on track change
}
```

In `Update()`, on `tickMsg`:

```go
case tickMsg:
    m.viz.update(m.player.GetAmplitude())
    return m, m.tickCmd()
```

On `trackFinishedMsg` and initial load — refresh `m.art`:

```go
m.art = renderArt(m.player.Metadata.CoverArt)
```

### 4. Verify compile

```bash
cd /Users/vule/Work/musiccli && go build ./...
```

## Success Criteria

- `go build ./...` passes
- `viz.render()` returns a non-empty string of block chars
- Bars animate on each tick when music plays, decay to `▁` when paused/stopped

## Notes

- `vizDecay = 0.75` → bars halve roughly every 3 ticks (300ms). Feels snappy but not jittery.
- `vizSensitivity = 8.0` → typical beep amplitude is 0.1–0.3 for normal music; scaling to 0.8–1.0 fills bars nicely.
- Album art cached in `m.art` string — re-render only on track change (not every tick) to avoid expensive image decode.
