---
phase: 4
name: Polish — cover art, shortcuts, responsive
status: pending
priority: medium
effort: 2-3 days
blockedBy: [phase-03]
---

# Phase 4: Polish

## Overview

Ship Spotify-like polish: embedded album art, directory fallback, keyboard shortcuts, responsive layout for phones/tablets.

## Requirements

### Functional

- [ ] Embedded cover art extracted to `~/.musiccli/covers/:album_id.jpg` on scan
- [ ] Directory-level fallback (`cover.jpg`, `folder.jpg`, `album.jpg`)
- [ ] `/cover/:album_id` serves cached image with long cache header
- [ ] Album cards and NowPlayingBar show artwork
- [ ] Keyboard shortcuts: Space, ←/→ (prev/next), ↑/↓ (volume), /, m (mute)
- [ ] Responsive: mobile <768px collapses sidebar, bar adapts
- [ ] Loading states: skeletons on library load, scanning indicator
- [ ] Empty states (no tracks, no playlists, no results)

### Non-Functional

- [ ] Cover img lazy-loaded
- [ ] 60fps scrolling on 1000-row list (virtualization via `react-window`)

## Files to Create

```
web/src/components/cover-image.tsx
web/src/components/skeleton-track-list.tsx
web/src/components/empty-state.tsx
web/src/hooks/use-keyboard-shortcuts.ts
web/src/styles/responsive.css
internal/library/cover-extractor.go
internal/web/handlers/cover.go
```

## Files to Modify

- `internal/library/scanner.go` → call `ExtractCover` per album
- `internal/library/store.go` → store `cover_path` on album
- `internal/web/server.go` → mount `/cover/:album_id`
- `web/src/App.tsx` → responsive breakpoints, mobile drawer
- `web/src/components/now-playing-bar.tsx` → cover thumbnail

## Implementation Steps

1. **Cover extraction**:
   - On first track of album, read embedded art (`dhowden/tag` `.Picture()`)
   - Else look for `cover.jpg`/`folder.jpg`/`album.jpg` in track's dir
   - Resize to 512x512 max via `disintegration/imaging` (cheap, pure Go)
   - Save to `~/.musiccli/covers/{album_id}.jpg`
2. **Cover handler**:
   - `GET /cover/:album_id` → serve file, `Cache-Control: max-age=604800`
   - Placeholder SVG if missing
3. **Keyboard shortcuts**:
   - `useKeyboardShortcuts` hook attaches `keydown` listener on window
   - Skip when `document.activeElement` is input/textarea
4. **Virtualization**:
   - `react-window` FixedSizeList for track lists >100 items
5. **Responsive**:
   - Tailwind (optional dep) OR hand-rolled CSS with CSS grid
   - Sidebar becomes hamburger <768px
   - Bar stacks controls on small screens
6. **Loading/empty states**:
   - Skeletons match final layout
   - Empty scan → "Point `--music-dir` at your library"

## Acceptance Criteria

- [ ] Album with embedded art shows cover
- [ ] Album with no embedded art but `cover.jpg` shows cover
- [ ] Kbd: Space toggles play, arrows work, ignored when typing
- [ ] Mobile view usable (iPhone Chrome tested)
- [ ] 5000-row list scrolls at 60fps
- [ ] Scanning indicator visible during rescan

## Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Image resize CPU cost | Slow initial scan | Do in goroutine pool, limit concurrency |
| Cover path traversal | Security | Use `album_id` int, not path; store under fixed dir |
| Tailwind adds JS bundle | Bundle size | Use only utilities; or stick with CSS modules |

## Testing

- Unit: cover extractor fallback chain
- Manual: various FLAC files, phone browser, keyboard in all views

## Post-Phase

- Update `docs/codebase-summary.md` with new packages
- Update `docs/system-architecture.md` with web diagram
- Update `README.md` with `serve` command + screenshots
- Tag `v2.0.0` release

## Definition of Done (Whole Project)

- All 4 phases merged
- `make build` produces single binary
- `scp bin/musiccli host:` + run works on fresh Linux box
- Family member successfully browses + plays from their phone on home WiFi
