# System Architecture

**Project:** musiccli (choinhaccli)  
**Version:** MVP (Phase 1)  
**Last Updated:** 2026-04-18

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI User Input                          │
│                                                                  │
│  $ musiccli play <path>     $ musiccli feel <audio_file>       │
└────────────────┬──────────────────────────────────────────────┘
                 │
        ┌────────▼──────────┐
        │  cmd/musiccli     │
        │  (Cobra Router)   │
        └────────┬──────────┘
                 │
        ┌────────┴──────────────────┐
        │                           │
    ┌───▼────────┐         ┌──────▼──────────┐
    │  play.go   │         │   feel.go       │
    │  Handler   │         │   Handler       │
    └───┬────────┘         └──────┬──────────┘
        │                         │
        │         ┌───────────────┤
        │         │               │
    ┌───▼──────────┴─┐    ┌──────▼──────────┐
    │ internal/     │    │ internal/       │
    │ audio/        │    │ agent/          │
    │ Player        │    │ Agent           │
    │ InitSpeaker   │    │ (LLMProvider)   │
    └───┬──────────┬─┘    └──────┬──────────┘
        │          │             │
        │          │    ┌────────▼─────────┐
        │          │    │ internal/        │
        │          │    │ analyzer/        │
        │          │    │ Client (HTTP)    │
        │          │    └────────┬─────────┘
        │          │             │
    ┌───▼──────┐ ┌──▼──────────┐ │
    │ beep/    │ │ internal/   │ │
    │ Speaker  │ │ ui/         │ │
    │ Streamer │ │ Model (TUI) │ │
    └───┬──────┘ └──┬─────────┬┘ │
        │           │         │  │
        │           │         │  │  ┌────────────────┐
        │           │         │  │  │ Python Service │
        │           │         │  │  │ (FastAPI)      │
        │           │         │  │  │ librosa DSP    │
        │           │         │  │  └─────────────────┘
        │           │         │  └─────────┬──────────┘
        │           │         │            │
        │  ┌────────▼─────────┼────────────┘
        │  │                  │
    ┌───▼──────────────────────▼───┐
    │   System Audio Device         │
    │   (Speaker Output)            │
    └──────────────────────────────┘
```

---

## Component Responsibilities

### Layer 1: CLI Entry Point (`cmd/musiccli`)

**Role:** Command routing and argument validation

**Components:**
- `main.go` — Minimal bootstrap; calls `cmd.Execute()`
- `cmd/root.go` — Cobra root command; registers subcommands
- `cmd/play.go` — `play` subcommand handler
- `cmd/feel.go` — `feel` subcommand handler

**Responsibilities:**
- Parse CLI arguments (file path, flags)
- Validate file existence and format (.flac, .wav)
- Build playlist from directory or single file
- Instantiate lower-level components (Player, Model, Agent)
- Wire components together
- Handle graceful shutdown (Ctrl+C)

**Data Flow: Play Command**
```
Args: [path] → File stat → Is directory? 
    ├─ Yes → Scan for .wav/.flac → Sort → Build playlist
    └─ No → Validate format → Playlist = [path]
