# Audio Analyzer Service

The `analyzer` directory contains a Python-based Audio Analyzer Service built with **FastAPI** and **librosa**. It is designed to extract digital signal processing (DSP) features from audio files and return them in a structured JSON format, which is primarily used by the Go CLI music appreciation agent.

## Features

This service extracts a wide variety of audio features to assist in music analysis and appreciation:

- **BPM (Beats Per Minute)**: Calculates the tempo of the track.
- **Key Detection**: Estimates the musical key (e.g., C major, A minor) using chroma features and the Krumhansl-Kessler profiles.
- **Spectral Centroid**: Measures the "brightness" of the sound.
- **Spectral Bandwidth**: Measures the width of the frequency band.
- **MFCCs (Mel-Frequency Cepstral Coefficients)**: Extracts 13 coefficients representing the short-term power spectrum.
- **RMS Energy**: Measures the power of the audio signal.
- **Zero Crossing Rate (ZCR)**: Measures the percussiveness or noisiness of the signal.
- **Chroma Features**: Energy across the 12 pitch classes.
- **Onset Strength**: Detects the strength of musical events or notes.
- **Energy Profile**: Splits the track into 20 segments to track energy levels over time.
- **Mood Keywords**: Heuristically derives mood tags (e.g., "energetic", "melancholic", "warm", "rhythmic") based on BPM, key, spectral centroid, RMS energy, and ZCR.
- **Vocal Separation**: Separates vocals from accompaniment using STFT and nearest-neighbor filtering (`librosa.decompose.nn_filter`).
- **Lyrics Transcription**: Transcribes the separated vocals into text using OpenAI's Whisper model.

## Requirements

The service requires Python and the following packages (listed in `requirements.txt`):
- `fastapi`
- `uvicorn[standard]`
- `librosa`
- `soundfile`
- `numpy`
- `python-multipart`
- `openai-whisper`

## Setup & Running

1. **Create a virtual environment (optional but recommended):**
   ```bash
   python3 -m venv .venv
   source .venv/bin/activate
   ```

2. **Install dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

3. **Run the server:**
   ```bash
   uvicorn main:app --host 127.0.0.1 --port 8000
   ```
   *(Or just run the server locally on the default port with `uvicorn main:app --reload`)*

## API Endpoints

### 1. `POST /analyze`
Analyzes a local audio file by providing its absolute path.

- **Content-Type:** `application/x-www-form-urlencoded`
- **Parameters:**
  - `path` (string): The absolute file path to the audio file on the server.
- **Returns:** JSON object containing the extracted audio features.

### 2. `POST /analyze/upload`
Analyzes an uploaded audio file.

- **Content-Type:** `multipart/form-data`
- **Parameters:**
  - `file` (file): The audio file to upload and analyze.
- **Returns:** JSON object containing the extracted audio features.

### 3. `POST /separate`
Extracts lyrics from a local audio file by separating vocals and transcribing them using Whisper.

- **Content-Type:** `application/x-www-form-urlencoded`
- **Parameters:**
  - `path` (string): The absolute file path to the audio file on the server.
- **Returns:** JSON object containing the transcribed lyrics:
  ```json
  {
    "lyrics": "Transcribed lyrics text here..."
  }
  ```

### 4. `GET /health`
Checks the health status of the service.
- **Returns:** `{"status": "ok"}`
