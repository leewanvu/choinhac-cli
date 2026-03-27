"""
Audio Analyzer Service — FastAPI + librosa.

Extracts DSP features from audio files and returns structured JSON
for the Go CLI music appreciation agent.
"""

import os
from pathlib import Path

import librosa
import numpy as np
from fastapi import FastAPI, HTTPException, UploadFile, File, Form
from fastapi.responses import JSONResponse
import soundfile as sf
import tempfile

app = FastAPI(title="Music Analyzer", version="1.0.0")


def _detect_key(chroma: np.ndarray) -> str:
    """Estimate musical key from chroma features."""
    pitch_classes = ["C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"]

    # Average chroma energy across time
    chroma_mean = np.mean(chroma, axis=1)

    # Major and minor profile templates (Krumhansl-Kessler)
    major_profile = np.array([6.35, 2.23, 3.48, 2.33, 4.38, 4.09, 2.52, 5.19, 2.39, 3.66, 2.29, 2.88])
    minor_profile = np.array([6.33, 2.68, 3.52, 5.38, 2.60, 3.53, 2.54, 4.75, 3.98, 2.69, 3.34, 3.17])

    best_corr = -1.0
    best_key = "C major"

    for shift in range(12):
        rolled = np.roll(chroma_mean, -shift)
        corr_major = float(np.corrcoef(rolled, major_profile)[0, 1])
        corr_minor = float(np.corrcoef(rolled, minor_profile)[0, 1])

        if corr_major > best_corr:
            best_corr = corr_major
            best_key = f"{pitch_classes[shift]} major"
        if corr_minor > best_corr:
            best_corr = corr_minor
            best_key = f"{pitch_classes[shift]} minor"

    return best_key


def _compute_mood_keywords(
    bpm: float,
    key: str,
    spectral_centroid_mean: float,
    rms_mean: float,
    zcr_mean: float,
) -> list[str]:
    """Derive mood keywords from audio features."""
    moods = []

    # Tempo
    if bpm < 80:
        moods.append("slow")
    elif bpm < 120:
        moods.append("moderate")
    else:
        moods.append("fast")

    if bpm > 130:
        moods.append("energetic")

    # Key (major vs minor)
    if "minor" in key:
        moods.append("melancholic")
    else:
        moods.append("bright")

    # Spectral centroid — brightness
    if spectral_centroid_mean > 3000:
        moods.append("brilliant")
    elif spectral_centroid_mean < 1500:
        moods.append("warm")
    else:
        moods.append("balanced")

    # Energy
    if rms_mean > 0.15:
        moods.append("powerful")
    elif rms_mean < 0.05:
        moods.append("gentle")

    # ZCR — percussiveness
    if zcr_mean > 0.1:
        moods.append("rhythmic")
    else:
        moods.append("smooth")

    return moods


