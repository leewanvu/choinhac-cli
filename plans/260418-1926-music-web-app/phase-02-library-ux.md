---
phase: 2
name: Library UX — Search, detail views, NowPlayingBar
status: pending
priority: high
effort: 3-4 days
blockedBy: [phase-01]
---

# Phase 2: Library UX

## Overview

Ship Spotify-like browsing: search, album detail, artist detail, persistent NowPlayingBar with full controls.

## Key Insights

- Search = `LIKE '%q%'` on indexed cols (title/artist/album). Trigram overkill for <10k tracks.
- NowPlayingBar = always-mounted component reading Zustand player state
- Progress/seek = `<audio>` `timeupdate` event → store → bar

## Requirements

### Functional

- [ ] `/api/search?q=` returns matches across tracks/albums/artists
- [ ] Album detail page: header + track list
- [ ] Artist detail page: albums grid + tracks
- [ ] NowPlayingBar: play/pause, prev/next, seek slider, volume, track meta
- [ ] Next/prev cycles through current context (album queue, search result, etc.)

### Non-Functional

- [ ] Search response <100ms for 10k tracks
- [ ] Bar update <50ms lag on seek
- [ ] No layout shift when bar appears

## Architecture

```
Search → /api/search?q  → SearchResults view
AlbumDetail → /api/library/albums/:id/tracks → AlbumView
NowPlayingBar ← zustand(player) ← audio/engine (timeupdate events)
```

## Files to Create

```
internal/web/handlers/search.go
web/src/pages/search.tsx
web/src/pages/album-detail.tsx
web/src/pages/artist-detail.tsx
web/src/components/now-playing-bar.tsx
web/src/components/seek-slider.tsx
web/src/components/volume-control.tsx
web/src/components/album-card.tsx
```

## Files to Modify

- `internal/library/store.go` → add `SearchTracks`, `SearchAlbums`, `SearchArtists`, `GetAlbum`, `GetArtist`
- `internal/web/server.go` → mount `/api/search`
- `web/src/App.tsx` → add React Router, mount NowPlayingBar globally
- `web/src/audio/engine.ts` → expose `timeupdate`, `ended` events
- `web/src/store/player.ts` → add `queue[]`, `currentIndex`, `volume`, `progress`, `duration`

## Implementation Steps

1. **Search backend**:
   - `SearchTracks(q, limit)` → `WHERE title LIKE ? OR artist.name LIKE ?`
   - Combined endpoint returns `{tracks: [], albums: [], artists: []}`
2. **Detail endpoints**:
   - `GET /api/library/albums/:id` → album + cover + tracks
   - `GET /api/library/artists/:id` → artist + albums
3. **Routing** (React Router v6):
   - `/` library, `/search?q=`, `/album/:id`, `/artist/:id`
4. **NowPlayingBar**:
   - Fixed bottom, reads Zustand, dispatches play/pause/seek/next/prev
   - Seek slider: debounced, `<audio>.currentTime = ...`
   - Volume: 0-100 mapped to `<audio>.volume = v/100`
5. **Queue context**:
   - When user plays from a list (album/search/playlist), store sets `queue = tracks[]`, `currentIndex`
   - `next()` increments, `prev()` decrements, loops optional
6. **Keyboard** (minimal here, polish in Phase 4):
   - Space=play/pause when not in input

## Acceptance Criteria

- [ ] Search "abc" returns matches <100ms
- [ ] Album page plays all tracks, next/prev works
- [ ] Bar persists across route changes
- [ ] Seek to middle of FLAC works in Chrome
- [ ] Volume control affects output
- [ ] ended event auto-advances queue

## Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Seek jank on large FLAC | UX | Relies on browser impl; accepted |
| LIKE search slow on 100k+ tracks | N/A for family scale | Defer FTS5 |
| Layout shift from bar | Visual bug | Reserve 80px bottom padding on root layout |

## Testing

- Unit: search store queries
- Integration: search handler returns expected shape
- Manual: seek/volume/next/prev flow

## Next Steps

Phase 3 depends on queue infrastructure (`store/player.ts`) landed here.
