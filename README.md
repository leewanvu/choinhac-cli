# CLI Music Player

A high-performance, minimalist, and robust Command Line Interface (CLI) music player written in Go. It supports FLAC and WAV streaming with high-fidelity audio playback and features a sleek minimal Terminal User Interface (TUI).

## Features

- **High-Fidelity AudioPlayback**: Powered by `gopxl/beep`, supporting bit-perfect delivery and real-time resampling for mixed sample rates.
- **Modern TUI**: Built with BubbleTea and Lipgloss for a dynamic, reactive, and beautiful terminal experience.
- **Audio Decoding**: Reads and decodes FLAC and WAV files natively in Go.
- **Metadata Support**: Extracts ID3 and FLAC tags using `dhowden/tag` to display Artist, Album, Title, Sample Rate length and volume.
- **AI Music Appreciation**: A `feel` command that analyzes digital signal processing (DSP) features using Python's `librosa` and uses LLMs (Gemini, Claude, OpenAI, OpenRouter) to write an emotional review of your music.
- **Vocal Separation & Lyrics Transcription**: A feature within the `feel` command that separates vocals and transcribes them to text using `librosa` and OpenAI Whisper.
- **Concurrent Design**: Clean separation of concerns between the Audio Thread (beep streams) and the UI Thread (BubbleTea event loop).

## Architecture

The project is structured into three main layers:
- **`internal/audio`**: The Audio Engine layer. Contains the `Player` struct which handles the `beep.Streamer`, decoding logic, state management (Play/Pause), volume controls (`beep/effects`), and resampling. Runs asynchronously with respect to the UI.
- **`internal/ui`**: The TUI layer. Contains the `Model` (BubbleTea framework) and styles (`lipgloss`). Uses a `tea.Tick` command to continually poll the audio thread for playback position safely behind mutexes.
- **`internal/agent` & `internal/analyzer`**: AI integration layers for DSP feature extraction via the Python Analyzer service and LLM prompting.
- **`cmd/player`**: Main entry point for the music player. Wires the speaker initialization, UI model, and audio controller together alongside elegant OS signal handling.
- **`cmd/feel`**: Main entry point for the AI Music Appreciation agent.

## Installation

Ensure you have Go 1.21+ installed.

```bash
git clone <repository>
cd choinhaccli
go mod tidy
go build -o build/player ./cmd/player
```

## Usage

Provide an absolute or relative path to a supported audio file (`.flac` or `.wav`).

### 1. Music Player
```bash
./build/player track.flac
# or
go run cmd/player/main.go track.flac
```

### 2. AI Music Appreciation (`feel`)
Before using `feel`, ensure you have the `analyzer` running and an LLM API key exported (e.g. `GEMINI_API_KEY`). See [cmd/feel/README.md](cmd/feel/README.md) for detailed setup.
```bash
# Standard analysis
go run cmd/feel/main.go track.wav

# Analysis with lyrics extraction
go run cmd/feel/main.go --separate track.wav
```

### Controls

- **`space`**: Play / Pause
- **`up` / `+`**: Increase Volume
- **`down` / `-`**: Decrease Volume
- **`q` or `ctrl+c`**: Quit player gracefully
