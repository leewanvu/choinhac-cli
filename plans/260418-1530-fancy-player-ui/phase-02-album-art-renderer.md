---
phase: 02
title: Album Art Renderer
status: pending
priority: high
effort: 45m
---

# Phase 02 — Album Art Renderer

## Overview

Extract embedded cover art from audio tags via `dhowden/tag` (already imported in `player.go`). Convert the image to Unicode half-block art (`▀▄`) with 24-bit terminal color. Show music-note placeholder when no cover art available.

## Related Files

- **Create:** `internal/ui/album_art.go`
- **Modify:** `internal/audio/player.go` — expose `CoverArt []byte` from metadata

## Architecture

```
player.extractMetadata()
    → tag.ReadFrom() → tag.Picture() → []byte (JPEG/PNG)
    → stored in Player.CoverArt

UI tick → albumArt(coverBytes, width, height) → string
    → decode image → scale to target size
    → half-block render: each pair of rows = one terminal row
    → lipgloss inline 24-bit colors
```

## Implementation Steps

### 1. Expose CoverArt from Player

Add `CoverArt []byte` to `TrackMetadata`:

```go
type TrackMetadata struct {
    Title      string
    Artist     string
    Album      string
    SampleRate int
    Duration   time.Duration
    CoverArt   []byte // raw JPEG/PNG bytes, nil if none
}
```

In `extractMetadata()`, after reading tags:

```go
if pic := m.Picture(); pic != nil {
    p.Metadata.CoverArt = pic.Data
}
```

### 2. Create `internal/ui/album_art.go`

```go
package ui

import (
    "bytes"
    "image"
    _ "image/jpeg"
    _ "image/png"
    "strings"

    "github.com/charmbracelet/lipgloss"
)

const (
    artWidth  = 20 // terminal columns
    artHeight = 10 // terminal rows (each row = 2 image rows via half-block)
)

// placeholder shown when no cover art is available
var artPlaceholder = []string{
    "┌──────────────────┐",
    "│                  │",
    "│   ♪  ♫  ♪  ♫   │",
    "│                  │",
    "│   ♫  ♬  ♫  ♬   │",
    "│                  │",
    "│   ♪  ♫  ♪  ♫   │",
    "│                  │",
    "│   ♫  ♬  ♫  ♬   │",
    "└──────────────────┘",
}

// renderArt converts cover art bytes to a half-block terminal string.
// Falls back to placeholder on any error.
func renderArt(data []byte) string {
    if len(data) == 0 {
        return renderPlaceholder()
    }
    img, _, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        return renderPlaceholder()
    }
    return renderImage(img)
}

func renderPlaceholder() string {
    style := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
    var sb strings.Builder
    for _, line := range artPlaceholder {
        sb.WriteString(style.Render(line))
        sb.WriteString("\n")
    }
    return sb.String()
}

func renderImage(img image.Image) string {
    bounds := img.Bounds()
    srcW := bounds.Max.X - bounds.Min.X
    srcH := bounds.Max.Y - bounds.Min.Y

    scaleX := float64(srcW) / float64(artWidth)
    scaleY := float64(srcH) / float64(artHeight*2) // *2: half-block uses 2 rows per terminal row

    var sb strings.Builder
    for row := 0; row < artHeight; row++ {
        for col := 0; col < artWidth; col++ {
            // Upper pixel of the half-block pair
            ux := bounds.Min.X + int(float64(col)*scaleX)
            uy := bounds.Min.Y + int(float64(row*2)*scaleY)
            // Lower pixel
            ly := bounds.Min.Y + int(float64(row*2+1)*scaleY)

            ur, ug, ub, _ := img.At(ux, uy).RGBA()
            lr, lg, lb, _ := img.At(ux, ly).RGBA()

            upper := lipgloss.Color(rgbHex(uint8(ur>>8), uint8(ug>>8), uint8(ub>>8)))
            lower := lipgloss.Color(rgbHex(uint8(lr>>8), uint8(lg>>8), uint8(lb>>8)))

            // '▀' = upper half block; foreground = upper color, background = lower color
            cell := lipgloss.NewStyle().
                Foreground(upper).
                Background(lower).
                Render("▀")
            sb.WriteString(cell)
        }
        sb.WriteString("\n")
    }
    return sb.String()
}

func rgbHex(r, g, b uint8) string {
    return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
```

Add `"fmt"` to imports.

### 3. Verify compile

```bash
cd /Users/vule/Work/musiccli && go build ./...
```

## Success Criteria

- `go build ./...` passes
- `renderArt(nil)` returns placeholder string
- `renderArt(validJpegBytes)` returns colored half-block string

## Notes

- `image/jpeg` and `image/png` are stdlib — no new deps
- 24-bit color (`#RRGGBB`) supported in iTerm2, Terminal.app (macOS 14+), most modern terminals
- Half-block technique: each `▀` char uses foreground = top pixel, background = bottom pixel → effectively doubles vertical resolution
- `artWidth=20` → 20 terminal columns wide. At typical 8px char width = 160px image display width
