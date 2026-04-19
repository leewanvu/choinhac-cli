# Code Standards & Conventions

**Project:** musiccli (choinhaccli)  
**Language:** Go 1.25  
**Last Updated:** 2026-04-18

## Go Conventions

### File & Package Naming

**Pattern:** `snake_case.go` for files; lowercase package names

```
internal/
├── audio/
│   ├── player.go          # Package: audio
│   └── amplitude_tracker.go # Package: audio
├── ui/
│   ├── model.go           # Package: ui
│   ├── view.go            # Package: ui
│   ├── album_art.go       # Package: ui
│   ├── visualizer.go      # Package: ui
│   └── style.go           # Package: ui
├── agent/
│   ├── agent.go           # Package: agent
│   ├── prompt.go          # Package: agent
│   └── providers/
│       ├── openai.go      # Package: providers
│       ├── gemini.go      # Package: providers
│       ├── claude.go      # Package: providers
│       └── openrouter.go  # Package: providers
└── analyzer/
    └── analyzer.go        # Package: analyzer
```

**Rule:** File name describes primary type/function (e.g., `player.go` contains `Player` type)

### Exported Types & Functions

**Pattern:** `PascalCase` for exported symbols; unexported helpers in `camelCase`

```go
// Exported type
type Player struct {
    ctrl *beep.Ctrl
    state State
}

// Exported method
func (p *Player) Play() error { }

// Exported function
func NewPlayer() *Player { }

// Exported constant
const StatePlaying State = iota

// Unexported helper
func (p *Player) closeStreamer() error { }
```

### Documentation Comments

**Rule:** Every exported type, function, constant, and variable has a doc comment (first line no period for single-line comments)

```go
// Player manages audio playback and playlist navigation
type Player struct { }

// NewPlayer creates a new player instance
func NewPlayer() *Player { }

// Play resumes playback from pause state
func (p *Player) Play() error { }

// State represents the playback state
type State int

const (
    StateStopped State = iota
    StatePlaying
    StatePaused
)
```

**Multi-line comments:**
```go
// Analyze sends the audio file to the Python DSP service,
// extracts features, and returns structured JSON.
// Returns an error if the service is unreachable.
func (c *Client) Analyze(filePath string) (*AudioFeatures, error) { }
```

### Error Handling

**Pattern:** Use `fmt.Errorf()` with `%w` for wrapping; add context to all returns

```go
// GOOD: wrapping with context
if err := speaker.Init(rate, bufferSize); err != nil {
    return fmt.Errorf("failed to init speaker: %w", err)
}

// BAD: naked return
if err != nil {
    return err
}

// BAD: generic message
if err != nil {
    return fmt.Errorf("error: %w", err)
}
```

**Pattern:** Check errors immediately after operations

```go
file, err := os.Open(path)
if err != nil {
    return fmt.Errorf("cannot open audio file: %w", err)
}
defer file.Close()
```

### Struct Layout

**Pattern:** Related fields grouped; zero values safe; channels/mutexes last

```go
type Player struct {
    // Beep components
    ctrl     *beep.Ctrl
    volume   *effects.Volume
    streamer beep.StreamSeekCloser
    format   beep.Format

    // State
    state       State
    Metadata    TrackMetadata
    Playlist    []string
    PlaylistIdx int

    // Channels
    done chan bool
}
```

**Rule:** Mutex/channel fields last (signals resource ownership)

### Interface Design

**Pattern:** Small, focused interfaces; embed standard library interfaces

```go
// LLMProvider defines the interface for LLM backends
type LLMProvider interface {
    Name() string
    Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// Composed interfaces acceptable
type Reader interface {
    io.Reader
    Seek(offset int64, whence int) (int64, error)
}
```

**Rule:** Interfaces with 1-3 methods; avoid "God" interfaces

### Pointer vs. Value Receivers

**Pattern:** Use pointer receivers for:
- Methods that modify receiver state
- Large structs (>128 bytes)
- Embedded mutexes/channels

Use value receivers for:
- Small immutable types
- Structs used as keys (map, set)

```go
// Pointer: modifies state
func (p *Player) Play() error { p.state = StatePlaying; ... }

// Value: small, immutable
func (m TrackMetadata) String() string { return m.Title }
```

### Package Organization

**Pattern:** One responsibility per package; public interface in root file

