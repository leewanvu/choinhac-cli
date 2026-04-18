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
│   └── player.go          # Package: audio
├── ui/
│   ├── model.go           # Package: ui
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
└── player.go          # Player, State, TrackMetadata, InitSpeaker, extractMetadata

internal/ui/
├── model.go           # Model, Init, Update, View (BubbleTea interface)
└── style.go           # Lipgloss color definitions

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

## Codebase Structure

### Directory Layout

```
musiccli/
├── cmd/musiccli/                   # Binary entry point
│   ├── main.go                     # func main() calls cmd.Execute()
│   └── cmd/
│       ├── root.go                 # Cobra root command
│       ├── play.go                 # play subcommand
│       └── feel.go                 # feel subcommand
├── internal/                       # Unexported packages
│   ├── audio/
│   │   └── player.go               # Audio engine
│   ├── ui/
│   │   ├── model.go                # TUI model
│   │   └── style.go                # Lipgloss styles
│   ├── agent/
│   │   ├── agent.go                # Agent orchestrator
│   │   ├── prompt.go               # Prompt builders
│   │   └── providers/
│   │       ├── openai.go
│   │       ├── gemini.go
│   │       ├── claude.go
│   │       └── openrouter.go
│   └── analyzer/
│       └── analyzer.go             # HTTP client for DSP
├── analyzer/                       # Python service
│   ├── main.py
│   ├── requirements.txt
│   └── README.md
├── go.mod
├── go.sum
├── README.md
└── docs/                           # This documentation
```

**Rule:** Use `internal/` for packages not intended as public API

### File Size Limits

| File Type | Limit | Guideline |
|-----------|-------|-----------|
| `.go` files | 300 LOC | Split if exceeds; one primary type per file |
| `.py` files | 500 LOC | Analyzer: one service per file |
| `.md` docs | 800 LOC | Split into subtopics if exceeds |

**Current Violations:** None

### Module Organization Rules

1. **One Responsibility Per Package**
   - `audio/` — Audio playback only
   - `ui/` — TUI rendering only
   - `agent/` — LLM orchestration only
   - `analyzer/` — HTTP client only

2. **No Circular Dependencies**
   - `cmd/` imports `internal/*`
   - `internal/ui` imports `internal/audio`
   - `internal/agent` imports `internal/analyzer` + `internal/audio`
   - No reverse imports

3. **Minimal Cross-Package State**
   - `cmd/` wires packages; minimal business logic
   - Each package manages its own state (no global singletons except `speakerInit`)

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

### Before Commit

- [ ] No syntax errors (compile with `go build ./...`)
- [ ] All exported types/functions have doc comments
- [ ] Error handling: all errors wrapped with context
- [ ] No hardcoded secrets (API keys, credentials)
- [ ] No circular package imports
- [ ] File size <300 LOC (split if exceeded)

### During Review

- [ ] Single responsibility per file/package
- [ ] Consistent naming (PascalCase exports, camelCase unexported)
- [ ] Error messages useful (actionable, not generic)
- [ ] Comments explain *why*, not *what* (code shows what)
- [ ] No dead code or commented-out logic
- [ ] Tests added for new functionality

### Performance & Safety

- [ ] No goroutine leaks (channels properly closed)
- [ ] Timeout on external HTTP calls (120s for analyzer)
- [ ] File handles closed (defer Close())
- [ ] Pointer receivers used for stateful types
- [ ] Interface{} used sparingly (prefer concrete types)

---

## Linting & Formatting

### Go Formatting

**Tool:** Built-in `gofmt` and `go fmt`

```bash
# Format current package
go fmt ./...

# Check all packages
gofmt -d ./internal/...
```

**Rule:** All code must pass `gofmt` before commit

### Import Organization

**Pattern:** stdlib → external → internal

```go
import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gopxl/beep"
    "github.com/spf13/cobra"

    "choinhaccli/internal/audio"
    "choinhaccli/internal/ui"
)
```

