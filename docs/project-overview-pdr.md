# Project Overview & Product Development Requirements

**Project Name:** musiccli  
**Module Name:** `choinhaccli`  
**Binary Name:** `musiccli`  
**Language:** Go 1.25 + Python 3.10+  
**Last Updated:** 2026-04-18

## Vision

A minimalist, high-fidelity CLI music player with AI-powered music appreciation. Designed for developers and audiophiles who prefer keyboard-driven interfaces and don't compromise on audio quality.

## Core Features

### 1. High-Fidelity Audio Playback
- **Engine:** gopxl/beep (real-time resampling, bit-perfect delivery)
- **Formats:** FLAC, WAV (24-bit/96kHz+ supported)
- **Control:** Play, pause, next, previous, random, volume adjustment
- **Metadata:** ID3/FLAC tag extraction (artist, album, title, sample rate)

### 2. Terminal User Interface (TUI)
- **Framework:** BubbleTea + Lipgloss
- **Theme:** Catppuccin Mocha color palette
- **Display:** Album art (JPEG/PNG), 24-bar visualizer, progress bar, metadata, playlist, status
- **Layout:** 2-column design with album art on left, track info on right
- **Visual FX:** Unicode half-block art rendering, animated frequency bars with decay, smooth gradient progress
- **Polling:** 100ms refresh for responsive UI with real-time amplitude tracking
- **Controls:** Keyboard-driven (space, arrow keys, +/-, q)

### 3. AI Music Appreciation (`feel` command)
- **Pipeline:** Audio DSP analysis → LLM review generation
- **DSP Engine:** Python FastAPI service + librosa (BPM, key, spectral features, mood keywords)
- **LLM Providers:** OpenAI, Google Gemini, Anthropic Claude, OpenRouter
- **Output Languages:** Vietnamese (default), English
- **Response:** 2-3 paragraph emotional/technical music review

### 4. Playlist Management
- **Source:** Single file or directory of .flac/.wav files
- **Navigation:** Next/prev/random with wraparound
- **Display:** 7-track window with current position indicator

## Technical Constraints

| Aspect | Constraint |
|--------|-----------|
| **Audio Format Support** | FLAC, WAV only (no MP3 yet—would require cgo binding) |
| **Platform** | macOS, Linux (tested); Windows untested (speaker init differs) |
| **DSP Service** | Optional; required only for `feel` command |
| **LLM API Keys** | Provider-specific env vars (OPENAI_API_KEY, GEMINI_API_KEY, etc.) |
| **Sample Rate** | 44.1 kHz base; resampling handles up to 192 kHz |

## Codebase Metrics

| Metric | Value |
|--------|-------|
| **Go LOC** | ~1800 lines across 18 files |
| **Packages** | `audio`, `ui`, `agent`, `analyzer` |
| **Commands** | `play`, `feel` (cobra subcommands) |
| **Tests** | None yet (no `*_test.go` files) |
| **New Components** | amplitude_tracker, album_art, visualizer, view (refactored model) |
| **Imports** | 4 external (beep, bubbletea, lipgloss, dhowden/tag) |

## Functional Requirements (FRs)

### FR1: Audio Playback
- [x] Load and decode FLAC/WAV from file or directory
- [x] Stream to system speaker with beep.Speaker
- [x] Extract and display metadata (artist, album, title, sample rate, duration)
- [x] Play, pause, stop, next, prev, random navigation
- [x] Volume control (gain adjustment via beep/effects.Volume)
- [x] Handle end-of-track → auto-play next in playlist

### FR2: TUI Display & Visualizations
- [x] Show metadata (artist, album, track), playback status, volume offset in 2-column layout
- [x] Display album art (JPEG/PNG) rendered as Unicode half-block characters
- [x] Real-time progress bar with smooth gradient fill and sub-block precision
- [x] 24-bar frequency visualizer with lock-free amplitude input and exponential decay
- [x] Playlist window (current track highlighted, ±3 tracks visible with track numbers)
- [x] Control hints (space, arrow keys, +/-, r, q)
- [x] Catppuccin Mocha color theme for all UI elements

### FR3: AI Music Appreciation
- [x] Invoke Python analyzer service (POST /analyze with file path)
- [x] Parse DSP features: BPM, key, spectral centroid, MFCCs, mood keywords, energy profile
- [x] Select LLM provider (OpenAI, Gemini, Claude, OpenRouter)
- [x] Generate system prompt (Vietnamese or English) with mood/feature context
- [x] Display formatted review with header, metadata, features, and review text