```
internal/audio/
├── player.go          # Player, State, TrackMetadata, InitSpeaker, extractMetadata
└── amplitude_tracker.go # amplitudeTracker type + newAmplitudeTracker constructor

internal/ui/
├── model.go           # Model, Init, Update (BubbleTea interface)
├── view.go            # View() rendering, layout helpers, visualizer update
├── album_art.go       # renderArt, album art decoding
├── visualizer.go      # visualizer type, update, render
└── style.go           # Catppuccin Mocha color definitions

internal/agent/
├── agent.go           # Agent, LLMProvider interface
├── prompt.go          # buildSystemPrompt, buildUserPrompt
└── providers/         # Separate package for implementations
    ├── openai.go
    ├── gemini.go
    ├── claude.go
    └── openrouter.go
```

**Rule:** Interfaces in package root; implementations in `providers` sub-package

---

## Codebase Structure & Module Organization

**Structure:** `cmd/musiccli/` → `internal/{audio,ui,agent,analyzer,library,config,web}` + `web/` (React SPA) + `analyzer/` (Python)

**Module Rules:**
1. One responsibility per package (audio → playback, ui → TUI, agent → LLM)
2. No circular dependencies (cmd imports internal; internal may cross)
3. Use `internal/` for non-public packages

**File Size Limits:** Go <300 LOC, Python <500 LOC, docs <800 LOC

---

## Common Patterns

### Audio Playback Flow

```go
// 1. Initialize speaker (once)
if err := audio.InitSpeaker(); err != nil {
    return err
}

// 2. Create player
p := audio.NewPlayer()

// 3. Load file
if err := p.LoadAndPlay("track.flac"); err != nil {
    return err
}

// 4. Poll state / respond to UI events
status := p.GetState()
pos := p.GetPosition()
p.TogglePause()  // or p.Next(), p.VolumeUp(), etc.
```

### TUI Control Loop (BubbleTea)

```go
// Model.Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "space":
            m.player.TogglePause()
        case "q":
            return m, tea.Quit
        }
    case tickMsg:
        return m, m.tickCmd()  // reschedule timer
    }
    return m, nil
}

// Model.View renders
func (m Model) View() string {
    title := titleStyle.Render("🎵 CLI Music Player")
    // ... build view string
    return title + metadata + progress + help
}
```

### LLM Provider Implementation

```go
type OpenAIProvider struct {
    apiKey string
    model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
    if model == "" {
        model = "gpt-3.5-turbo"
    }
    return &OpenAIProvider{
        apiKey: apiKey,
        model:  model,
    }
}

func (p *OpenAIProvider) Name() string {
    return "openai"
}

func (p *OpenAIProvider) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
    // Call OpenAI API
    // Return model response
    return "", nil
}
```

### HTTP Client Pattern

```go
type Client struct {
    baseURL    string
    httpClient *http.Client
}

func NewClient(baseURL string) *Client {
    if baseURL == "" {
        baseURL = DefaultURL
    }
    return &Client{
        baseURL: strings.TrimRight(baseURL, "/"),
        httpClient: &http.Client{
            Timeout: 120 * time.Second,
        },
    }
}

func (c *Client) HealthCheck() error {
    resp, err := c.httpClient.Get(c.baseURL + "/health")
    if err != nil {
        return fmt.Errorf("service unreachable: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("service unhealthy: status %d", resp.StatusCode)
    }
    return nil
}
```

---

## Cobra CLI Patterns

### Subcommand Definition

```go
var playCmd = &cobra.Command{
    Use:   "play <path>",
    Short: "Play an audio file or directory",
    Long:  "Play a .flac or .wav file, or all supported files in a directory.",
    Args:  cobra.ExactArgs(1),
    RunE:  runPlay,
}

func runPlay(cmd *cobra.Command, args []string) error {
    path := args[0]
    // Implementation
    return nil
}

func init() {
    rootCmd.AddCommand(playCmd)
}
```

### Flag Handling

```go
var flags struct {
    provider    string
    model       string
    lang        string
    analyzerURL string
}

func init() {
    feelCmd.Flags().StringVar(&flags.provider, "provider", "openrouter", "LLM provider")
    feelCmd.Flags().StringVar(&flags.lang, "lang", "vi", "Output language")
}

func runFeel(cmd *cobra.Command, args []string) error {
    // Use flags.provider, flags.lang
    return nil
}
```

---

## Python Conventions (Analyzer Service)

### File & Function Naming

```python
# snake_case for functions, variables
def detect_key(chroma: np.ndarray) -> str:
    pitch_classes = ["C", "C#", ...]
    return best_key

# PascalCase for classes
class AudioFeatureExtractor:
    pass

# UPPER_SNAKE_CASE for constants
DEFAULT_SAMPLE_RATE = 22050
BUFFER_SIZE = 4096
```

### Type Hints

```python
# All function signatures include type hints
def extract_features(file_path: str) -> dict[str, Any]:
    """Extract DSP features from audio file."""
    pass

# Return types annotated
def detect_key(chroma: np.ndarray) -> str:
    pass

# Container types specific
from typing import Any
features: dict[str, Any] = {}
moods: list[str] = []
```

