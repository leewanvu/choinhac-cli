# System Architecture

**Project:** musiccli  
**Version:** v2.0.0 (all 4 phases complete)  
**Last Updated:** 2026-04-19

## High-Level Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           User Interfaces                                │
│  CLI: musiccli play   │  CLI: musiccli feel   │  WEB: localhost:8080     │
│  (TUI Playback)       │  (AI Review)          │  (React SPA — Spotify-like) │
└─────────────┬─────────────────────────────────────────────────────────────┘
              │
     ┌────────▼──────────────────────┐
     │  cmd/musiccli (Cobra Router)  │
     └────────┬──────────────────────┘
              │
     ┌────────┴─────────────────────────────────┐
     │                                          │
 ┌───▼────────┐  ┌──────▼──────┐     ┌─────────▼──────────┐
 │  play.go   │  │  feel.go    │     │  serve.go          │
 │  (TUI)     │  │  (AI)       │     │  (Web Server)      │
 └───┬────────┘  └──────┬──────┘     └─────────┬──────────┘
     │                  │                       │
 ┌───▼──────────┐  ┌────▼──────────┐  ┌────────▼──────────────────┐
 │ audio/ui     │  │ agent/        │  │ config/ library/ web/     │
 │ Player + TUI │  │ analyzer      │  │ Store + Scanner + Server  │
 └──────────────┘  └───────────────┘  └───────────────────────────┘