### FR4: CLI Entrypoint
- [x] Cobra-based root command with `play` and `feel` subcommands
- [x] Argument validation (file path required, format checking)
- [x] Graceful shutdown on Ctrl+C

## Non-Functional Requirements (NFRs)

| Requirement | Target | Status |
|-------------|--------|--------|
| **Startup Latency** | <500ms (before music plays) | ✓ |
| **UI Responsiveness** | <100ms between keypress and response | ✓ |
| **Audio Quality** | Bit-perfect, no loss | ✓ |
| **Memory Usage** | <100MB for typical 3-5min track | ✓ |
| **Error Recovery** | Graceful exit with clear messages | ✓ |
| **Test Coverage** | Target: 70% (currently 0%) | ✗ |
| **Documentation** | API docs, code standards, architecture | In Progress |

## Acceptance Criteria

### Play Command
- [ ] Load single .flac or .wav file → play audio with TUI
- [ ] Load directory → build playlist, play first file, navigate with n/p/r
- [ ] Reject unsupported formats with error message
- [ ] Handle missing files gracefully
- [ ] Volume control: ±20dB range per adjustment
- [ ] Pause/resume without click artifacts
- [ ] Quit with Ctrl+C without segfault

### Feel Command
- [ ] Load .flac or .wav → call analyzer service
- [ ] Parse features → build user prompt with metadata + features
- [ ] Select LLM provider via --provider flag
- [ ] Generate 2-3 paragraph review in specified language (--lang)
- [ ] Display formatted output with emoji headers and color styling
- [ ] Exit gracefully if analyzer unreachable

### Code Quality
- [ ] No compile errors or warnings
- [ ] Consistent error handling (fmt.Errorf with context)
- [ ] Comments on exported functions and complex logic
- [ ] Follow Go conventions (CamelCase, doc comments)

## Known Limitations

1. **No MP3 Support:** Requires CGO binding for libmpg123 or similar (out of scope for MVP)
2. **No Seeking:** Current implementation plays from start; seek() not implemented in beep.Streamer
3. **No Config File:** All settings via CLI flags or env vars; no ~/.musicclirc yet
4. **No Gapless Playback:** Track switching has ~200-400ms gap between tracks
5. **No Output Device Selection:** Uses system default speaker
6. **Limited DSP Features:** librosa provides 13 main features; could expand to chromagram, tempogram, etc.
7. **No Test Suite:** Manual testing only; integration tests needed
8. **No CI/CD:** No GitHub Actions or automated testing pipeline yet

## Roadmap

### Phase 1: MVP (Current)
- ✓ Audio playback (FLAC, WAV)
- ✓ TUI with real-time controls
- ✓ AI music appreciation (`feel`)
- ✓ Multi-provider LLM support

### Phase 2: Enhancement
- [ ] Test suite (unit + integration)
- [ ] MP3 support (requires cgo)
- [ ] Audio seeking (progress bar click-to-seek)
- [ ] Config file support (~/.config/musiccli/config.yaml)
- [ ] Playback history and stats

### Phase 3: Advanced
- [ ] Gapless playback
- [ ] Audio output device selection
- [ ] Equalizer / audio filters
- [ ] Caching of analysis results
- [ ] Streaming source support (HTTP, S3)

## Security Considerations

1. **File Path Traversal:** Validate file paths before passing to analyzer service
2. **API Keys:** Use environment variables; never hardcode secrets
3. **HTTP Client:** 120s timeout on analyzer requests (DoS protection)
4. **Error Messages:** Avoid exposing absolute paths in user-facing output
5. **Analyzer Service:** No auth currently; assume same machine or trusted network

## Dependencies

### Go Modules
- `github.com/gopxl/beep` — audio decoding and playback
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — styling
- `github.com/spf13/cobra` — CLI command framework
- `github.com/dhowden/tag` — ID3/FLAC metadata extraction

### Python (Analyzer Service)
- `fastapi` — HTTP API framework
- `librosa` — DSP feature extraction
- `soundfile` — audio I/O
- `numpy` — numerical operations

## Success Metrics

| Metric | Target | Method |
|--------|--------|--------|
| **Time to play** | <500ms from CLI invocation | Manual stopwatch |
| **UI refresh latency** | <100ms | Event logging |
| **Audio quality** | Bit-perfect (no conversion loss) | Spectral analysis |
| **DSP accuracy** | BPM within ±2% of ground truth | Compare with Audacity |
| **LLM response quality** | Coherent, relevant reviews (subjective) | Manual review |