def _analyze_file(file_path: str) -> dict:
    """Core analysis logic. Returns feature dict."""
    if not os.path.isfile(file_path):
        raise FileNotFoundError(f"File not found: {file_path}")

    # Load audio (mono, default sr=22050)
    y, sr = librosa.load(file_path, sr=None, mono=True)
    duration = float(librosa.get_duration(y=y, sr=sr))

    # --- Feature extraction ---

    # BPM
    tempo, _ = librosa.beat.beat_track(y=y, sr=sr)
    bpm = float(np.atleast_1d(tempo)[0])

    # Chroma & Key detection
    chroma = librosa.feature.chroma_cqt(y=y, sr=sr)
    key = _detect_key(chroma)

    # Build chroma dict
    pitch_classes = ["C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"]
    chroma_mean = np.mean(chroma, axis=1)
    chroma_dict = {pitch_classes[i]: round(float(chroma_mean[i]), 4) for i in range(12)}

    # Spectral centroid
    sc = librosa.feature.spectral_centroid(y=y, sr=sr)
    sc_mean = float(np.mean(sc))

    # Spectral bandwidth
    sb = librosa.feature.spectral_bandwidth(y=y, sr=sr)
    sb_mean = float(np.mean(sb))

    # MFCCs (13 coefficients)
    mfcc = librosa.feature.mfcc(y=y, sr=sr, n_mfcc=13)
    mfcc_means = [round(float(m), 4) for m in np.mean(mfcc, axis=1)]

    # RMS energy
    rms = librosa.feature.rms(y=y)
    rms_mean = float(np.mean(rms))

    # Zero crossing rate
    zcr = librosa.feature.zero_crossing_rate(y)
    zcr_mean = float(np.mean(zcr))

    # Onset strength
    onset_env = librosa.onset.onset_strength(y=y, sr=sr)
    onset_mean = float(np.mean(onset_env))

    # Energy profile (split into 20 segments)
    n_segments = 20
    rms_flat = rms.flatten()
    segment_len = max(1, len(rms_flat) // n_segments)
    energy_profile = []
    for i in range(n_segments):
        start = i * segment_len
        end = min(start + segment_len, len(rms_flat))
        if start < len(rms_flat):
            energy_profile.append(round(float(np.mean(rms_flat[start:end])), 4))

    # Mood keywords
    mood_keywords = _compute_mood_keywords(bpm, key, sc_mean, rms_mean, zcr_mean)

    return {
        "bpm": round(bpm, 1),
        "key": key,
        "spectral_centroid_mean": round(sc_mean, 2),
        "spectral_bandwidth_mean": round(sb_mean, 2),
        "mfcc_means": mfcc_means,
        "rms_energy_mean": round(rms_mean, 4),
        "zero_crossing_rate_mean": round(zcr_mean, 4),
        "chroma_features": chroma_dict,
        "onset_strength_mean": round(onset_mean, 2),
        "duration_seconds": round(duration, 2),
        "energy_profile": energy_profile,
        "mood_keywords": mood_keywords,
    }


@app.post("/analyze")
async def analyze_by_path(path: str = Form(...)):
    """Analyze a local audio file by path."""
    try:
        features = _analyze_file(path)
        return JSONResponse(content=features)
    except FileNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Analysis failed: {e}")


@app.post("/analyze/upload")
async def analyze_by_upload(file: UploadFile = File(...)):
    """Analyze an uploaded audio file."""
    suffix = Path(file.filename or "audio.wav").suffix
    try:
        with tempfile.NamedTemporaryFile(suffix=suffix, delete=False) as tmp:
            content = await file.read()
            tmp.write(content)
            tmp_path = tmp.name

        features = _analyze_file(tmp_path)
        return JSONResponse(content=features)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Analysis failed: {e}")
    finally:
        if "tmp_path" in locals():
            os.unlink(tmp_path)


@app.get("/health")
async def health():
    return {"status": "ok"}


def _extract_lyrics(file_path: str) -> str:
    """Separate vocals from accompaniment and transcribe lyrics using Whisper."""
    import whisper

    if not os.path.isfile(file_path):
        raise FileNotFoundError(f"File not found: {file_path}")

    # Step 1: Load audio and separate vocals via nearest-neighbor filter
    y, sr = librosa.load(file_path, sr=None, mono=True)

    S_full, phase = librosa.magphase(librosa.stft(y))

    width = int(librosa.time_to_frames(2, sr=sr))
    S_filter = librosa.decompose.nn_filter(S_full,
                                           aggregate=np.median,
                                           metric='cosine',
                                           width=width)

    S_filter = np.minimum(S_full, S_filter)

    margin_v = 10
    power = 2

    mask_v = librosa.util.softmask(S_full - S_filter,
                                   margin_v * S_filter,
                                   power=power)

    S_foreground = mask_v * S_full
    y_vocals = librosa.istft(S_foreground * phase)

    # Step 2: Save vocals to a temp file for Whisper
    with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as tmp:
        sf.write(tmp.name, y_vocals, sr)
        tmp_path = tmp.name

    try:
        # Step 3: Transcribe with Whisper
        model = whisper.load_model("base")
        result = model.transcribe(tmp_path, language=None)
        lyrics = result.get("text", "").strip()
    finally:
        os.unlink(tmp_path)

    return lyrics


@app.post("/separate")
async def separate_by_path(path: str = Form(...)):
    """Extract lyrics from a local audio file by separating vocals and transcribing."""
    try:
        lyrics = _extract_lyrics(path)
        return JSONResponse(content={"lyrics": lyrics})
    except FileNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Lyrics extraction failed: {e}")
