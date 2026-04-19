---
name: musicweb — Spotify-like Web App
date: 2026-04-18
status: pending
type: feature
slug: music-web-app
blockedBy: []
blocks: []
brainstorm: /Users/vule/Work/musiccli/plans/reports/brainstorm-260418-1926-music-web-app.md
---

# musicweb — Spotify-like Web App Implementation Plan

**Goal:** Add Go-backed web UI for local music library. Self-hosted LAN app, <10 users, Chrome-only, no auth.

## Context Links

- Brainstorm: [brainstorm-260418-1926-music-web-app.md](../reports/brainstorm-260418-1926-music-web-app.md)
- Existing docs: [system-architecture.md](../../docs/system-architecture.md), [codebase-summary.md](../../docs/codebase-summary.md), [code-standards.md](../../docs/code-standards.md)

## Architecture Summary

```
Chrome → React SPA (Vite) ──HTTP──► Go server (chi)
                                    ├─ /api/*  JSON
                                    ├─ /stream/:id  HTTP Range
                                    └─ /*  embed.FS → SPA
                                        ↓
                                    SQLite (library.db)
                                        ↑
                                    Scanner (dhowden/tag) ← Music dir
```

**Zero coupling** with `internal/audio/` (beep TUI engine remains untouched).

## New Modules

| Path | Purpose |
|------|---------|
| `cmd/musiccli/cmd/serve.go` | New Cobra subcommand |
| `internal/web/` | HTTP server, handlers, DTOs, embed SPA |
| `internal/library/` | Scanner, SQLite store, models |
| `web/` | Vite+React+TS frontend source |

## Phases

| # | Phase | Status | File |
|---|-------|--------|------|
| 1 | Foundation: Scanner + DB + API + SPA skeleton + streaming | ✅ complete | [phase-01-foundation.md](./phase-01-foundation.md) |
| 2 | Library UX: Search + detail views + NowPlayingBar controls | ✅ complete | [phase-02-library-ux.md](./phase-02-library-ux.md) |
| 3 | Playlists + Queue (CRUD, reorder, drawer) | ✅ complete | [phase-03-playlist-queue.md](./phase-03-playlist-queue.md) |
| 4 | Polish: cover art, shortcuts, responsive | ✅ complete | [phase-04-polish.md](./phase-04-polish.md) |

## Key Dependencies

- Go: `github.com/go-chi/chi/v5`, `modernc.org/sqlite`
- JS: `react`, `react-dom`, `vite`, `typescript`, `zustand`

## Success Criteria

- `musiccli serve` starts <2s on empty DB
- 1000-track scan <30s
- Play latency <500ms from click to audio
- Seek works on FLAC in Chrome
- Single binary deploy
- All 4 phases merged without touching `internal/audio/`

## Unresolved Questions

1. Config format: YAML (lean) vs TOML vs JSON
2. Commit `web/dist/` or rebuild in CI
3. Single `musiccli serve` binary vs separate `musicweb` binary (lean single)
4. Fallback cover art: directory `cover.jpg` / `folder.jpg`?
5. Default port `8080` vs `8888` vs other

## Risk Summary

See phase files for phase-level risks. Top-level risks:
- No auth → accepted (LAN trust, bind to LAN IP only)
- FLAC on Safari → accepted (Chrome required, document)
- Frontend build pipeline → mitigate via Makefile + pinned Node version
