# CLI Music Player

A high-performance, minimalist, and robust Command Line Interface (CLI) music player written in Go. It supports FLAC and WAV streaming with high-fidelity audio playback and features a sleek minimal Terminal User Interface (TUI).

## Features

- **High-Fidelity AudioPlayback**: Powered by `gopxl/beep`, supporting bit-perfect delivery and real-time resampling for mixed sample rates.
- **Modern TUI**: Built with BubbleTea and Lipgloss for a dynamic, reactive, and beautiful terminal experience.
- **Audio Decoding**: Reads and decodes FLAC and WAV files natively in Go.
- **Metadata Support**: Extracts ID3 and FLAC tags using `dhowden/tag` to display Artist, Album, Title, Sample Rate length and volume.
- **Concurrent Design**: Clean separation of concerns between the Audio Thread (beep streams) and the UI Thread (BubbleTea event loop).

## Architecture

The project is structured into three main layers:
- **`internal/audio`**: The Audio Engine layer. Contains the `Player` struct which handles the `beep.Streamer`, decoding logic, state management (Play/Pause), volume controls (`beep/effects`), and resampling. Runs asynchronously with respect to the UI.
- **`internal/ui`**: The TUI layer. Contains the `Model` (BubbleTea framework) and styles (`lipgloss`). Uses a `tea.Tick` command to continually poll the audio thread for playback position safely behind mutexes.
- **`cmd/player`**: Main entry point. Wires the speaker initialization, UI model, and audio controller together alongside elegant OS signal handling (quitting stops audio gracefully).

## Installation

Ensure you have Go 1.21+ installed.

```bash
git clone <repository>
cd cli-music-player
go mod tidy
go build -o player ./cmd/player
```

## Usage

Provide an absolute or relative path to a supported audio file (`.flac` or `.wav`).

```bash
./player track.flac
```

### Controls

- **`space`**: Play / Pause
- **`up` / `+`**: Increase Volume
- **`down` / `-`**: Decrease Volume
- **`q` or `ctrl+c`**: Quit player gracefully
