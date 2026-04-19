---
type: brainstorm
date: 2026-04-18
slug: music-web-app
status: approved
---

# Brainstorm: musicweb — Spotify-like Web Page for Local Library

## Problem Statement

Add web-based music listening UI (Spotify-like) on top of existing `musiccli` project. Current project is Go CLI + TUI playing local FLAC/WAV; web app is additive, not replacement.

## Confirmed Requirements

| Dimension | Decision |
|-----------|----------|
| Platform | Separate web app (browser UI) |
| Music source | Local files only (FLAC/WAV, same as CLI) |
| MVP features | Library browse + Search, Playlist management, Now Playing + Queue |
| Users | Family/friends <10, self-hosted, LAN |
| Playback location | Browser of each user (HTML5 `<audio>`) |
| Stack | Go backend + embedded React SPA + SQLite |
| Auth | None (trust LAN) |
| Safari | Unsupported; Chrome required |

## Evaluated Approaches

### A. Go monolith + HTMX + server templates
- **Pros:** no JS build, fastest ship, single language
- **Cons:** queue drag-drop awkward, less Spotify-like polish
- **Verdict:** rejected — user wants Spotify-like UX

### B. Go API + separate Next.js frontend (2 deployables)
- **Pros:** modern DX, SSR if needed
- **Cons:** 2 deploy artifacts, Node runtime required on host, overkill for 10 users
- **Verdict:** rejected — violates KISS for self-hosted scale

### C. Go monolith + embedded React SPA (Vite) — **CHOSEN**
- **Pros:** single binary (`//go:embed` SPA dist), modern UX, reuses Go stack, deploy = `scp binary && run`
- **Cons:** frontend build step added to repo
- **Verdict:** best balance

### D. Full rewrite in Node/Next.js
- **Pros:** unified JS stack
- **Cons:** abandons existing Go investment, no reuse of CLI metadata code
- **Verdict:** rejected

## Final Architecture

```
Chrome → React SPA (Vite) → HTTP → Go server (chi)
                                   ├─ /api/*  JSON
                                   ├─ /stream/:id  Range
                                   └─ /*  embed.FS → SPA
                                       ↓
                            SQLite (~/.musiccli/library.db)
                                       ↑
                            Scanner (dhowden/tag) ← music dir
```

**Zero coupling** with existing `internal/audio/` (beep). Web = new module tree.

## Module Layout (new)

```
cmd/musiccli/cmd/serve.go            ← new Cobra subcommand
internal/web/
  ├─ server.go           (chi router, embed.FS SPA)
  ├─ handlers/           (library, search, playlist, stream)
  ├─ dto.go              (request/response types)
  └─ spa_embed.go        (//go:embed web/dist)
internal/library/
  ├─ scanner.go          (walk + metadata extract)
  ├─ store.go            (SQLite repo)
  └─ models.go
web/                     ← Vite+React+TS
  ├─ src/{pages,components,store,api,audio}
  ├─ vite.config.ts
  └─ package.json
```

All new Go files <200 LOC target (per development-rules.md).

## Data Model

```sql
tracks(id, path UNIQUE, title, artist_id, album_id, duration_ms,
       sample_rate, bit_depth, format, mtime, added_at)
artists(id, name UNIQUE)
albums(id, name, artist_id, year, cover_path)
playlists(id, name, created_at, updated_at)
playlist_tracks(playlist_id, track_id, position, added_at)
scan_state(key, value)
```

No users table. Playlists shared globally (LAN family usage).

## API Surface

```
GET    /api/library/tracks?limit&offset&sort
GET    /api/library/albums
GET    /api/library/albums/:id/tracks
GET    /api/library/artists
GET    /api/search?q=
GET    /api/playlists
POST   /api/playlists
GET    /api/playlists/:id
PUT    /api/playlists/:id
DELETE /api/playlists/:id
POST   /api/playlists/:id/tracks
DELETE /api/playlists/:id/tracks/:track_id
PUT    /api/playlists/:id/reorder
POST   /api/scan
GET    /api/scan/status
GET    /stream/:track_id        (HTTP Range, audio/flac|audio/wav)
GET    /cover/:album_id
```

Queue = client-side only (localStorage). No server queue state → KISS.

## Audio Streaming

- `http.ServeContent` handles Range automatically
- Browser `<audio src="/stream/:id">` seeks natively on FLAC/WAV (Chrome)
- No transcoding; document "Chrome required" on landing

## Library Indexing

- Startup: if DB empty or `--rescan`, walk music dir
- Metadata: `github.com/dhowden/tag` (already in go.mod)
- Incremental: compare file mtime vs stored mtime
- Config: `~/.config/musiccli/config.yaml` with `music_dir:`
- Cover art: extract once → `~/.musiccli/covers/:album_id.jpg`

## New Dependencies

- Go: `github.com/go-chi/chi/v5`, `modernc.org/sqlite` (pure-Go, no CGO)
- JS: `react`, `react-dom`, `vite`, `typescript`, `zustand`, `@tanstack/react-query` (optional)

## New Command

```bash
musiccli serve --port 8080 --music-dir ~/Music --db ~/.musiccli/library.db
```

Existing `play` and `feel` commands untouched.

## Phased Delivery

| Phase | Deliverable |
|-------|------------|
| 1 | Scanner + SQLite + `/api/library/*` + SPA skeleton + `/stream/:id` + single-track playback |
| 2 | Search + album/artist detail views + NowPlayingBar (play/pause/seek/next/prev) |
| 3 | Playlist CRUD + reorder + Queue drawer |
| 4 | Polish: cover art display, keyboard shortcuts, responsive layout |

## Risks

| Risk | Mitigation |
|------|-----------|
| Large library scan blocks startup | Async scan + progress endpoint |
| No auth → anyone on LAN can CRUD playlists | Accepted; bind to LAN IP only; document |
| FLAC fails on Safari/iOS | Document Chrome-only; no transcode |
| Frontend build couples repo to Node | Pin Node version; commit `web/dist/` as fallback |
| SQLite corruption on concurrent writes | `modernc.org/sqlite` WAL mode; single writer |

## Out of Scope (MVP)

- AI `feel` integration (deferred Phase 5)
- Recommendations / history tracking
- Multi-user auth / per-user playlists
- Mobile-specific PWA
- Gapless playback / equalizer / EQ
- Transcoding / Safari support

## Success Criteria

- `musiccli serve` starts <2s on empty DB
- Library scan of 1000 tracks <30s
- Browser play latency <500ms after click
- Seek on FLAC works in Chrome
- Playlist CRUD round-trip <100ms
- Single binary deploy (`scp musiccli host:` then run)

## Unresolved Questions

1. Config file format: YAML vs TOML vs JSON? (lean YAML, matches roadmap)
2. Should `web/dist/` be committed or built in CI? (default: gitignored, build script in Makefile)
3. Binary naming: same `musiccli` binary adds `serve` subcommand vs separate `musicweb` binary? (lean single binary — matches current Cobra pattern)
4. Cover art for tracks without embedded art: fall back to directory `cover.jpg` / `folder.jpg`?
5. Desired port default: 8080 vs something less common?
