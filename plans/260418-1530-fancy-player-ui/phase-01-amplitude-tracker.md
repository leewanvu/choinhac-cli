---
phase: 01
title: Amplitude Tracker (Audio Layer)
status: pending
priority: high
effort: 30m
---

# Phase 01 — Amplitude Tracker

## Overview

Wrap the beep `Streamer` in a thin pass-through that captures peak amplitude per audio chunk. Expose via `Player.GetAmplitude()` so the UI visualizer can poll it every 100ms tick.

## Related Files

- **Create:** `internal/audio/amplitude_tracker.go`
- **Modify:** `internal/audio/player.go`

## Architecture

```
beep.Ctrl → amplitudeTracker → effects.Volume → speaker
                  ↓
          atomic float64 (peak)
                  ↓
         Player.GetAmplitude() ← UI tick polls
```

## Implementation Steps

### 1. Create `internal/audio/amplitude_tracker.go`

```go
package audio

import (
    "math"
    "sync/atomic"
    "math/bits"
    "unsafe"

    "github.com/gopxl/beep"
)

// amplitudeTracker wraps a Streamer and captures peak amplitude per chunk.
type amplitudeTracker struct {
    beep.Streamer
    peak atomic.Uint64 // stores float64 bits via math.Float64bits
}

func newAmplitudeTracker(s beep.Streamer) *amplitudeTracker {
    return &amplitudeTracker{Streamer: s}
}

func (a *amplitudeTracker) Stream(samples [][2]float64) (n int, ok bool) {
    n, ok = a.Streamer.Stream(samples)
    peak := 0.0
    for _, s := range samples[:n] {
        if v := math.Abs(s[0]); v > peak {
            peak = v
        }
        if v := math.Abs(s[1]); v > peak {
            peak = v
        }
    }
    a.peak.Store(math.Float64bits(peak))
    return
}

func (a *amplitudeTracker) amplitude() float64 {
    return math.Float64frombits(a.peak.Load())
}
```

> Note: `atomic.Uint64` with `math.Float64bits` is the standard lock-free float64 atomic pattern in Go.

### 2. Modify `internal/audio/player.go`

Add `tracker *amplitudeTracker` field to `Player` struct:

```go
type Player struct {
    // ...existing fields...
    tracker *amplitudeTracker
}
```

In `LoadAndPlay()`, wrap the streamer **before** passing to `beep.Ctrl`:

```go
// After: var finalStream beep.Streamer = looped / resampled
p.tracker = newAmplitudeTracker(finalStream)
p.ctrl = &beep.Ctrl{Streamer: p.tracker, Paused: false}
```

Add `GetAmplitude()` method:

```go
// GetAmplitude returns peak amplitude [0.0, 1.0] from last audio chunk
func (p *Player) GetAmplitude() float64 {
    if p.tracker == nil || p.state != StatePlaying {
        return 0
    }
    return p.tracker.amplitude()
}
```

### 3. Verify compile

```bash
cd /Users/vule/Work/musiccli && go build ./...
```

## Success Criteria

- `go build ./...` passes with no errors
- `Player.GetAmplitude()` returns 0 when stopped/paused, non-zero when playing

## Notes

- `atomic.Uint64` requires Go 1.19+. Project uses Go 1.25.1 — fine.
- Amplitude range: beep samples are `[-1.0, 1.0]` normalized float64 per channel. Peak will be in `[0.0, 1.0]`.
- No lock needed — atomic read/write is sufficient since UI only reads, audio goroutine only writes.
