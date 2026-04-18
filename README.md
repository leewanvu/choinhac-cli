# musiccli — AI-Powered Terminal Music Player

A high-performance, minimalist CLI music player written in Go with AI-powered music appreciation. Plays FLAC and WAV files with bit-perfect audio quality, offers an intuitive Terminal User Interface (TUI), and can analyze tracks with DSP + LLMs to generate emotional music reviews.

**Binary:** `musiccli`  
**Module:** `choinhaccli`  
**Language:** Go 1.25 + Python 3.10+

## Features

### 🎵 Music Playback
- **High-Fidelity Audio:** Powered by `gopxl/beep` — bit-perfect streaming with real-time resampling (24-bit/192kHz support)
- **Format Support:** FLAC and WAV files natively decoded in Go
- **Metadata:** ID3 v2.3 and FLAC tags (artist, album, title, sample rate, duration)
- **Controls:** Play/pause, next/prev, random, volume (±20dB range)
- **Playlist:** Single file or directory auto-load with seamless navigation

### 🖥️ Modern TUI
- **Framework:** BubbleTea + Lipgloss for responsive, reactive terminal UI
- **Real-Time Display:** Album art, 24-bar visualizer, progress bar with smooth gradient fill, metadata, playlist window (7 tracks visible), status
- **Visual Enhancements:** Unicode half-block album art (JPEG/PNG cover extraction), animated frequency visualizer with decay
- **Theme:** Catppuccin Mocha palette with consistent color styling
- **Keyboard-Driven:** Efficient control without mouse
- **Responsive:** 100ms refresh rate for smooth experience

### 🤖 AI Music Appreciation (`feel` command)
- **DSP Analysis:** Python FastAPI service using `librosa` for audio feature extraction
- **Features:** BPM, musical key, spectral centroid, MFCCs, mood keywords, energy profile
- **LLM Integration:** Multi-provider support (OpenAI, Google Gemini, Claude, OpenRouter)
- **Output Languages:** Vietnamese (default) or English
- **Result:** 2-3 paragraph emotional + technical music review

## Installation

