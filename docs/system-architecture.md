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

## Related Docs

- **Codebase Summary:** `docs/codebase-summary.md`
- **Project Overview:** `docs/project-overview-pdr.md`
- **Plan:** `plans/260418-1926-music-web-app/plan.md`
