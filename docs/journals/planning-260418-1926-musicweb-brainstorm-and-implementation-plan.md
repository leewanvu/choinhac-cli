# musicweb: Brainstorm + Implementation Plan Complete

**Date**: 2026-04-18 19:26
**Severity**: Info
**Component**: Project Planning / Architecture
**Status**: Completed

## What Happened

Completed full brainstorm and implementation plan for musicweb — a new Spotify-like web listening app built on top of musiccli. Generated 1 brainstorm report + plan directory with overview + 4 phase files. Zero code written; planning only.

## Key Architectural Decisions

**Zero coupling with beep engine**: Playback driven by browser-side HTML5 `<audio>` + HTTP Range streaming via `http.ServeContent`. Did NOT try to reuse internal/audio/beep for web context. Cleaner separation, avoids forcing browser tech into CLI audio layer.

**Pure Go stack**: Chi router backend, embedded React+Vite SPA, pure-Go SQLite (modernc.org/sqlite). Single binary option viable.

**LAN-first scope**: Self-hosted, <10 family users, local files only. No user auth. Client-side queue in localStorage. Global shared playlists (no user isolation).

## Scope Boundaries

- Chrome only (no FLAC transcode for Safari compatibility)
- New cobra subcommand: `musiccli serve`
- New modules: internal/web/, internal/library/, web/
- 4-phase rollout: Foundation → Library UX → Playlists+Queue → Polish

## Unresolved Questions

- Config format (YAML lean vs other)
- web/dist commit policy (git track or gitignore)
- Single vs separate binary strategy
- Cover art fallback paths
- Default listen port

## Artifacts

- Brainstorm report: `plans/reports/brainstorm-260418-1926-music-web-app.md`
- Plan directory: `plans/260418-1926-music-web-app/` with plan.md + phase files
- Next: Phase 1 implementation (Scanner + DB + API + SPA skeleton + streaming)
