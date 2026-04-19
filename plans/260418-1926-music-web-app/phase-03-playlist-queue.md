---
phase: 3
name: Playlists + Queue drawer
status: pending
priority: high
effort: 2-3 days
blockedBy: [phase-02]
---

# Phase 3: Playlists + Queue

## Overview

Add CRUD playlists, track-to-playlist add/remove, reorder. Queue drawer shows upcoming tracks with drag-reorder.

## Key Insights

- Playlists shared globally (no users) → no auth-in-request concerns
- Queue = client-side (localStorage persisted) per brainstorm decision
- Playlist reorder = server `PUT /reorder` with full ordered ID list (simplest)
- Drag-drop: `dnd-kit` (React) for both playlist and queue

## Requirements

### Functional

- [ ] Create/rename/delete playlists
- [ ] Add track to playlist from any track-row context menu
- [ ] Remove track from playlist
- [ ] Reorder tracks within playlist (drag handle)
- [ ] Sidebar lists all playlists, click → playlist detail
- [ ] Queue drawer: see upcoming, remove, reorder, play-next action
- [ ] Queue persisted in localStorage (survives refresh)

### Non-Functional

- [ ] Playlist CRUD round-trip <100ms
- [ ] Drag-drop feels native (<16ms/frame)
- [ ] Queue drawer opens in <100ms

## Architecture

```
PlaylistSidebar ──(list)──► /api/playlists
PlaylistDetail ──(tracks)─► /api/playlists/:id
  on reorder ─────────────► PUT /api/playlists/:id/reorder

QueueDrawer ── zustand(player.queue) ── (localStorage sync middleware)
```

## Files to Create

```
internal/web/handlers/playlist.go
web/src/pages/playlist-detail.tsx
web/src/components/playlist-sidebar.tsx
web/src/components/queue-drawer.tsx
web/src/components/track-context-menu.tsx
web/src/components/add-to-playlist-dialog.tsx
web/src/hooks/use-playlists.ts
```

## Files to Modify

- `internal/library/store.go` → playlist CRUD methods
- `internal/web/server.go` → mount `/api/playlists`
- `internal/web/dto.go` → Playlist, PlaylistTrack DTOs
- `web/package.json` → add `@dnd-kit/core`, `@dnd-kit/sortable`
- `web/src/store/player.ts` → localStorage persist middleware for queue
- `web/src/App.tsx` → add sidebar + queue drawer

## Implementation Steps

1. **Store methods**:
   ```go
   CreatePlaylist(name) (int64, error)
   RenamePlaylist(id, name)
   DeletePlaylist(id)
   ListPlaylists() []Playlist
   GetPlaylist(id) (Playlist, []Track)
   AddTrackToPlaylist(pid, tid, position int)
   RemoveTrackFromPlaylist(pid, tid)
   ReorderPlaylist(pid int64, orderedTrackIDs []int64)  // transaction
   ```
2. **Handlers** (`playlist.go`): map to REST per plan.md API spec
3. **Sidebar**: fetch `/api/playlists` on mount, invalidate on mutation
4. **Detail page**: list tracks, drag handle per row, `dnd-kit/sortable`
5. **Add-to-playlist**: modal on track context menu, lists playlists, confirm adds via POST
6. **Queue drawer**:
   - Right-side drawer toggled from NowPlayingBar icon
   - Shows `player.queue.slice(currentIndex + 1)` plus "Now Playing"
   - Per-row: drag, remove, play-now
7. **LocalStorage persist**:
   - Zustand middleware `persist(partialize: state => ({queue, currentIndex, volume}))`

## Acceptance Criteria

- [ ] Create playlist "Chill" → appears in sidebar
- [ ] Add 5 tracks → visible in detail
- [ ] Drag track to top → order persists after refresh
- [ ] Delete playlist → removed, tracks not deleted
- [ ] Queue drawer shows upcoming, drag reorders, remove drops track
- [ ] Refresh browser → queue state restored from localStorage

## Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Reorder race (two tabs) | Last-write-wins ok for family | Accept; no CRDT |
| localStorage quota on huge queue | Queue >5MB unlikely | Cap queue at 1000 |
| Orphaned `playlist_tracks` on track delete | N/A — tracks only removed via rescan cleanup | FK ON DELETE CASCADE |

## Testing

- Unit: reorder transaction, list playlists
- Integration: full playlist CRUD via httptest
- Manual: drag, multi-tab, refresh persistence

## Next Steps

Phase 4 polish only; no blockers.
