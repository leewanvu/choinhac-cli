# Codebase Summary

**Module:** `choinhaccli`  
**Total Go LOC:** ~1800  
**Go Files:** 18  
**Python Files:** 1 (analyzer service)  
**Last Updated:** 2026-04-18

## Directory Structure

```
musiccli/
├── cmd/musiccli/
│   ├── main.go                    — Entry point (calls cmd.Execute())
│   └── cmd/
│       ├── root.go                — Cobra root command setup (2 subcommands)
│       ├── play.go                — `play` subcommand (TUI playback)
│       └── feel.go                — `feel` subcommand (AI appreciation)
├── internal/
│   ├── audio/
│   │   ├── player.go              — Audio engine (298 LOC)
│   │   └── amplitude_tracker.go    — Lock-free amplitude tracking (atomic)
│   ├── ui/
│   │   ├── model.go               — BubbleTea TUI model (~85 LOC, refactored)
│   │   ├── view.go                — TUI layout & rendering (~165 LOC)
│   │   ├── album_art.go           — Album art decoder & Unicode renderer
│   │   ├── visualizer.go          — 24-bar frequency visualizer
│   │   └── style.go               — Catppuccin Mocha palette (replaced)
│   ├── agent/
│   │   ├── agent.go               — Agent orchestrator (49 LOC)
│   │   ├── prompt.go              — Prompt builders (Vietnamese/English)
│   │   └── providers/
│   │       ├── openai.go          — OpenAI implementation
│   │       ├── gemini.go          — Google Gemini implementation
│   │       ├── claude.go          — Anthropic Claude implementation
│   │       └── openrouter.go      — OpenRouter proxy implementation
│   └── analyzer/
│       └── analyzer.go            — HTTP client for Python DSP service (93 LOC)
├── analyzer/
│   ├── main.py                    — FastAPI service (librosa DSP)
│   ├── requirements.txt            — Python dependencies
│   └── README.md                   — Analyzer setup guide
├── go.mod                          — Go module definition (choinhaccli)
└── README.md                        — Main project README
```

## Package Overview

### `cmd/musiccli` (Entry Point, ~287 LOC)

**Files:**
- `main.go` (7 LOC) — Minimal entry point; calls `cmd.Execute()`
- `cmd/root.go` (25 LOC) — Cobra root command; registers `play` and `feel` subcommands
- `cmd/play.go` (76 LOC) — `musiccli play <path>` handler
- `cmd/feel.go` (209 LOC) — `musiccli feel <audio_file>` handler with LLM integration

**Responsibilities:**
- CLI argument parsing and validation
- Subcommand routing (Cobra)
- File existence checks, format validation (.flac, .wav)
- Playlist building from directories

**Key Functions:**
- `Execute()` — Runs root Cobra command
- `runPlay()` — Play subcommand entry point; loads player, initializes speaker, runs TUI
- `runFeel()` — Feel subcommand entry point; extracts metadata, calls analyzer, generates review

---

### `internal/audio` (Audio Engine, ~300+ LOC)

**Files:**
- `player.go` — Audio playback engine (298 LOC)
- `amplitude_tracker.go` — Lock-free amplitude capture (atomic operations)

**Types:**

```go
type TrackMetadata struct {
    Title, Artist, Album string
    SampleRate int
    Duration time.Duration
    CoverArt []byte  // NEW: JPEG/PNG album art data
}

type State int  // StateStopped, StatePlaying, StatePaused

type Player struct {
    ctrl *beep.Ctrl
    volume *effects.Volume
    streamer beep.StreamSeekCloser
    format beep.Format
    state State
    Metadata TrackMetadata
    Playlist []string
    PlaylistIdx int
    tracker *amplitudeTracker  // NEW: amplitude measurement
    done chan bool
}

type amplitudeTracker struct {
    wrapped beep.Streamer
    peak atomic.Uint64  // Lock-free peak storage
}
```

**Key Functions:**