→ audio.InitSpeaker() → audio.NewPlayer() → LoadPlaylist()
→ ui.NewModel(player) → tea.NewProgram(model).Run()
```

**Data Flow: Feel Command**
```
Args: [audio_file] → Validate format → Absolute path
→ Extract metadata (ID3/FLAC) → Print header
→ analyzer.NewClient() → Analyze(path) → [AudioFeatures]
→ Print features → Select LLM provider → agent.Feel(features, metadata)
→ LLM API call → Print review
```

---

### Layer 2A: Audio Engine (`internal/audio`)

**Role:** Decode, stream, and manage audio playback

**Core Type:**
```go
type Player struct {
    ctrl     *beep.Ctrl           // Playback control
    volume   *effects.Volume      // Gain adjustment
    streamer beep.StreamSeekCloser // Decoder output
    format   beep.Format          // Sample rate, channels
    state    State                // Stopped/Playing/Paused
    Metadata TrackMetadata        // ID3/FLAC tags
    Playlist []string             // Track list
    PlaylistIdx int               // Current index
    done     chan bool            // Track finish signal
}
```

**Key Responsibilities:**
1. **Audio Decoding:** Load FLAC/WAV files via gopxl/beep decoders
2. **Metadata Extraction:** Parse ID3 v2.3 / FLAC tags using dhowden/tag
3. **Playback Control:** Play, pause, stop, resume via beep.Ctrl
4. **Volume Management:** Adjust gain via beep/effects.Volume (+/-20dB range)
5. **Playlist Navigation:** Next, prev, random with wraparound
6. **State Management:** Track current playback state (Playing/Paused/Stopped)
7. **Position Tracking:** Report elapsed time to UI for progress bar

**State Machine:**
```
┌─────────┐
│Stopped  │ ◄─── Initial state
└────┬────┘
     │ LoadAndPlay()
     ▼
┌─────────┐
│Playing  │ ◄─── Actively streaming audio
└─┬───┬───┘
  │   │ Pause()
  │   ▼
  │ ┌─────────┐
  │ │Paused   │ ◄─── Paused, can Resume
  │ └────┬────┘
  │      │ Play()
  │      ▼
  └───► Playing

Any state ──Stop()──► Stopped (close streamer)
Playing ───────────────► Stopped (when track finishes)
```

**Audio Streaming Pipeline:**
```
File (.flac/.wav)
    ↓
os.Open()
    ↓
flac.Decode() or wav.Decode()  [gopxl/beep decoders]
    ↓
beep.Resample()  [if sample rate ≠ 44.1kHz]
    ↓
beep.Ctrl (PlaybackControl)
    ↓
effects.Volume (Gain adjustment)
    ↓
speaker.Play()  [Output to system speaker]
```

**Concurrency:**
- Main goroutine: UI polling (BubbleTea event loop)
- Audio goroutine: beep.Speaker (internal to gopxl/beep)
- Track finish: Goroutine waits on `p.done` channel; signals UI

**Thread-Safe Operations:**
- State read/write via getters/setters (no mutex; atomic-like behavior)
- Position tracking: Safe read from beep.Streamer via mutex in beep
- Volume adjustment: Safe modification via beep/effects.Volume

---

### Layer 2B: TUI / User Interface (`internal/ui`)

**Role:** Real-time terminal UI using BubbleTea framework

**Core Type:**
```go
type Model struct {
    player *audio.Player  // Reference to audio engine
    width  int           // Terminal width (for responsive layout)
    err    error         // Error state (displayed if non-nil)
}
```

**BubbleTea Lifecycle:**

```
Program Start
    ↓
Model.Init()  ───► Returns initial Cmd
    ↓                 ├─ tickCmd() [100ms polling]
    ├─── Render ─────► Model.View() ───► Draw to terminal
    ├─── Wait for Msg
    │    ├─ KeyMsg (keyboard input)
    │    ├─ tickMsg (100ms poll)
    │    ├─ trackFinishedMsg (track end signal)
    │    └─ WindowSizeMsg (terminal resize)
    │
    ├─ Call Model.Update(msg)
    │  ├─ Update state (player.Play(), etc.)
    │  └─ Return (model, cmd)
    │
    ├─ Execute cmd (if any)
    │    ├─ Schedule next tick
    │    ├─ Listen for track finish
    │    └─ Schedule quit
    │
    └─► Loop until tea.Quit