### Linter Goals

**Target (not enforced yet):**
- No unused variables or imports
- No shadowing of variables
- Consistent receiver names (use `p` for `*Player`, `m` for `Model`, etc.)

---

## Documentation Standards

### README.md

- **Location:** `/Users/vule/Work/musiccli/README.md`
- **Sections:** Features, architecture, installation, usage, controls, links to docs
- **Style:** Concise, example-driven

### API Documentation

- **Type Documentation:** Doc comments on exported types
- **Function Documentation:** Describe parameters, return values, error cases
- **Example Comments:** Include usage snippets for complex functions

### Docs Directory Structure

```
docs/
├── README.md                      # (Links to all docs)
├── project-overview-pdr.md        # Vision, requirements, roadmap
├── codebase-summary.md            # Package-by-package summary
├── code-standards.md              # This file
├── system-architecture.md         # Architecture diagrams, data flows
└── project-roadmap.md             # Milestones, timeline, progress
```

**Rule:** Keep doc files under 800 LOC; split into subtopics if exceeded

---

## Commit Message Format

**Pattern:** Conventional commits with semantic prefixes

```
feat: add random track navigation
fix: prevent goroutine leak in UI poll loop
docs: update architecture diagram with DSP flow
refactor: extract prompt builder to separate file
test: add unit tests for Player.Next()
chore: update go.mod dependencies
```

**Rules:**
- No "AI" references (no "AI-generated", "generated by Claude", etc.)
- Lowercase imperative mood ("add", not "adds", "added")
- First line <70 chars
- Link to issue if applicable (Fixes #123)
- Include breaking changes in footer (BREAKING CHANGE: ...)

---

## Environment Variables

### Analyzer Service

```bash
# Python FastAPI service (feel command)
ANALYZER_URL=http://localhost:8000     # Default
```

### LLM Provider Keys

```bash
# OpenAI
OPENAI_API_KEY=sk-...

# Google Gemini
GEMINI_API_KEY=AIzaSy...

# Anthropic Claude
CLAUDE_API_KEY=sk-ant-...

# OpenRouter
OPENROUTER_API_KEY=sk-or-...
```

**Rule:** Never commit .env files; use .env.example as template

---

## Performance Optimization Guidelines

### Audio Processing

- **Buffering:** Use beep's default buffer size (441 frames @ 44.1kHz = ~10ms)
- **Resampling:** Built into beep; no manual optimization needed
- **Memory:** Streaming decoder keeps <1MB in memory at a time

### UI Rendering

- **Polling Interval:** 100ms (trade-off: responsiveness vs. CPU)
- **String Building:** Use `strings.Builder` for large concatenations
- **Style Caching:** Lipgloss styles defined once at module level

### LLM Requests

- **Timeout:** 30-60s for API calls (provider-specific)
- **Retries:** None implemented (MVP); add exponential backoff if needed
- **Context Cancellation:** Use `context.WithTimeout()` for all API calls

---

## Security Considerations

### File Path Handling

```go
// SAFE: use absolute path resolution
absPath, err := filepath.Abs(filePath)
if err != nil {
    return fmt.Errorf("invalid path: %w", err)
}
// Avoid path traversal with ../ or symlinks
```

### API Key Management

```go
// SAFE: read from env vars only
apiKey := os.Getenv("OPENAI_API_KEY")
if apiKey == "" {
    return fmt.Errorf("OPENAI_API_KEY not set")
}

// NEVER hardcode keys, commit .env files, or log keys
```

### HTTP Requests

```go
// SAFE: set timeout to prevent hanging
client := &http.Client{
    Timeout: 120 * time.Second,
}
```

### Error Messages

```go
// SAFE: Don't expose absolute paths in user-facing errors
fmt.Println("Error: unable to load audio")

// UNSAFE: Leaks file system structure
fmt.Println("Error: /Users/vule/Work/musiccli/tracks/lost.flac not found")
```

