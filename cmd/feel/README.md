# `feel` - AI Music Appreciation Agent

The `feel` command is an AI-powered music appreciation feature for the `choinhaccli` music player. It listens to your audio files (`.wav`, `.flac`), analyzes their digital signal processing (DSP) features such as tempo, key, and mood, and then uses a Large Language Model (LLM) to write a subjective, emotional review of the track.

## How it Works

1. **Metadata Extraction:** Extracts basic tags like Title, Artist, and Album from the audio file using Go.
2. **Audio Analysis:** Sends the absolute file path to the local Python **[Analyzer Service](../../analyzer/README.md)** via HTTP. The analyzer extracts features like BPM, key, energy, and mood using `librosa`.
3. **AI Generation:** Combines the metadata and extracted audio features, and prompts an LLM provider (Gemini, Claude, OpenAI, or OpenRouter) to generate an emotional appreciation review.
4. **Lyrics Extraction (Optional):** If the `--separate` flag is used, the command will separate vocals and transcribe them into text using Whisper.

## Requirements

1. **Analyzer Service:** The Python Analyzer service must be running locally.
   ```bash
   cd analyzer
   source .venv/bin/activate
   uvicorn main:app --port 8000
   ```
2. **LLM API Key:** You must have the corresponding API key exported in your environment for the chosen provider:
   - Gemini: `export GEMINI_API_KEY="your-api-key"`
   - OpenAI: `export OPENAI_API_KEY="your-api-key"`
   - Claude: `export ANTHROPIC_API_KEY="your-api-key"`
   - OpenRouter: `export OPENROUTER_API_KEY="your-api-key"` (Optional: `export OPENROUTER_MODEL="meta-llama/llama-3.3-70b-instruct:free"`)

## Usage

```bash
go run cmd/feel/main.go [flags] <audio_file>
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--provider` | `openrouter` | The LLM provider to use (`gemini`, `openai`, `claude`, `openrouter`). |
| `--model` | *(Default model for provider)* | The specific LLM model version you want to query. |
| `--lang` | `vi` | Output language for the AI's response (`vi` for Vietnamese, `en` for English). |
| `--analyzer-url` | `http://localhost:8000` | The URL where the Python Analyzer Service is running. |
| `--separate` | `false` | If true, extracts and displays lyrics by separating vocals and transcribing with Whisper. |

### Examples

**Basic usage (defaults to OpenRouter and Vietnamese):**
```bash
go run cmd/feel/main.go ~/Music/track.wav
```

**Use a specific provider and output in English:**
```bash
go run cmd/feel/main.go --provider claude --lang en ~/Music/track.flac
```

**Extract lyrics before analysis:**
```bash
go run cmd/feel/main.go --separate ~/Music/track.wav
```

**Specify a custom model and the Analyzer's exact URL:**
```bash
go run cmd/feel/main.go --provider openai --model gpt-4o --analyzer-url "http://127.0.0.1:8000" ~/Music/track.wav
```

## Supported Formats

- `.wav`
- `.flac`