Program End
```

**Key Responsibilities:**
1. **Event Handling:** Keyboard input → Player method calls
2. **Polling:** 100ms tick to fetch player state (position, status)
3. **Track Finish Detection:** Goroutine waits on `player.Done()` channel
4. **Responsive Layout:** Calculate progress bar width based on terminal width
5. **Rendering:** Compose view from metadata, progress, playlist, help text

**Control Mapping:**

| Key(s) | Action | Player Method |
|--------|--------|---------------|
| `space` | Play/Pause | `TogglePause()` |
| `n`, `→` | Next track | `Next()` |
| `p`, `←` | Previous track | `Prev()` |
| `r` | Random track | `Random()` |
| `+`, `↑` | Volume up | `VolumeUp()` |
| `-`, `↓` | Volume down | `VolumeDown()` |
| `q`, `Ctrl+C` | Quit | `Stop()` + `tea.Quit` |

**View Components:**
```
┌─────────────────────────────────────────┐
│ 🎵 CLI Music Player                     │  ← Title
├─────────────────────────────────────────┤
│ Artist: Tame Impala                     │  ← Metadata
│ Album: Currents                         │
│ Track: The Less I Know The Better [1/5] │
├─────────────────────────────────────────┤
│ Sample Rate: 44100 Hz | Playing | Vol: +1.5dB  ← Stats
│ ████████░░░░░░░░░░ 1:45 / 3:20         │  ← Progress
├─────────────────────────────────────────┤
│ Playlist:                               │  ← Playlist window
│   1. Theresa's Sound World              │
│ ▶ 2. The Less I Know The Better         │
│   3. Eventually                         │
│     ...                                 │
├─────────────────────────────────────────┤
│ space: play/pause • n/→: next • ...     │  ← Help
└─────────────────────────────────────────┘
```

**Polling Loop:**
```
Every 100ms:
    ├─ Fetch player.GetPosition() → elapsed time
    ├─ Fetch player.GetState() → playback status
    ├─ Fetch player.GetVolume() → volume offset
    ├─ Fetch player.Metadata → artist, album, title
    └─ Re-render Model.View() ───► Terminal output