### Prerequisites
- **Go 1.25+** — Download from [golang.org](https://golang.org/dl)
- **Python 3.10+** — For DSP analyzer service (optional if not using `feel`)

### Build

```bash
git clone <repository>
cd musiccli
go mod tidy
go build -o musiccli ./cmd/musiccli
```

The binary will be available as `./musiccli`.

### Python Analyzer Service (for `feel` command)

```bash
cd analyzer
python3 -m venv .venv
source .venv/bin/activate  # or: .venv\Scripts\activate on Windows
pip install -r requirements.txt
uvicorn main:app --port 8000
```

The service will run at `http://localhost:8000`.

## Quick Start

### Play an Audio File

```bash
# Play a single file
./musiccli play ~/music/song.flac

# Or play all FLAC files in a directory
./musiccli play ~/music/album/
```

### Get AI Review of Your Music

```bash
# Requires: Python analyzer running on localhost:8000
# Requires: LLM API key in environment (e.g., OPENAI_API_KEY)

./musiccli feel ~/music/song.flac --provider openrouter --lang vi
```

Available providers: `openai`, `gemini`, `claude`, `openrouter`  
Available languages: `vi` (Vietnamese), `en` (English)

## Controls

### Play Command

| Key(s) | Action |
|--------|--------|
| `space` | Play / Pause |
| `n` / `→` | Next track |
| `p` / `←` | Previous track |
| `r` | Random track |
| `+` / `↑` | Volume up |
| `-` / `↓` | Volume down |
| `q` / `Ctrl+C` | Quit |

### Feel Command

No interactive controls — generates review and exits.

## Configuration

### Environment Variables

**For `feel` command:**

```bash
# LLM API Keys (choose one)
export OPENAI_API_KEY="sk-..."         # OpenAI
export GEMINI_API_KEY="AIzaSy..."      # Google Gemini
export CLAUDE_API_KEY="sk-ant-..."     # Anthropic Claude
export OPENROUTER_API_KEY="sk-or-..."  # OpenRouter

# Analyzer service (optional, defaults to localhost:8000)
export ANALYZER_URL="http://localhost:8000"
```

## Architecture

```
musiccli play <path>                musiccli feel <audio_file>
        ↓                                    ↓
    cmd/musiccli ────────────────────────────┤
    (Cobra Router)                          │
        │                                    │
    ┌───┴──────────────────────┐            │
    │                          │            │
internal/audio            internal/agent   │
(beep Streamer)           (LLMProvider)    │
    │                          │            │
    ├─ Player.Play()           ├─ Agent.Feel()
    ├─ Player.Pause()          └─ provider.Chat()
    ├─ Player.Next()                │
    └─ metadata extraction          │
    │                          internal/analyzer
    internal/ui                (HTTP client)
    (BubbleTea Model)              │
    │                          Python FastAPI
    └─────────────────────────────►(librosa DSP)
         TUI Rendering
```

See [`docs/system-architecture.md`](docs/system-architecture.md) for detailed architecture diagrams and data flows.

## Project Structure

```
musiccli/
├── cmd/musiccli/
│   ├── main.go           — Entry point
│   └── cmd/
│       ├── root.go       — Cobra command router
│       ├── play.go       — `play` subcommand
│       └── feel.go       — `feel` subcommand
├── internal/
│   ├── audio/
│   │   ├── player.go     — Audio engine (298 LOC)
│   │   └── amplitude_tracker.go — Lock-free amplitude capture (atomic)
│   ├── ui/
│   │   ├── model.go      — BubbleTea model (~85 LOC, refactored)
│   │   ├── view.go       — TUI layout & rendering (~165 LOC)
│   │   ├── album_art.go  — Album art decoder & Unicode renderer
│   │   ├── visualizer.go — 24-bar frequency visualizer
│   │   └── style.go      — Catppuccin Mocha styling
│   ├── agent/
│   │   ├── agent.go      — LLM orchestrator
│   │   ├── prompt.go     — Prompt builders
│   │   └── providers/    — OpenAI, Gemini, Claude, OpenRouter
│   └── analyzer/
│       └── analyzer.go   — HTTP client for DSP service
├── analyzer/
│   ├── main.py          — FastAPI + librosa service
│   ├── requirements.txt  — Python dependencies
│   └── README.md         — Analyzer setup guide
├── docs/
│   ├── project-overview-pdr.md    — Vision & requirements
│   ├── codebase-summary.md        — Package breakdown
│   ├── code-standards.md          — Go/Python conventions
│   ├── system-architecture.md     — Architecture & data flows
│   └── project-roadmap.md         — Features & timeline
└── README.md (this file)
```

## Documentation

- **[project-overview-pdr.md](docs/project-overview-pdr.md)** — Vision, features, functional requirements, roadmap
- **[codebase-summary.md](docs/codebase-summary.md)** — Package-by-package breakdown, type signatures, dependencies
- **[code-standards.md](docs/code-standards.md)** — Go/Python conventions, error handling, patterns
- **[system-architecture.md](docs/system-architecture.md)** — Architecture diagrams, data flows, concurrency model
- **[project-roadmap.md](docs/project-roadmap.md)** — Milestones, Phase 2-3 features, timeline

## Known Limitations

- **No Seeking:** Can't click progress bar to jump to time (roadmap for Phase 2)
- **No MP3:** Only FLAC/WAV supported (MP3 requires CGO; planned Phase 2)
- **No Config File:** CLI flags + env vars only (planned Phase 2)
- **200-400ms Track Gap:** Gapless playback not yet implemented (Phase 2)
- **No Tests Yet:** Manual testing only; automated tests planned Phase 2 (target 70% coverage)
- **Windows Untested:** Built for macOS/Linux; Windows support TBD

## Performance

| Metric | Value | Notes |
|--------|-------|-------|
| **App Startup** | <100ms | Go binary + speaker init |
| **UI Responsiveness** | <100ms | 100ms polling interval |
| **Memory (playback)** | ~15-20MB | Audio buffers + metadata |
| **CPU (idle playback)** | <5% | Decoder + streaming |
| **DSP Analysis** | 2-5s | librosa feature extraction |
| **LLM API Response** | 5-30s | Depends on model + network |

## System Requirements

| Component | Requirement |
|-----------|-------------|
| **OS** | macOS 10.14+, Linux (tested); Windows (untested) |
| **Go** | 1.25+ |
| **Python** | 3.10+ (for `feel` command only) |
| **Audio** | System speaker/headphones |
| **Network** | Optional (only for `feel` command) |

## Building from Source

```bash
# Clone and prepare
git clone <repo> && cd musiccli
go mod download

# Build binary
go build -o musiccli ./cmd/musiccli

# Run tests (when available)
go test ./...

# Format code
go fmt ./...
```

## Troubleshooting

### "Failed to initialize audio speaker"
- **Linux:** Install ALSA/PulseAudio: `sudo apt install libasound2-dev`
- **macOS:** Ensure CoreAudio is available (should be by default)
- **Windows:** May require DirectSound setup

### "Analyzer service not available"
- Start the Python service: `cd analyzer && uvicorn main:app --port 8000`
- Check URL: `curl http://localhost:8000/health`

### "LLM API Error"
- Verify API key is exported: `echo $OPENAI_API_KEY` (or other provider)
- Check internet connectivity
- Verify provider API endpoint is accessible

## Future Roadmap

**Phase 2 (2026 Q3-Q4):**
- [ ] Test suite (70% coverage)
- [ ] MP3 support
- [ ] Audio seeking (progress bar click)
- [ ] Config file support (~/.config/musiccli/)
- [ ] Playback history & stats

**Phase 3 (2027+):**
- [ ] Gapless playback
- [ ] Audio output device selection
- [ ] Equalizer / audio filters
- [ ] HTTP/S3 streaming sources
- [ ] Real-time spectrum visualization
- [ ] Remote control / headless mode

See [`project-roadmap.md`](docs/project-roadmap.md) for detailed phases and timelines.

## License

TBD — Likely MIT or Apache 2.0

## Contributing

Contributions welcome! Please see [`code-standards.md`](docs/code-standards.md) for development guidelines.

## Support

- **Bug reports:** Create an issue on GitHub
- **Feature requests:** Check the roadmap first
- **Questions:** Open a discussion thread