| Function | Signature | Purpose |
|----------|-----------|---------|
| `InitSpeaker()` | `() error` | Initialize global speaker (once only) |
| `NewPlayer()` | `() *Player` | Create player instance |
| `LoadPlaylist()` | `(paths []string, startIdx int) error` | Load playlist, play first track |
| `LoadAndPlay()` | `(path string) error` | Load, decode, and play audio file |
| `Play()` | `() error` | Resume playback (ctrl.Paused = false) |
| `Pause()` | `() error` | Pause playback (ctrl.Paused = true) |
| `Stop()` | `() error` | Stop and close streamer |
| `Next()`, `Prev()` | `() error` | Navigate playlist with wraparound |
| `Random()` | `() error` | Jump to random track in playlist |
| `VolumeUp()`, `VolumeDown()` | `() error` | Adjust volume (±1 dB per call) |
| `TogglePause()` | `() error` | Toggle between play/pause |
| `GetState()` | `() State` | Return current playback state |
| `GetPosition()` | `() time.Duration` | Get current playback position |
| `GetVolume()` | `() float64` | Get current volume offset (dB) |
| `Done()` | `() <-chan bool` | Return done channel (signals track end) |
| `GetAmplitude()` | `() float64` | Return current audio amplitude (0.0-1.0) |
| `extractMetadata()` | `(path string) *TrackMetadata` | Parse ID3/FLAC tags + cover art using dhowden/tag |

**Implementation Details:**
- Uses `gopxl/beep` for streaming, resampling, effects
- Metadata extracted via `dhowden/tag` package
- Volume control: `beep/effects.Volume` with gain adjustment
- Playlist wraparound: `(idx + 1) % len(playlist)`
- Random: `rand.Int() % len(playlist)`
- State transitions: Stopped → Playing / Paused (no Playing → Stopped without explicit Stop)

**Notes:**
- No seeking support (beep.StreamSeekCloser interface present but not implemented)
- ~200-400ms gap between tracks (beep speaker flush)
- Sample rate base: 44.1 kHz; resampling handles up to 192 kHz

---

### `internal/ui` (TUI Layer, ~250+ LOC)

**Files:**
- `model.go` (~85 LOC) — BubbleTea Model (refactored, View moved to view.go)
- `view.go` (~165 LOC) — TUI rendering & layout (NEW)
- `visualizer.go` — 24-bar frequency visualizer (NEW)
- `album_art.go` — Album art decoder & Unicode renderer (NEW)
- `style.go` — Catppuccin Mocha palette (refactored)

**BubbleTea Model:**

```go
type Model struct {
    player *audio.Player
    viz visualizer          // NEW: frequency visualization
    art string              // NEW: cached rendered album art
    lastTrackIdx int        // NEW: track change detection
    width int
    err error
}

type visualizer struct {
    bars [24]float64        // 24-bar display
    // ... decay & sensitivity parameters
}
```

**Key Methods:**

| Method | Signature | Purpose |
|--------|-----------|---------|
| `NewModel()` | `(p *audio.Player) Model` | Create TUI model |
| `Init()` | `() tea.Cmd` | Setup initial commands (tick, track finished listener) |
| `Update()` | `(msg tea.Msg) (tea.Model, tea.Cmd)` | Handle keyboard, tick, window size, errors |
| `View()` | `() string` | Render TUI (metadata, progress, playlist, help) |

**Control Bindings:**

| Key | Action |
|-----|--------|
| `space` | Play/pause |
| `n` / `→` | Next track |
| `p` / `←` | Previous track |
| `r` | Random track |
| `+` / `↑` | Volume up |
| `-` / `↓` | Volume down |
| `q` / `Ctrl+C` | Quit |

**Polling Strategy:**
- `tea.Tick()` every 100ms → triggers Update with `tickMsg`
- `waitForTrackFinished()` goroutine → sends `trackFinishedMsg` when track ends
- Auto-play next on track finish

**Display Layout:**
- Title: "🎵 CLI Music Player"
- Metadata: Artist, Album, Track [N/Total]
- Stats: Sample rate, playback status (Playing/Paused/Stopped), volume offset
- Progress: Bar (ASCII blocks) + elapsed/total time
- Playlist: Current track + ±3 neighbors (window scrolls)
- Help: Control hints

**Layout (New 2-Column Design):**
```
┌────────────────────────────────────┐
│  Album Art (20x10) │ Track Info    │  ← 2-column top panel
├────────────────────────────────────┤
│  24-Bar Visualizer (full width)    │  ← Live amplitude bars
├────────────────────────────────────┤
│  Progress Bar + Time | Status, Vol │
├────────────────────────────────────┤
│  Playlist (scrollable, 7 tracks)   │  ← Current track highlighted
├────────────────────────────────────┤
│  Help text                         │
└────────────────────────────────────┘
```

