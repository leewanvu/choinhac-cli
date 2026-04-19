# Project Roadmap

**Project:** musiccli (choinhaccli)  
**Current Phase:** MVP Phase 1 (Complete), Phase 2 Planning  
**Last Updated:** 2026-04-19

## Roadmap Overview

```
Phase 1 (MVP) ───── Phase 2 (Enhancement) ───── Phase 3 (Advanced)
  ✓ Complete       [ ] In Planning               [ ] Planned for future
```

---

## Phase 1: MVP (CURRENT — Complete ✓)

**Status:** Complete ✓  
**Timeline:** Completed  
**Focus:** Core audio playback + AI appreciation + Enhanced UI/UX

### Completed Features

- [x] **Audio Playback Engine**
  - FLAC/WAV format support (gopxl/beep)
  - Real-time streaming with resampling
  - Play, pause, resume, stop controls
  - Next/prev/random navigation
  - Volume control (±20dB range)
  - Lock-free amplitude tracking (atomic operations)

- [x] **Metadata Extraction & Album Art**
  - ID3 v2.3 tag parsing + cover art extraction
  - FLAC vorbis comment + embedded art parsing
  - dhowden/tag integration
  - Display: artist, album, title, sample rate, duration, album art

- [x] **Enhanced Terminal User Interface**
  - BubbleTea + Lipgloss + Catppuccin Mocha theme
  - 2-column layout: album art (left) + track info (right)
  - 24-bar real-time frequency visualizer
  - Unicode half-block album art rendering (JPEG/PNG)
  - Smooth gradient progress bar with sub-block characters
  - Playlist view (7-track window, current track highlighted)
  - Control hints and status display
  - 100ms refresh rate with visual decay effects

- [x] **Playlist Management**
  - Single file or directory of audio files
  - Auto-sort by filename
  - Seamless track switching
  - Wraparound navigation

- [x] **AI Music Appreciation (`feel` command)**
  - Python FastAPI analyzer service
  - librosa DSP feature extraction
  - Multi-provider LLM support (OpenAI, Gemini, Claude, OpenRouter)
  - System + user prompt engineering
  - Vietnamese and English output
  - Structured audio feature analysis

- [x] **CLI Framework**
  - Cobra-based command routing
  - `musiccli play <path>` subcommand (TUI playback)
  - `musiccli feel <audio_file>` subcommand (AI appreciation)
  - `musiccli serve` subcommand (web server)
  - Graceful error handling
  - Friendly user feedback

- [x] **Web Server & Music Library Management** (Phase 1 Extension)
  - SQLite library with artist/album/track/playlist tables
  - Async directory scanner with mtime-based incremental updates
  - Cover art extraction and caching
  - REST API for library queries (paginated tracks, search, filter)
  - HTTP Range support for audio streaming
  - Favicon, album cover serving with cache headers

- [x] **Web UI (Spotify-like React SPA)**
  - Vite 6 + React 18 + TypeScript + Zustand
  - Responsive layout: static sidebar (≥768px), hamburger mobile
  - Library page with virtualized track list (5000+ tracks)
  - Search across tracks, albums, artists
  - Album/Artist detail pages
  - Playlist management (CRUD, drag-to-reorder)
  - Now playing bar with cover, seek bar, volume controls
  - Queue drawer and add-to-playlist dialog
  - Keyboard shortcuts (space, arrows, M for mute)
  - localStorage persistence for player state

### MVP Success Criteria

| Criterion | Status | Notes |
|-----------|--------|-------|
| Play FLAC/WAV files | ✓ | Tested on macOS |
| TUI renders correctly | ✓ | Responsive to 100ms tick |
| Audio quality (bit-perfect) | ✓ | No conversion loss |
| CLI subcommands work | ✓ | play + feel functional |
| Metadata displays correctly | ✓ | ID3/FLAC tag parsing |
| AI review generation works | ✓ | All 4 providers tested |
| No segfaults on shutdown | ✓ | Graceful Ctrl+C handling |

---

## Phase 2: Enhancement (PLANNED)

**Estimated Timeline:** Q3-Q4 2026  
**Focus:** Testing, polish, new features, configuration

### High Priority

#### 2.1 Test Suite
**Status:** Not started  
**Effort:** 2-3 weeks  
**Acceptance Criteria:**
- [ ] Unit tests for `internal/audio/` (Player methods)
- [ ] Unit tests for `internal/agent/` (Agent.Feel, prompts)
- [ ] Unit tests for `internal/analyzer/` (HTTP client with mocks)
- [ ] Integration tests for `cmd/musiccli/cmd/` (CLI arg parsing)
- [ ] Overall coverage: 70%
- [ ] All tests pass on CI/CD (GitHub Actions)
- [ ] No flaky tests (run 5x consistently)

**Tasks:**
- [ ] Add testing framework (stdlib `testing` or testify)
- [ ] Create test fixtures (sample FLAC/WAV files)
- [ ] Mock HTTP server for analyzer tests
- [ ] Mock LLM API responses for agent tests
- [ ] Set up code coverage reporting
- [ ] Add GitHub Actions workflow

