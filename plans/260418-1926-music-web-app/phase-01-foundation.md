---
phase: 1
name: Foundation — Scanner + DB + API + SPA skeleton + streaming
status: complete
priority: critical
effort: 3-5 days
completedDate: 2026-04-18
---

# Phase 1: Foundation

## Overview

Ship walking skeleton: user can start server, scan music dir, browse tracks in basic UI, click play → audio streams in browser.

## Key Insights

- `dhowden/tag` already in `go.mod` → reuse for metadata
- `http.ServeContent` handles Range requests natively
- `modernc.org/sqlite` = pure Go, no CGO → single-binary deploy preserved
- Embed frontend build via `//go:embed web/dist/*`

## Requirements

### Functional

- [x] `musiccli serve` subcommand boots HTTP server on configurable port
- [x] First-run scans music dir, populates SQLite
- [x] `/api/library/tracks` returns paginated JSON list
- [x] `/stream/:id` streams FLAC/WAV with Range support
- [x] Minimal SPA: track list + click-to-play via `<audio>`
- [x] Graceful shutdown on SIGINT

### Non-Functional

- [x] Empty DB startup <2s
- [x] 1000-track scan <30s
- [x] Binary includes SPA (no separate static serve)
- [x] Each Go file <200 LOC

## Architecture

```
cmd/musiccli/cmd/serve.go
   └─ internal/web/server.go (chi Mux, routes)
        ├─ internal/web/handlers/library.go → internal/library/store.go
        ├─ internal/web/handlers/stream.go  → os.Open + ServeContent
        └─ internal/web/spa_embed.go        → embed.FS fallback

internal/library/
   ├─ scanner.go   (filepath.WalkDir + dhowden/tag)
   ├─ store.go     (sqlite CRUD)
   └─ models.go    (Track, Album, Artist structs)
```

## Files to Create

```
cmd/musiccli/cmd/serve.go
internal/web/server.go
internal/web/spa_embed.go
internal/web/handlers/library.go
internal/web/handlers/stream.go
internal/web/dto.go
internal/library/scanner.go
internal/library/store.go
internal/library/models.go
internal/library/migrations.go
internal/config/config.go           (load YAML from ~/.config/musiccli/)
web/package.json
web/vite.config.ts
web/tsconfig.json
web/index.html
web/src/main.tsx
web/src/App.tsx
web/src/pages/library.tsx
web/src/components/track-row.tsx
web/src/api/client.ts
web/src/audio/engine.ts
web/src/store/player.ts
Makefile                            (build web → embed → go build)
```

## Files to Modify

- `go.mod` → add chi, modernc.org/sqlite, yaml.v3
- `cmd/musiccli/cmd/root.go` → register `serve` subcommand
- `README.md` → add `serve` quick start
- `.gitignore` → add `web/node_modules/`, `web/dist/` (TBD commit policy)

## Implementation Steps

1. **DB schema + migrations** (`internal/library/migrations.go`):
   - Embed SQL via `//go:embed sql/*.sql`
   - Apply on store open
2. **Scanner** (`internal/library/scanner.go`):
   - Accept `root string, store *Store`
   - `WalkDir` → filter `.flac`/`.wav` → open tag reader → upsert track (plus artist/album)
   - Incremental: if `mtime == stored.mtime` skip
   - Expose `ScanAsync(ctx) <-chan Progress`
3. **Store** (`internal/library/store.go`):
   - Methods: `UpsertTrack`, `ListTracks(limit, offset, sort)`, `GetTrack(id)`, `ListAlbums`, etc.
   - Use `database/sql` + `modernc.org/sqlite` driver
   - Enable WAL: `PRAGMA journal_mode=WAL`
4. **Config** (`internal/config/config.go`):
   - Load `~/.config/musiccli/config.yaml` (fall back to CLI flags)
   - Fields: `music_dir`, `db_path`, `port`, `bind_addr`
5. **Server** (`internal/web/server.go`):
   - chi router, middleware: logger, recoverer, CORS (localhost)
   - Mount `/api/library` group
   - Mount `/stream/:id`
   - Mount `spa_embed` at `/*` as SPA fallback (catch-all → index.html)
6. **Library handlers** (`handlers/library.go`):
   - `GET /api/library/tracks` → paginated
   - `GET /api/library/albums` + `/albums/:id/tracks`
   - `GET /api/library/artists`
7. **Stream handler** (`handlers/stream.go`):
   - Lookup track by id → `os.Open` → detect MIME from format col → `http.ServeContent` (handles Range)
8. **Scan endpoints**:
   - `POST /api/scan` → kick off async scan
   - `GET /api/scan/status` → progress state
9. **SPA skeleton** (`web/`):
   - Vite+React+TS init
   - `api/client.ts` fetch wrapper
   - `audio/engine.ts` wraps `<audio>`: play/pause/seek/onEnded
   - `store/player.ts` Zustand: `currentTrack`, `isPlaying`
   - `pages/library.tsx` fetch `/api/library/tracks`, list, click → `engine.play('/stream/:id')`
10. **Build integration** (`Makefile`):
    - `make web` → `cd web && npm run build`
    - `make build` → `make web && go build -o bin/musiccli ./cmd/musiccli`
    - `make dev` → run Vite dev server + Go server w/ air (optional)
11. **Embed SPA** (`internal/web/spa_embed.go`):
    - `//go:embed all:web/dist`
    - `http.FS` wrapper with index.html fallback for SPA routes

## Acceptance Criteria

- [x] `make build && ./bin/musiccli serve --music-dir ~/Music` boots
- [x] Browser on `localhost:8080` shows track list
- [x] Click track → audio plays in Chrome
- [x] `curl -I -H "Range: bytes=0-1023" localhost:8080/stream/1` returns 206
- [x] Second run skips unchanged files (mtime check)
- [x] `go build` produces single binary, SPA embedded

## Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| SQLite driver not pure-Go | Break single-binary promise | Use `modernc.org/sqlite` explicitly |
| dhowden/tag fails on some FLAC | Track missing | Log + skip, continue scan |
| Large file scan blocks | UX bad | Async scan + `/api/scan/status` polling |
| SPA routing conflicts with API | 404 on deep links | `/api/*` precedence, catch-all → index.html |

## Security Considerations

- Bind default `127.0.0.1` unless `--bind 0.0.0.0` (prevent accidental WAN exposure)
- Validate track `id` is integer (no path traversal via stream handler)
- File reads restricted to music_dir (resolve + `filepath.Rel` check)

## Testing

- Unit: scanner on fixture dir, store upsert/list
- Integration: scan → query → stream round-trip via `httptest`
- Manual: play FLAC in Chrome, seek, skip

## Next Steps

Phase 2 depends on this. No handoff blockers; `api/client.ts` contract locked.
