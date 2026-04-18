---
title: Fancy Music Player UI — Visual Wow
status: pending
created: 2026-04-18
blockedBy: []
blocks: []
---

# Fancy Player UI

Visual upgrade: ASCII album art, simulated audio visualizer, animated gradient progress bar, better playlist.

## Phases

| # | Phase | Status | Est |
|---|-------|--------|-----|
| 01 | Amplitude tracker (audio layer) | pending | 30m |
| 02 | Album art renderer | pending | 45m |
| 03 | Visualizer component | pending | 30m |
| 04 | Refactor UI — new layout & view.go | pending | 60m |
| 05 | Style update | pending | 20m |

## Key Constraints

- No new Go deps — `dhowden/tag` already in go.mod, already used in `player.go`
- `model.go` at 234 LOC — must extract `View()` into `view.go` during phase 04
- `player.go` at 299 LOC — add only a thin wrapper + method, keep changes minimal
- 100ms tick loop drives all UI updates including visualizer decay

## Files

**New:**
- `internal/audio/amplitude-tracker.go`
- `internal/ui/album-art.go`
- `internal/ui/visualizer.go`
- `internal/ui/view.go`

**Modified:**
- `internal/audio/player.go` — wrap streamer, expose `GetAmplitude()`
- `internal/ui/model.go` — add visualizer/art state, update struct
- `internal/ui/style.go` — new color scheme + new styles