#### 2.2 MP3 Support
**Status:** Not started  
**Effort:** 1-2 weeks  
**Acceptance Criteria:**
- [ ] Load and play .mp3 files
- [ ] Metadata extraction (ID3 tags)
- [ ] Equivalent audio quality to FLAC/WAV
- [ ] No additional dependencies in go.mod

**Challenges:**
- gopxl/beep has no native MP3 decoder
- Would require CGO binding (libmpg123 or mpg123)
- Cross-platform compilation complexity
- macOS: homebrew formula; Linux: system package; Windows: build from source

**Alternative Approach:**
- Use pure Go MP3 decoder (go-mp3, minimp3-go)
- Trade-off: performance vs. simplicity

#### 2.3 Audio Seeking
**Status:** Not started  
**Effort:** 1-2 weeks  
**Acceptance Criteria:**
- [ ] Click on progress bar to seek
- [ ] Keyboard shortcuts: arrows for ±5s seek
- [ ] Display time scrubbing feedback
- [ ] Smooth seeking without artifacts

**Implementation:**
- Implement `beep.StreamSeekCloser.Seek()` for both FLAC and WAV decoders
- Add mouse click detection in TUI (BubbleTea mouse support)
- Update progress bar rendering to show seekable region

#### 2.4 Configuration File
**Status:** Not started  
**Effort:** 1 week  
**Acceptance Criteria:**
- [ ] Read from `~/.config/musiccli/config.yaml` (XDG standard)
- [ ] Support settings: default provider, language, analyzer URL, volume limit
- [ ] CLI flags override config file
- [ ] Generate default config on first run
- [ ] Clear documentation on config options

**Config Format:**
```yaml
# ~/.config/musiccli/config.yaml
audio:
  volume_limit: 100    # Max volume in dB
  auto_normalize: false

feel:
  provider: openrouter  # openai, gemini, claude, openrouter
  model: meta-llama/llama-2-70b
  language: vi         # vi or en
  analyzer_url: http://localhost:8000

ui:
  playlist_window_size: 7
  refresh_rate_ms: 100
```

### Medium Priority

#### 2.5 Playback History & Stats
**Status:** Not started  
**Effort:** 1-2 weeks  
**Acceptance Criteria:**
- [ ] Track play count per file
- [ ] Store in `~/.config/musiccli/history.json`
- [ ] Display stats: most played, recently played, total playtime
- [ ] `musiccli stats` subcommand

#### 2.6 Keyboard Shortcuts Customization
**Status:** Not started  
**Effort:** 1 week  
**Acceptance Criteria:**
- [ ] User-defined key bindings in config
- [ ] Fallback to defaults if not specified
- [ ] Display custom shortcuts in help text

#### 2.7 Windows Support (Testing)
**Status:** Not started  
**Effort:** 1-2 weeks  
**Acceptance Criteria:**
- [ ] Build and test on Windows 10/11
- [ ] Speaker initialization works
- [ ] Audio output functional
- [ ] TUI renders correctly
- [ ] Document platform-specific notes

#### 2.8 Gapless Playback
**Status:** Not started  
**Effort:** 2-3 weeks (requires pre-loading + buffer management)  
**Acceptance Criteria:**
- [ ] Reduce gap between tracks from 200-400ms to <20ms
- [ ] Pre-load next track while current plays
- [ ] Seamless stream concatenation
- [ ] No audio pops/clicks

**Approach:**
- Implement double-buffering in Player
- Load next track header while current streams
- Queue both streamers in beep.Ctrl

---

## Phase 3: Advanced Features (FUTURE)

**Estimated Timeline:** 2027+  
**Focus:** Advanced audio, ecosystem integration, streaming

### Features

#### 3.1 Audio Output Device Selection
**Status:** Not started  
**Effort:** 1-2 weeks  
**Acceptance Criteria:**
- [ ] Enumerate available output devices
- [ ] User selection via `--device` flag or config
- [ ] Fallback to system default
- [ ] List devices: `musiccli devices` subcommand

**Complexity:** Platform-specific (CoreAudio, ALSA, DirectSound)

#### 3.2 Equalizer / Audio Filters
**Status:** Not started  
**Effort:** 2-3 weeks  
**Acceptance Criteria:**
- [ ] Expose beep/effects filters (bass, treble, reverb, etc.)
- [ ] User-adjustable via keyboard/config
- [ ] Preset profiles (normal, bass boost, clarity, etc.)
- [ ] Real-time adjustment (no restart needed)

#### 3.3 Streaming Source Support (HTTP/S3)
**Status:** Not started  
**Effort:** 2-3 weeks  
**Acceptance Criteria:**
- [ ] Load and play from HTTP URLs
- [ ] S3 bucket support (with credentials)
- [ ] Buffering strategy (circular buffer)
- [ ] Resume on network interruption
- [ ] `musiccli play https://example.com/song.flac`