### FastAPI Endpoint Pattern

```python
from fastapi import FastAPI, HTTPException, Form

app = FastAPI(title="Music Analyzer", version="1.0.0")

@app.get("/health")
def health_check() -> dict[str, str]:
    return {"status": "ok"}

@app.post("/analyze")
async def analyze_audio(path: str = Form(...)) -> dict[str, Any]:
    """Analyze audio features from file path."""
    try:
        features = extract_features(path)
        return features
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
```

---

## TypeScript & React Conventions

### File & Component Naming

**Pattern:** `kebab-case.tsx` for React components; `camelCase.ts` for utilities

```
web/src/
├── pages/
│   ├── library.tsx              # React page component
│   ├── search.tsx
│   ├── album-detail.tsx
│   └── artist-detail.tsx
├── components/
│   ├── now-playing-bar.tsx      # Self-contained UI component
│   ├── track-row.tsx
│   ├── cover-image.tsx
│   ├── queue-drawer.tsx
│   └── add-to-playlist-dialog.tsx
├── hooks/
│   ├── use-keyboard-shortcuts.ts # Custom React hook (camelCase)
│   └── use-playlists.ts
├── store/
│   ├── player.ts                # Zustand store
│   └── ui.ts
├── api/
│   └── client.ts                # API client
└── audio/
    └── engine.ts                # Audio engine wrapper
```

**Rule:** Export component as default; use filename as component name

### React Component Structure

**Pattern:** Functional component with TypeScript interface for props

```typescript
// Components with typed props
interface TrackRowProps {
  track: Track;
  isPlaying: boolean;
  onPlay: (track: Track) => void;
  onAddToPlaylist: (track: Track) => void;
}

export default function TrackRow({
  track,
  isPlaying,
  onPlay,
  onAddToPlaylist,
}: TrackRowProps) {
  return (
    <div className="track-row">
      {/* JSX */}
    </div>
  );
}

// Hooks call must be at component root (not in callbacks)
// State management via Zustand, not useState for shared data
```

**Rules:**
- Props destructured with TypeScript interface
- Default export for page/major components
- Named exports for shared utilities
- Hooks called only at component root (BubbleTea update rule)

### Zustand Store Pattern

```typescript
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface PlayerStore {
  queue: Track[];
  currentIndex: number;
  volume: number;
  setQueue: (queue: Track[]) => void;
  setCurrentIndex: (idx: number) => void;
  setVolume: (vol: number) => void;
}

export const usePlayerStore = create<PlayerStore>()(
  persist(
    (set) => ({
      queue: [],
      currentIndex: 0,
      volume: 1.0,
      setQueue: (queue) => set({ queue }),
      setCurrentIndex: (idx) => set({ currentIndex: idx }),
      setVolume: (vol) => set({ volume: vol }),
    }),
    {
      name: 'player-storage', // localStorage key
    }
  )
);
```

**Rules:**
- Store actions use `set()` to update state
- Type interface for all store state
- Use `persist` middleware for data that survives page refresh
- Selectors use `usePlayerStore((state) => state.field)`

### Inline Styles & CSS

**Pattern:** React.CSSProperties for inline styles (no CSS framework)

```typescript
import React from 'react';

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
    padding: '16px',
    backgroundColor: '#282828',
    color: '#1db954',
  },
  button: {
    padding: '8px 16px',
    backgroundColor: '#1db954',
    color: '#000',
    border: 'none',
    cursor: 'pointer',
    borderRadius: '4px',
  },
};

export default function MyComponent() {
  return <div style={styles.container}>Content</div>;
}
```

**Rules:**
- All styles defined as inline objects
- camelCase property names (cssProperty → cssProperty in JS)
- Color palette: dark #282828, accent #1db954 (Spotify green), text #fff
- Responsive: use media queries via window.matchMedia(), not CSS breakpoint classes

### API Client Pattern

```typescript
// client.ts — type-safe fetch wrapper
export interface TracksResponse {
  tracks: Track[];
  total: number;
  limit: number;
  offset: number;
}

export async function getTracks(limit = 20, offset = 0): Promise<TracksResponse> {
  const res = await fetch(`/api/library/tracks?limit=${limit}&offset=${offset}`);
  if (!res.ok) throw new Error(`API error: ${res.statusCode}`);
  return res.json();
}

export async function getAlbum(id: number): Promise<Album & { tracks: Track[] }> {
  const res = await fetch(`/api/library/albums/${id}`);
  if (!res.ok) throw new Error(`Failed to fetch album: ${res.statusCode}`);
  return res.json();
}
```