**Styling (Catppuccin Mocha):**
- Album art border: Rosewater (#F5E0DC)
- Info panel background: Surface0 (#313244)
- Track title: Subtext1 (#BAC2DE)
- Artist: Text (#CDD6F4)
- Visualizer bars: Lavender (#B4A7E5) → Flamingo (#F38BA8)
- Progress fill: Sapphire (#89B4FA) with gradient
- Current track highlight: Maroon (#EBA0AC)
- Playlist numbers: Overlay0 (#6C7086)

---

### `internal/agent` (AI Orchestration, ~150+ LOC)

**Files:**
- `agent.go` (49 LOC) — Agent orchestrator
- `prompt.go` — Prompt builders
- `providers/*.go` — LLM implementations (OpenAI, Gemini, Claude, OpenRouter)

**Types:**

```go
type LLMProvider interface {
    Name() string
    Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type Agent struct {
    provider LLMProvider
    lang string  // "vi" or "en"
}
```

**Key Functions:**

| Function | Signature | Purpose |
|----------|-----------|---------|
| `NewAgent()` | `(provider LLMProvider, lang string) *Agent` | Create agent with provider and language |
| `(a *Agent) Feel()` | `(ctx context.Context, features *AudioFeatures, metadata *TrackMetadata) (string, error)` | Generate music review |
| `buildSystemPrompt()` | `(lang string) string` | Build Vietnamese or English system prompt |
| `buildUserPrompt()` | `(features *AudioFeatures, metadata *TrackMetadata) string` | Format audio features into user prompt |

**Provider Interface:**
All providers must implement `LLMProvider`:
- `openai.OpenAIProvider` — Uses OpenAI API (GPT-4, GPT-3.5)
- `gemini.GeminiProvider` — Uses Google Gemini API
- `claude.ClaudeProvider` — Uses Anthropic Claude API
- `openrouter.OpenRouterProvider` — Proxy to multiple model endpoints

**Prompt Format:**

**System Prompt (Vietnamese):**
```
Bạn là một nhà phê bình âm nhạc tài năng và có cảm xúc.
Hãy viết một bài review ngắn (2-3 đoạn) về bản nhạc được mô tả dưới đây...
```

**User Prompt:**
```
Title: {title}
Artist: {artist}
Album: {album}
Duration: {duration}

Audio Features:
- BPM: {bpm}
- Key: {key}
- Mood Keywords: {keywords}
- Spectral Centroid: {centroid} Hz
...
```

**Response:** 2-3 paragraph emotional/technical review

---

### `internal/analyzer` (DSP Client, ~93 LOC)

**File:** `analyzer.go`

**Types:**

```go
type AudioFeatures struct {
    BPM float64
    Key string
    SpectralCentroidMean, SpectralBandwidthMean float64
    MFCCMeans []float64
    RMSEnergyMean, ZeroCrossingRateMean float64
    ChromaFeatures map[string]float64
    OnsetStrengthMean float64
    DurationSeconds float64
    EnergyProfile []float64
    MoodKeywords []string
}

type Client struct {
    baseURL string
    httpClient *http.Client
}
```

**Key Functions:**

| Function | Signature | Purpose |
|----------|-----------|---------|
| `NewClient()` | `(baseURL string) *Client` | Create client (default: http://localhost:8000) |
| `(c *Client) Analyze()` | `(filePath string) (*AudioFeatures, error)` | POST to /analyze, parse JSON response |
| `(c *Client) HealthCheck()` | `() error` | GET /health to verify service is up |

**HTTP Contract:**

**POST /analyze**
```
Form Data: path=<file_path>
Response: JSON(AudioFeatures)
```

**GET /health**
```
Response: 200 OK if running, error otherwise
```

**Timeout:** 120s (audio analysis can be slow)

---

### `analyzer/main.py` (DSP Service, ~330 LOC)

**Language:** Python 3.10+ (FastAPI + librosa)

**Dependencies:**
- `fastapi` — HTTP API framework
- `librosa` — DSP feature extraction
- `soundfile` — Audio I/O
- `numpy` — Numerical operations

**Key Functions:**

| Function | Purpose |
|----------|---------|
| `_detect_key(chroma)` | Estimate musical key (C major/minor, etc.) via Krumhansl-Kessler correlation |
| `_compute_mood_keywords()` | Derive mood tags from BPM, key, spectral features, RMS energy, ZCR |
| `extract_features(file_path)` | Load audio, compute librosa features (BPM, chroma, MFCC, spectral, etc.) |
| `POST /analyze` | Accept file path, extract features, return JSON |
| `GET /health` | Return 200 OK |

**Features Extracted:**
- **BPM:** Tempo via `librosa.beat.tempo()`
- **Key:** Musical key from chroma features (major/minor)
- **Spectral Centroid:** Brightness/sharpness of sound
- **Spectral Bandwidth:** Frequency spread
- **MFCCs:** 13 Mel-Frequency Cepstral Coefficients (pitch/timbre)
- **RMS Energy:** Overall loudness envelope
- **Zero Crossing Rate:** Brightness proxy
- **Chroma Features:** 12-bin pitch class distribution
- **Onset Strength:** Attack/percussion detection
- **Energy Profile:** Per-frame energy over time
- **Mood Keywords:** Inferred from features (slow/fast, bright/dark, melancholic/energetic, etc.)

**Run:** `cd analyzer && uvicorn main:py --port 8000`

---

## Dependency Graph

```
cmd/musiccli/cmd/
  └─ play.go
     ├─ internal/audio (Player, InitSpeaker)
     ├─ internal/ui (Model, NewModel)
     └─ github.com/charmbracelet/bubbletea (tea.Run)

cmd/musiccli/cmd/
  └─ feel.go
     ├─ internal/analyzer (Client, Analyze)
     ├─ internal/agent (Agent, Feel)
     ├─ internal/agent/providers (LLMProvider implementations)
     ├─ internal/audio (extractMetadata)
     └─ github.com/charmbracelet/lipgloss (styling)

internal/audio/
  ├─ github.com/gopxl/beep (Streamer, Speaker, effects)
  ├─ github.com/dhowden/tag (metadata extraction)
  └─ go stdlib (fmt, os, path, time, math/rand)

internal/ui/
  ├─ internal/audio (Player state)
  └─ github.com/charmbracelet/bubbletea (tea.Model, tea.Cmd, tea.Msg)

internal/agent/
  ├─ internal/analyzer (AudioFeatures)
  ├─ internal/audio (TrackMetadata)
  └─ internal/agent/providers (LLMProvider interface)

internal/analyzer/
  └─ net/http (HTTP client)

analyzer/main.py
  ├─ fastapi (FastAPI app)
  ├─ librosa (DSP extraction)
  ├─ soundfile (audio I/O)
  └─ numpy (math operations)
```

## Code Patterns

### Error Handling
```go
// Pattern: wrap errors with context
if err != nil {
    return fmt.Errorf("action failed: %w", err)
}
```

### Audio Loading
```go
// Load → decode → extract metadata → stream to speaker
file, _ := os.Open(path)
streamer, format, _ := flac.Decode(file)
ctrl := &beep.Ctrl{Streamer: streamer}
speaker.Play(ctrl)
```

### TUI Update Loop
```go
// Tick every 100ms + listen for track finish
case tickMsg:
    return m, m.tickCmd()  // reschedule tick
case trackFinishedMsg:
    m.player.Next()  // auto-advance
```

### LLM Provider Pattern
```go
// All providers implement LLMProvider interface
type LLMProvider interface {
    Name() string
    Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}
```

---

## Key Design Decisions

1. **Cobra for CLI:** Multi-command structure (play, feel) scales better than simple flags
2. **BubbleTea for TUI:** Functional model-update-view pattern; reactive keyboard handling
3. **100ms UI Polling:** Balance between responsiveness and CPU usage
4. **Async Track Finish:** Separate goroutine waits for `p.done` channel; auto-play next
5. **Provider Interface:** Pluggable LLM backends; easy to add Hugging Face, Azure, etc.
6. **HTTP Client for DSP:** Decouples analyzer service; allows Python implementation
7. **No Config File:** MVP simplicity; env vars + CLI flags sufficient

---

## Testing Status

| Component | Tests | Status |
|-----------|-------|--------|
| Audio playback | None | ✗ Manual testing only |
| TUI rendering | None | ✗ Renders but untested |
| LLM providers | None | ✗ Requires API keys |
| Analyzer client | None | ✗ Integration only |
| DSP feature extraction | None | ✗ Python service untested |

**Coverage Target:** 70% (currently 0%)

---

## Performance Characteristics

| Operation | Latency | Notes |
|-----------|---------|-------|
| Speaker init | <50ms | System speaker interface |
| FLAC decode/buffer | <300ms | Depends on file size |
| Metadata extraction | <10ms | ID3/FLAC tag parsing |
| First UI render | <100ms | BubbleTea init + first View() |
| DSP analysis | 2-5s | librosa feature extraction |
| LLM response | 5-30s | Depends on model and network |
| Track switch | 200-400ms | Speaker buffer flush |
| UI refresh | <50ms | 100ms polling interval |