#### 3.4 Visualization / Spectrum Display
**Status:** Not started  
**Effort:** 3-4 weeks  
**Acceptance Criteria:**
- [ ] Real-time FFT-based spectrum analyzer in TUI
- [ ] Animated bars or waveform display
- [ ] Toggle via keyboard shortcut
- [ ] Configurable update rate
- [ ] No performance degradation

**Challenges:** FFT computation, 60+ FPS rendering, cross-platform terminal graphics

#### 3.5 Remote Control / Headless Mode
**Status:** Not started  
**Effort:** 2-3 weeks  
**Acceptance Criteria:**
- [ ] HTTP API for playback control (play, pause, next, seek)
- [ ] Network socket-based communication
- [ ] `musiccli daemon` — background service
- [ ] `musiccli remote <cmd>` — send commands to daemon
- [ ] Web dashboard (optional Vue/React SPA)

#### 3.6 Caching & Analysis Results
**Status:** Not started  
**Effort:** 1 week  
**Acceptance Criteria:**
- [ ] Store analyzer results locally (per file hash)
- [ ] Skip analysis on second run (unless forced)
- [ ] Cache location: `~/.cache/musiccli/features/`
- [ ] `musiccli cache clear` — cleanup old entries

#### 3.7 Playlist Management
**Status:** Not started  
**Effort:** 1-2 weeks  
**Acceptance Criteria:**
- [ ] Save/load playlists (.m3u format)
- [ ] `musiccli playlist save <name> <file>`
- [ ] `musiccli playlist load <name>`
- [ ] Playlist editor (reorder, remove tracks)

#### 3.8 Tagging / Smart Collections
**Status:** Not started  
**Effort:** 2-3 weeks  
**Acceptance Criteria:**
- [ ] User-defined tags/labels per track
- [ ] Filtered playlists by tag
- [ ] `musiccli search <query>` — full-text search
- [ ] Smart playlists (e.g., "Recently Played > 2 weeks")

---

## Dependency Upgrades & Maintenance

### Go Module Updates
- Keep `gopxl/beep`, `bubbletea`, `lipgloss`, `cobra` current (check quarterly)
- No major breaking changes expected in next 12 months

### Python Service
- `librosa` — Stable; monitor for DSP improvements
- `fastapi` — Active development; stay within MAJOR version
- `numpy` — Stable; required by librosa

### CI/CD Setup
- [ ] GitHub Actions: build on push (Linux, macOS, Windows)
- [ ] Automated testing: go test -v ./...
- [ ] Code coverage reporting: codecov or codacy
- [ ] Release workflow: tag → build → publish releases

---

## Known Issues & Technical Debt

### Current MVP

| Issue | Severity | Status | Notes |
|-------|----------|--------|-------|
| No seeking support | High | Not started | Requires beep.Streamer changes |
| 200-400ms track gap | Medium | Acceptable for MVP | Gapless is Phase 2 candidate |
| Windows untested | Low | Not started | Platform-specific speaker init |
| No config file | Low | Acceptable for MVP | Phase 2 candidate |
| Zero test coverage | High | Not started | Priority for Phase 2 |
| No CI/CD | Medium | Not started | Phase 2 candidate |

### Future Considerations

1. **Performance at Scale:**
   - Test with 10k+ song playlists
   - Optimize metadata loading (lazy load if needed)

2. **Memory Leaks:**
   - Audit goroutine lifecycle (especially in playlist switching)
   - Profile with pprof

3. **Cross-Platform Audio:**
   - Test Windows speaker init thoroughly
   - Document platform-specific quirks

4. **Python Service Stability:**
   - Add timeouts for librosa operations
   - Handle malformed audio files gracefully

---

## Success Metrics by Phase

### Phase 1 (MVP) — ACHIEVED
- ✓ Play FLAC/WAV without crashes
- ✓ TUI responsive and intuitive
- ✓ Audio quality verified (bit-perfect)
- ✓ CLI usable without documentation
- ✓ AI appreciation working for all 4 providers

### Phase 2 (Enhancement)
- [ ] 70% test coverage
- [ ] MP3 support working
- [ ] Seeking functional
- [ ] Config file used by 100% of installations
- [ ] Zero high-severity issues
- [ ] Windows builds passing CI

### Phase 3 (Advanced)
- [ ] Streaming sources working (HTTP/S3)
- [ ] Visualization rendering at 60+ FPS
- [ ] Remote control API stable
- [ ] 1000+ concurrent connections (if daemon mode added)

---

## Community & Contributions

### Target Contributors
- Go developers interested in audio
- Terminal UI enthusiasts
- Music data analysis researchers
- Audio plugin developers

### Contribution Areas
- [ ] Additional LLM providers (HuggingFace, Azure)
- [ ] New DSP features (tempogram, chromagram, etc.)
- [ ] Visualization improvements
- [ ] Platform-specific optimizations
- [ ] Documentation translations

### License
- TBD (recommend MIT or Apache 2.0)

---

## References

- [gopxl/beep documentation](https://github.com/gopxl/beep)
- [BubbleTea guide](https://github.com/charmbracelet/bubbletea)
- [librosa tutorial](https://librosa.org/)
- [Cobra CLI framework](https://cobra.dev/)