**Rules:**
- One function per API endpoint
- Return types explicitly typed
- Error handling: throw on non-2xx status
- Query params built via URL constructor (not string concat)

### TypeScript Type Conventions

**Pattern:** Interfaces for data, strict null checking enabled

```typescript
// Data types
interface Track {
  id: number;
  title: string;
  artist: string;
  album: string;
  duration: number;
  filePath: string;
  format: 'flac' | 'wav' | 'mp3'; // discriminated union
}

interface Album {
  id: number;
  title: string;
  artist: string;
  year?: number; // optional
  coverPath?: string;
}

// Utility types
type TrackMap = Record<number, Track>;
type Nullable<T> = T | null;

// Function signatures
function filterByArtist(tracks: Track[], artist: string): Track[] {
  return tracks.filter((t) => t.artist === artist);
}

// No `any` except in catch blocks for error handling
function handleError(err: unknown) {
  const message = err instanceof Error ? err.message : String(err);
  console.error(message);
}
```

**Rules:**
- `interface` for object types, not `type`
- Optional fields marked with `?`
- Use discriminated unions for variants (e.g., `format: 'flac' | 'wav'`)
- No `any`; use `unknown` then narrow with `instanceof` or type guards

---

## Testing Guidelines (To Be Implemented)

### Unit Test Structure

```go
// player_test.go
package audio

import "testing"

func TestPlayerNewPlayer(t *testing.T) {
    p := NewPlayer()
    if p.state != StateStopped {
        t.Errorf("expected StateStopped, got %d", p.state)
    }
}

func TestPlayerLoadAndPlay(t *testing.T) {
    p := NewPlayer()
    err := p.LoadAndPlay("testdata/sample.flac")
    if err != nil {
        t.Fatalf("LoadAndPlay failed: %v", err)
    }
    if p.state != StatePlaying {
        t.Errorf("expected StatePlaying after load")
    }
}
```

### Test Files Needed

| File | Focus | Status |
|------|-------|--------|
| `internal/audio/player_test.go` | Load, play, pause, next, prev, volume, metadata | ✗ TODO |
| `internal/agent/agent_test.go` | Agent.Feel, prompt building | ✗ TODO |
| `internal/analyzer/analyzer_test.go` | HTTP client, health check (mock service) | ✗ TODO |
| `cmd/musiccli/cmd/play_test.go` | CLI arg parsing, directory handling | ✗ TODO |
| `cmd/musiccli/cmd/feel_test.go` | CLI flags, LLM provider selection | ✗ TODO |

### Coverage Target

- **Overall:** 70%
- **Critical paths:** >80% (audio playback, error handling)
- **UI/rendering:** >50% (harder to unit test)
- **LLM providers:** >60% (mock API responses)

---

## Code Review Checklist

| Category | Item |
|----------|------|
| **Compilation** | No syntax errors (`go build ./...`) |
| **Documentation** | Exported types/functions have doc comments |
| **Error Handling** | All errors wrapped with context (`fmt.Errorf`) |
| **Security** | No hardcoded secrets (API keys, credentials) |
| **Structure** | No circular imports; <300 LOC per file |
| **Naming** | PascalCase exports, camelCase unexported, descriptive |
| **Comments** | Explain *why*, not *what*; code shows what |
| **Performance** | No goroutine leaks; timeouts on HTTP calls (120s) |
| **Resource Mgmt** | File handles closed (`defer Close()`) |
| **Types** | Pointer receivers for stateful; avoid `interface{}` |

---

## Linting & Formatting

**Go Formatting:** Use `go fmt ./...` before commit (stdlib → external → internal imports)

**Linter Targets:** No unused imports, consistent receiver names (`p` for `*Player`, `m` for `Model`)

## Documentation Standards

- **Type Docs:** Doc comments on exported types, functions, constants
- **Examples:** Include usage snippets for complex functions
- **Files:** Keep under 800 LOC; split if exceeded

---

## Commit Message Format

**Pattern:** `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:` (lowercase, <70 chars, no AI references)

## Environment Variables & Security

**LLM API Keys:** `OPENAI_API_KEY`, `GEMINI_API_KEY`, `CLAUDE_API_KEY`, `OPENROUTER_API_KEY` (env vars only; never hardcode)

**Analyzer Service:** `ANALYZER_URL=http://localhost:8000` (default)

**Security Rules:**
- Never commit `.env` files
- File paths: use `filepath.Abs()` to prevent traversal attacks
- HTTP calls: set 120s timeout (DoS protection)
- Error messages: never expose absolute paths
- Tokens: read from env vars only; log nothing with secrets