```

## Web Server Architecture (`musiccli serve`)

### Command: `musiccli serve [--music-dir] [--port] [--bind] [--db]`

**Flow:**
1. Load config (YAML + CLI flags)
2. Open SQLite library DB (`~/.config/musiccli/library.db`)
3. Derive `coverDir = dataDir/covers/`
4. Start background scan of music directory (async, extracts cover art)
5. Start chi HTTP server
6. Serve API endpoints + cover images + embedded React SPA

### Architecture Diagram

```
Chrome (SPA) ──HTTP──► Go server (chi)
                        ├── /api/library/*    JSON (tracks, albums, artists)
                        ├── /api/search       Full-text search
                        ├── /api/playlists/*  CRUD + reorder
                        ├── /api/scan         Async scan + cover extraction
                        ├── /stream/{id}      HTTP Range (FLAC/WAV/MP3)
                        ├── /cover/{albumID}  JPEG cover (7-day cache)
                        └── /*                embed.FS → React SPA
                             ↓
                        SQLite (library.db)
                             ↑
                        Scanner (dhowden/tag) ← Music dir
                        CoverExtractor        ← ~/.config/musiccli/covers/
```

### Core Components

**1. Configuration** (`internal/config/`)
- Load YAML from `~/.config/musiccli/config.yaml`
- CLI flag overrides: `--music-dir`, `--port`, `--bind`, `--db`
- Defaults: port 8080, bind 127.0.0.1, db `~/.config/musiccli/library.db`
- `coverDir` derived automatically as `{dbDir}/covers/`

**2. Library DB** (`internal/library/`)
- SQLite tables: `artists`, `albums` (with `cover_path`), `tracks`, `playlists`, `playlist_tracks`
- `scanner.go` — async `WalkDir` + metadata via `dhowden/tag`; incremental (mtime skip)
- `cover_extractor.go` — per-album: tries embedded art → directory `cover.jpg`/`folder.jpg`/`album.jpg`; saves `{albumID}.jpg`
- `store.go` — full CRUD: tracks, albums, artists, playlists, playlist_tracks

**3. HTTP Server** (`internal/web/`)
- chi router with Logger/Recoverer/CORS middleware
- `handlers/library.go` — paginated tracks, album/artist detail, full-text search
- `handlers/stream.go` — `http.ServeContent` with Range support (206 Partial Content)
- `handlers/scan.go` — async scan progress polling
- `handlers/playlist.go` — playlist CRUD + position-based reorder
- `handlers/cover.go` — serves `covers/{albumID}.jpg` with `Cache-Control: max-age=604800`; SVG placeholder if missing

**4. Frontend SPA** (`web/`)
- Vite 6 + React 18 + TypeScript
- State: Zustand (`player.ts` with persist, `ui.ts`)
- Routing: React Router v6 (Library / Search / Album / Artist / Playlist)
- Virtualization: `react-window` FixedSizeList (56px rows) for up to 5000 tracks
- Drag-reorder: `@dnd-kit` in playlist detail view
- Responsive: static sidebar on ≥768px; hamburger + overlay on mobile

### Data Flow: Scan + Cover Extraction

```
POST /api/scan
  ↓
ScanAsync(musicDir, store, coverDir) — goroutine
  ↓ for each audio file:
  ├── read mtime → skip if unchanged
  ├── dhowden/tag.ReadFrom() → title, artist, album, track#
  ├── store.UpsertTrack()
  └── ExtractCover(albumID, trackPath, coverDir)
       ├── try tag.Picture().Data → write {albumID}.jpg
       └── else try cover.jpg / folder.jpg / album.jpg in same dir
  ↓
GET /api/scan/status → { running, scanned, total, done }
```

### Data Flow: Playback

```
User clicks track in SPA
  ↓
player.playTrack(track, queue) → engine.play('/stream/{id}')
  ↓
new Audio('/stream/{id}').play()
  ↓
GET /stream/{id}
  ├── store.GetTrack(id) → FilePath
  └── http.ServeContent() → Range-aware stream
  ↓
Browser decodes + plays (Chrome: FLAC/WAV/MP3)
  ↓
engine.on('ended') → player.next() → advance queue
```

### API Contracts

**GET /api/library/albums**
```json
{ "albums": [{ "id": 1, "title": "Album", "artist": "Artist", "year": 2020, "cover_path": "/path" }] }
```

**GET /cover/{albumID}**
- Returns: JPEG image or SVG placeholder
- `Cache-Control: max-age=604800` on hit
- `Cache-Control: max-age=3600` on SVG placeholder

**GET /stream/{id}**
- Returns: Raw audio with `Content-Type: audio/flac` etc.
- Supports: `Range: bytes=N-M` → `206 Partial Content`

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| `modernc.org/sqlite` | Pure Go, no CGO → single-binary deploy |
| `chi` router | Lightweight, composable middleware |
| Embedded SPA | `go:embed web/dist/*` → single binary, no file serving |
| Async scanner | Non-blocking; progress channel pattern |
| `http.ServeContent` | Native Range support for seeking |
| Zustand persist | Player state survives page refresh |
| `react-window` | Virtualizes 5000+ tracks without DOM bloat |
| Cover cache dir | Files keyed by album ID → no path traversal possible |
| CORS dev-only | Only allows `localhost:5173`; production is same-origin |

## Performance Targets

| Operation | Target | Status |
|-----------|--------|--------|
| Empty DB startup | <2s | ✓ |
| 1000-track scan | <30s | ✓ |
| API track list | <100ms | ✓ |
| Audio stream first byte | <50ms | ✓ |
| SPA load | <500ms | ✓ |
| 5000-row scroll | 60fps | ✓ (react-window) |

## Security

- Bind default: `127.0.0.1` (localhost only)
- Track/album IDs: integer validation; no path traversal
- Cover dir: files named `{integer}.jpg`; never user-controlled path
- No auth (LAN trust model, documented)

## Testing Status

| Component | Tests | Status |
|-----------|-------|--------|
| Library store | Yes (`store_test.go`) | ✓ |
| HTTP handlers | Yes (`handlers_test.go`) | ✓ |
| Scanner | Integration via store tests | ✓ |
| SPA | Manual + browser | ✓ |

## TUI Architecture (`musiccli play`)

### Command Flow

```
play.go → cmd.RunE
  ↓
1. Validate file/directory path
2. Build playlist from file(s)
3. audio.InitSpeaker() — initialize system speaker (global, once only)
4. Create Player + load first track
5. Create TUI Model + run bubbletea.Run()
```

### BubbleTea Model Loop

```go
type Model struct {
    player *audio.Player     // Playback controller
    viz visualizer          // 24-bar frequency analyzer
    art string              // Cached rendered album art
    width, height int       // Terminal dimensions
}

// Model implements tea.Model interface:
// Init() — Setup commands (tick, listeners)
// Update(msg) — Handle keyboard, tick, events
// View() — Render TUI to string
```

**Message Types:**

| Message | Source | Purpose |
|---------|--------|---------|
| `tea.KeyMsg` | Keyboard input | User controls (space, arrows, q, r, +/-, n/p) |
| `tickMsg` | Timer (100ms) | Update frequency bars, refresh progress bar |
| `trackFinishedMsg` | Goroutine listener | Auto-play next track on end |
| `tea.WindowSizeMsg` | Terminal | Handle resize; recalculate layout |

**Update Loop:**
```
100ms tick → visualizer.update(amplitude) → View() re-renders → 24 bars animated
             ↓
Keyboard input → player.Play/Pause/Next/etc → state change → View() re-renders
             ↓
Track end (via done channel) → trackFinishedMsg → auto-next
```

### Display Layout

```
┌────────────────────────────────────────────────┐
│  Album Art   │  Track Info (20x10 cells)      │  ← Top panel (30% height)
│  (Unicode    │  Title, Artist, Album          │
│  half-blocks │  Sample Rate, Duration, Status │
│  + border)   │  Volume offset, Time position  │
├────────────────────────────────────────────────┤
│  ▁▂▃▄▅▆▇█ ▆▅▄▃▂▁ ▂▃▄▅▆▇█▇▆▅▄▃▂  (24-bar viz)│  ← Visualizer (10% height)
├────────────────────────────────────────────────┤
│ [████████░░░░░░░░░░░░░░] 1:23 / 3:45 Playing │  ← Progress (5% height)
├────────────────────────────────────────────────┤
│ 1. Current Track (highlighted)                 │  ← Playlist (50% height)
│ 2. Next track                                  │     7 tracks visible
│ 3. Next + 1                                    │
│ ...                                            │
├────────────────────────────────────────────────┤
│ space: play/pause  n/→: next  p/←: prev       │  ← Help footer (5% height)
│ r: random  +/↑: volume up  -/↓: down  q: quit │
└────────────────────────────────────────────────┘
```

### Album Art Rendering

**Steps:**
1. Extract JPEG/PNG from metadata (dhowden/tag)
2. Decode image bytes → raw pixels
3. Resize to 20x10 cells via bilinear sampling
4. Map pixels to Unicode half-block characters: ▀▁▂▃▄▅▆▇█
5. Apply true-color ANSI codes per cell
6. Render with border (Rosewater #F5E0DC)

**Performance:** <50ms for typical 500x500 JPEG

### Visualizer (24-Bar Frequency Display)

**Algorithm:**
```
amplitude = player.GetAmplitude()  // [0.0, 1.0] from atomic tracker
  ↓
bars[i] = amplitude + random_noise * 0.1  // Add noise spread
  ↓
if bars[i] > prev_bars[i]:
  bars[i] = bars[i]  // Rise instantly
else:
  bars[i] = prev_bars[i] * 0.75  // Decay (slow fall)
  ↓
Render each bar as ▁▂▃▄▅▆▇█ (height = bars[i] * 8)
```

**Colors:** Gradient Lavender → Flamingo (Catppuccin Mocha)

**Update Rate:** Every 100ms tick

### Keyboard Control Flow

```
tea.KeyMsg(key) → Model.Update()
  ↓
switch key.String():
  "space" → player.TogglePause()
  "n", "right" → player.Next()
  "p", "left" → player.Prev()
  "r" → player.Random()
  "+", "up" → player.VolumeUp()
  "-", "down" → player.VolumeDown()
  "q", "ctrl+c" → tea.Quit()
  ↓
State change → View() re-renders immediately
```

---

## Audio Engine (beep Layer)

### Speaker Initialization

```go
// GlobalSpeaker initialized once per process
var (
  speaker beep.Speaker
  once sync.Once
)

func InitSpeaker() error {
  var err error
  once.Do(func() {
    sampleRate := beep.SampleRate(44100)  // 44.1kHz base
    bufferSize := sampleRate.N(100 * time.Millisecond) // 4410 samples
    err = speaker.Init(sampleRate, bufferSize)
  })
  return err
}
```

**Key Points:**
- 44.1kHz base rate (industry standard)
- 100ms buffer → ~4400 samples
- Resampling handled automatically by beep for 48kHz, 96kHz, 192kHz sources
- Lock-free amplitude tracking via atomic operations

### Playback Pipeline

```
Audio File (FLAC/WAV)
  ↓ [file handler selects decoder]
gopxl/beep FLAC Decoder | WAV Decoder
  ↓ [raw PCM samples]
Resampler (if rate != 44.1kHz)
  ↓ [amplitude tracking]
AmplitudeTracker (atomic.Uint64 for lock-free reads)
  ↓ [volume control]
effects.Volume (gain adjustment, dB)
  ↓ [global control]
beep.Ctrl (pause/resume state machine)
  ↓ [playback]
GlobalSpeaker.Play() → System Audio Output
```

### Amplitude Tracking (Lock-Free)

```go
type amplitudeTracker struct {
  wrapped beep.Streamer
  peak atomic.Uint64
}

// Lock-free write (no mutex, no channels)
func (a *amplitudeTracker) Stream(samples [][2]float64) (int, bool) {
  n, ok := a.wrapped.Stream(samples)
  var maxAmp float64
  for _, [2]float64 { ch1, ch2 } := range samples[:n] {
    amp := (math.Abs(ch1) + math.Abs(ch2)) / 2
    if amp > maxAmp {
      maxAmp = amp
    }
  }
  // Atomic store (no locks)
  a.peak.Store(math.Float64bits(maxAmp))
  return n, ok
}

// Lock-free read (from UI update goroutine)
func (p *Player) GetAmplitude() float64 {
  bits := p.tracker.peak.Load()
  return math.Float64frombits(bits)
}
```

**Rationale:** UI poll rate (100ms) vs. audio stream rate (~44kHz) creates lock contention; atomic operations eliminate mutex overhead.

### Playlist Navigation

**State Machine:**
```
[Stopped] --LoadAndPlay--> [Playing] --Pause--> [Paused] --Play--> [Playing]
            or Play                                          or toggle
                                                            
[Playing] --Stop--> [Stopped]
[Paused] --Stop--> [Stopped]

Next/Prev/Random work in Playing or Paused state.
```

**Wraparound:**
```go
// Next wraps to first track
nextIdx := (p.PlaylistIdx + 1) % len(p.Playlist)

// Prev wraps to last track
prevIdx := (p.PlaylistIdx - 1 + len(p.Playlist)) % len(p.Playlist)

// Random: uniform distribution
randomIdx := rand.Intn(len(p.Playlist))
```

---

## Related Docs

- **Codebase Summary:** `docs/codebase-summary.md`
- **Project Overview:** `docs/project-overview-pdr.md`
- **Code Standards:** `docs/code-standards.md`
- **Plan:** `plans/260418-1926-music-web-app/plan.md`