```

---

### Layer 2C: AI Music Appreciation (`internal/agent`)

**Role:** Orchestrate DSP analysis and LLM-based review generation

**Core Types:**
```go
type LLMProvider interface {
    Name() string
    Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type Agent struct {
    provider LLMProvider
    lang     string  // "vi" or "en"
}
```

**Processing Pipeline:**
```
Audio File
    ↓
analyzer.Analyze(path) ──► AudioFeatures (JSON from Python service)
    │
    ├─ BPM: 120
    ├─ Key: "C major"
    ├─ MFCCs: [0.5, 1.2, 0.8, ...]
    ├─ Mood Keywords: ["energetic", "bright", "fast"]
    └─ ...
    ↓
buildUserPrompt(features, metadata) ──► Text prompt
    │
    ├─ "Title: The Less I Know The Better"
    ├─ "Artist: Tame Impala"
    ├─ "Audio Features:"
    ├─ "- BPM: 120"
    ├─ "- Key: C major"
    ├─ "- Mood: energetic, bright"
    └─ ...
    ↓
buildSystemPrompt(lang) ──► System instruction
    │
    └─ "You are a music critic. Write 2-3 paragraphs..." (Vietnamese or English)
    ↓
agent.Feel(features, metadata)
    │
    ├─ LLM provider selection (OpenAI, Gemini, Claude, OpenRouter)
    ├─ API call: provider.Chat(systemPrompt, userPrompt)
    └─ Context timeout: 30-60s
    ↓
Review (2-3 paragraphs from LLM)
    │
    └─ Output to user via formatted printing
```

**LLM Provider Implementations:**

All providers implement the `LLMProvider` interface:

1. **OpenAI** (`providers/openai.go`)
   - Model: gpt-3.5-turbo (default) or gpt-4
   - API: https://api.openai.com/v1/chat/completions
   - Auth: Bearer token from OPENAI_API_KEY

2. **Google Gemini** (`providers/gemini.go`)
   - Model: gemini-pro (default)
   - API: https://generativelanguage.googleapis.com/v1beta/models/...
   - Auth: API key from GEMINI_API_KEY

3. **Anthropic Claude** (`providers/claude.go`)
   - Model: claude-3-sonnet (default) or claude-3-opus
   - API: https://api.anthropic.com/v1/messages
   - Auth: x-api-key header from CLAUDE_API_KEY

4. **OpenRouter** (`providers/openrouter.go`)
   - Model: Proxy to multiple models (default: meta-llama/llama-2-70b)
   - API: https://openrouter.ai/api/v1/chat/completions
   - Auth: Bearer token from OPENROUTER_API_KEY

**Prompt Structure:**

**System Prompt (Vietnamese):**
```
Bạn là một nhà phê bình âm nhạc tài năng và có cảm xúc.
Hãy viết một bài review ngắn (2-3 đoạn) về bản nhạc được mô tả dưới đây.
Nhận xét về những đặc tính âm học, cảm xúc, và sự sáng tạo.
Sử dụng ngôn ngữ tự nhiên và thân thiện.
```

**System Prompt (English):**
```
You are a talented and emotional music critic.
Write a brief review (2-3 paragraphs) of the music described below.
Comment on acoustic characteristics, emotions, and creativity.
Use natural, friendly language.
```

**User Prompt:**
```
Title: The Less I Know The Better
Artist: Tame Impala
Album: Currents
Duration: 3:20

Audio Features:
- BPM: 120
- Key: C major
- Spectral Centroid: 2500 Hz
- Mood Keywords: energetic, bright, fast, rhythmic
- RMS Energy Mean: 0.45

Please analyze this music and share your feelings about it.
```

**Response Format:**
```
Đây là một bài hát pop điện tử đầy năng lượng với nhịp độ nhanh...
[2-3 paragraphs from LLM]
```

---

### Layer 3: External Services

#### 3A: Python DSP Analyzer (`analyzer/main.py`)

**Role:** Extract audio features using librosa DSP

**Service Type:** FastAPI HTTP microservice

**Endpoints:**

```
GET /health
Response: {"status": "ok"}

POST /analyze
Form Data: path=<file_path>
Response: AudioFeatures JSON

Example:
{
    "bpm": 120.5,
    "key": "C major",
    "spectral_centroid_mean": 2500.0,
    "spectral_bandwidth_mean": 1200.0,
    "mfcc_means": [0.5, 1.2, 0.8, ...],
    "rms_energy_mean": 0.45,
    "zero_crossing_rate_mean": 0.12,
    "chroma_features": {"C": 0.8, "C#": 0.2, ...},
    "onset_strength_mean": 0.6,
    "duration_seconds": 200.0,
    "energy_profile": [0.1, 0.2, 0.5, ...],
    "mood_keywords": ["energetic", "bright", "fast"]
}
```

**Features Extracted:**

| Feature | Tool | Purpose |
|---------|------|---------|
| **BPM** | `librosa.beat.tempo()` | Tempo/pace |
| **Key** | `librosa.feature.chroma_cqt()` + correlation | Musical tonality |
| **Spectral Centroid** | `librosa.feature.spectral_centroid()` | Brightness |
| **Spectral Bandwidth** | `librosa.feature.spectral_bandwidth()` | Frequency spread |
| **MFCCs** | `librosa.feature.mfcc()` | 13-bin timbre representation |
| **RMS Energy** | `librosa.feature.rms()` | Loudness |
| **Zero Crossing Rate** | `librosa.feature.zero_crossing_rate()` | Noisiness |
| **Chroma Features** | `librosa.feature.chroma_cqt()` | Pitch distribution |
| **Onset Strength** | `librosa.onset.onset_strength()` | Attack/percussion |
| **Mood Keywords** | Custom logic from above | Derived metadata |

**Processing Time:** 2-5 seconds per file (depends on file size and CPU)

**Deployment:** `cd analyzer && uvicorn main:app --host 0.0.0.0 --port 8000`

#### 3B: System Audio Device

**Role:** Physical audio output

**Integration:** via gopxl/beep's `speaker.Init()` and `speaker.Play()`

**Supported Platforms:**
- macOS: CoreAudio
- Linux: PulseAudio / ALSA
- Windows: DirectSound (untested)

**Configuration:**
- Sample Rate: 44.1 kHz (resampling handles other rates)
- Buffer Size: 441 frames (~10ms latency)
- Channels: Stereo (no mono/surround support)

---

## Data Flow Diagrams

### Play Command Data Flow

```
User Input: musiccli play ~/music/album
    ↓
cmd/play.go::runPlay()
    ├─ os.Stat(path)
    ├─ IsDir? → ReadDir + filter .flac/.wav → Sort
    └─ Else → Single file (format check)
    ↓
audio.InitSpeaker()  [Once only]
    ↓
audio.NewPlayer()
    ↓
p.LoadPlaylist(playlist, 0)
    ├─ p.LoadAndPlay(playlist[0])
    │  ├─ os.Open(path)
    │  ├─ flac.Decode() or wav.Decode()
    │  ├─ extractMetadata(path) → ID3/FLAC tags
    │  ├─ Create beep.Ctrl + effects.Volume
    │  ├─ speaker.Play(ctrl)
    │  └─ Goroutine: wait for <-p.done
    └─ p.state = StatePlaying
    ↓
ui.NewModel(p)
    ↓
tea.NewProgram(model).Run()
    ├─ Model.Init() → tickCmd + waitForTrackFinished
    └─ Event Loop:
        ├─ Every 100ms: tickMsg → Model.View() → Render
        ├─ On keypress: KeyMsg → Update() → player method
        │  (e.g., space → player.TogglePause())
        ├─ On track end: trackFinishedMsg → player.Next()
        └─ On Ctrl+C: KeyMsg("q") → player.Stop() → tea.Quit
    ↓
speaker.Close()  [On exit]
```

### Feel Command Data Flow

```
User Input: musiccli feel ~/music/track.flac --provider openrouter --lang vi
    ↓
cmd/feel.go::runFeel()
    ├─ os.Stat(path)
    ├─ Validate format (.flac or .wav)
    └─ filepath.Abs(path)
    ↓
extractMetadata(path)  [dhowden/tag]
    └─ Return: TrackMetadata{Title, Artist, Album, SampleRate, Duration}
    ↓
Print: "🎵 AI Music Appreciation" + metadata
    ↓
analyzer.NewClient("http://localhost:8000")
    ↓
client.HealthCheck()  [GET /health]
    ├─ If error: Print error → Exit
    └─ Success: Continue
    ↓
client.Analyze(absPath)  [POST /analyze with path]
    ↓
Python Service (FastAPI)
    ├─ librosa.load(path)
    ├─ Extract 13 features (BPM, key, spectral, MFCC, RMS, ZCR, chroma, onset, mood)
    └─ Return JSON
    ↓
Parse JSON → AudioFeatures struct
    ↓
Print: Features (BPM, Key, Centroid, etc.)
    ↓
createProvider(feelFlags.provider, feelFlags.model)
    └─ Return: LLMProvider (OpenAI, Gemini, Claude, or OpenRouter)
    ↓
buildSystemPrompt(feelFlags.lang)  [Vietnamese or English]
    ↓
buildUserPrompt(features, metadata)  [Format features as text]
    ↓
agent.Feel(ctx, features, metadata)
    ├─ provider.Chat(systemPrompt, userPrompt)
    └─ LLM API call (OpenAI, Gemini, Claude, or OpenRouter)
    ↓
Return: Review (2-3 paragraphs)
    ↓
Print: Formatted review with color styling
```

---

## State Management

### Player State Lifecycle

```
┌─────────────────────────────────────┐
│  NewPlayer() → state = Stopped      │
└────────┬────────────────────────────┘
         │
    LoadPlaylist(paths)
         │
         ▼
    LoadAndPlay(path[0])
    ├─ Open file
    ├─ Decode (FLAC/WAV)
    ├─ Extract metadata
    ├─ Create Ctrl + Volume
    ├─ speaker.Play(ctrl)
    ├─ Start goroutine: wait <-p.done
    └─ state = StatePlaying
         │
    ┌────┴─────────────────────┐
    │                          │
TogglePause()          TogglePause()
    │                          │
    ▼                          ▼
state = StatePaused     state = StatePlaying
    │                          │
    │                          │
    └────────────┬─────────────┘
                 │
            Next() / Prev() / Random()
                 │
         ┌───────┴────────┐
         │                │
         ├─ Close streamer
         ├─ LoadAndPlay(path[idx])
         └─ state = StatePlaying
                 │
         [Repeat cycle]

Any state ──Stop()──► state = Stopped
                     (Close streamer, flush speaker)
```

### UI Model State

```
┌─────────────────────────────────────┐
│  NewModel(player) → width = 0       │
└────────┬────────────────────────────┘
         │
    Model.Init()
    ├─ Schedule tickCmd (every 100ms)
    └─ Schedule waitForTrackFinished()
         │
    ┌────▼──────────────────────────────────┐
    │  Event Loop (BubbleTea)                │
    │                                        │
    │  tickMsg (100ms)                       │
    │  ├─ Fetch: player.GetPosition()        │
    │  ├─ Fetch: player.GetState()           │
    │  ├─ View() → Render to terminal        │
    │  └─ Schedule next tick                 │
    │                                        │
    │  KeyMsg (keyboard)                     │
    │  ├─ Case "space": player.TogglePause() │
    │  ├─ Case "n": player.Next()            │
    │  ├─ Case "q": player.Stop() + Quit     │
    │  └─ ...                                │
    │                                        │
    │  WindowSizeMsg                         │
    │  ├─ Update width                       │
    │  └─ Recalculate layout                 │
    │                                        │
    │  trackFinishedMsg                      │
    │  ├─ player.Next()                      │
    │  └─ Reschedule waitForTrackFinished()  │
    └────┬──────────────────────────────────┘
         │
    Quit (user presses q)
         │
         ▼
    Program stops
```

---

## Error Handling Strategy

### Error Types

| Error | Source | Handling | User Feedback |
|-------|--------|----------|---------------|
| **File Not Found** | os.Stat() | Return error → Exit | "✗ File not found: {path}" |
| **Unsupported Format** | filepath.Ext() | Return error → Exit | "✗ Unsupported format: {ext}" |
| **Speaker Init Failed** | audio.InitSpeaker() | Return error → Exit | "failed to initialize audio speaker: {err}" |
| **Decode Error** | flac/wav.Decode() | Return error → Stop playback | Displayed in TUI error state |
| **Metadata Parse Error** | tag.ReadFrom() | Fallback to defaults | Continue without crash |
| **Analyzer Unreachable** | analyzer.HealthCheck() | Return error → Exit | "✗ Analyzer service not available at {url}" |
| **Analyzer Timeout** | HTTP 120s timeout | Return error → Exit | "✗ Analysis failed: {err}" |
| **LLM API Error** | provider.Chat() | Return error → Exit | "✗ LLM (openai) failed: {err}" |

### Error Propagation

```
Application level (cmd/)
    ├─ Check errors immediately
    ├─ Wrap with context: fmt.Errorf("action: %w", err)
    ├─ Print user-friendly message
    └─ Exit with code 1

Internal level (internal/*)
    ├─ Create clear error messages
    ├─ Avoid panics (only in truly unrecoverable situations)
    ├─ Return (nil, error) pairs
    └─ Caller decides recovery strategy

UI level (internal/ui)
    ├─ Store error in Model.err
    ├─ Render error message in View()
    ├─ Allow user to acknowledge (Ctrl+C to quit)
    └─ Don't crash program
```

---

## Concurrency & Synchronization

### Goroutine Model

```
Main Goroutine (CLI + UI Event Loop)
    ├─ Initializes speaker + player
    ├─ Runs BubbleTea event loop (blocking)
    ├─ Handles keyboard events
    ├─ Calls player methods (Play, Pause, etc.)
    └─ Gracefully shuts down on Quit

Audio Goroutine (internal to beep.Speaker)
    ├─ Continuously reads from beep.Streamer
    ├─ Applies volume effects
    ├─ Streams to system speaker
    └─ Signals completion via player.done channel

Track Finish Listener Goroutine (TUI)
    ├─ Waits on player.done channel
    ├─ Sends trackFinishedMsg to Model.Update()
    ├─ Reschedules itself
    └─ Repeats for each track

Analyzer Request Goroutine (async, if implemented later)
    └─ [Future: Could spawn background analyze tasks]
```

### Synchronization Points

```
player.state (read/write)
    ├─ UI reads: Every 100ms in View()
    └─ Player writes: On Play, Pause, Stop, Next, Prev

player.Playlist + PlaylistIdx
    ├─ UI reads: In View() for playlist display
    └─ Player writes: In LoadPlaylist, Next, Prev, Random

player.Metadata
    ├─ UI reads: Every 100ms in View()
    └─ Player writes: In LoadAndPlay after tag.ReadFrom

player.done channel
    ├─ Audio goroutine sends: When streamer finishes
    ├─ Track listener receives: Sends trackFinishedMsg
    └─ [No mutex needed: channel is thread-safe]

speaker (global state)
    ├─ Initialized once: audio.InitSpeaker()
    ├─ Used by: Player, audio goroutine
    └─ Protected by: speakerInit flag + sync.Once (via beep)
```

**Race Conditions Avoided:**
- Player state is read-only from UI perspective (no writes from UI)
- Metadata is updated before state changes
- Channel reads/writes are thread-safe by design
- No shared buffers or memory

---

## Performance Characteristics

### Latency

| Operation | Latency | Bottleneck |
|-----------|---------|-----------|
| **App startup** | <100ms | Go binary startup |
| **Speaker init** | <50ms | System audio subsystem |
| **FLAC load/decode** | <300ms | File I/O + decoder |
| **UI first render** | <100ms | BubbleTea init |
| **UI refresh (100ms tick)** | <50ms | View() rendering |
| **Volume adjustment** | <10ms | beep/effects.Volume |
| **Track switch** | 200-400ms | Speaker buffer flush |
| **DSP analysis** | 2-5s | librosa feature extraction |
| **LLM API call** | 5-30s | Network + model inference |

### Memory

| Component | Typical Usage | Notes |
|-----------|---------------|-------|
| **Go binary (musiccli play)** | ~10MB | Includes beep, bubbletea |
| **Audio buffers** | ~5MB | 10ms buffer @ 44.1kHz stereo |
| **Player struct** | <1MB | Streamer, metadata, playlist |
| **TUI model** | <100KB | Minimal state |
| **Total for playback** | ~15-20MB | Stable across playlist |
| **Total for feel command** | ~100MB+ | Includes Python service (librosa) |

### CPU

| Operation | CPU Usage | Duration | Notes |
|-----------|-----------|----------|-------|
| **Playback (idle)** | <5% | Continuous | Depends on audio codec |
| **UI rendering (100ms tick)** | <2% per tick | <50ms | BubbleTea rendering |
| **Volume adjustment** | <1% | <10ms | Single effect update |
| **FLAC streaming** | 5-10% | Continuous | Decoder + resampling |
| **DSP analysis** | 50-80% (Python) | 2-5s | librosa + numpy |
| **LLM inference** | N/A (remote) | 5-30s | Provider API |

---

## Future Architecture Improvements

1. **Seeking:** Implement `beep.StreamSeekCloser.Seek()` for progress bar seek
2. **Gapless Playback:** Pre-load next track; reduce gap to <10ms
3. **Config File:** `~/.config/musiccli/config.yaml` for persistent settings
4. **Caching:** Store analyzer results to avoid re-processing
5. **Equalizer:** Expose beep/effects filters (bass/treble/reverb)
6. **Test Coverage:** Unit + integration tests (target 70%)

