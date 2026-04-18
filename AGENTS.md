# AGENTS.md - Music CLI Player Development Guidelines

## Overview
This document provides guidelines for agentic coding agents working on the musiccli repository. It covers build/test/lint commands, code style, and best practices.

## Project Structure
- `cmd/` - Application entry points (player, feel)
- `internal/` - Private application code
  - `audio/` - Audio playback functionality
  - `ui/` - Terminal user interface using BubbleTea
  - `agent/` - AI agent providers
  - `analyzer/` - Music analysis tools (Python-based)
- `go.mod` - Go module definition

## Build & Run Commands

### Building
```bash
# Build binaries to build/ directory
go build -o build/musicplayer ./cmd/player
go build -o build/musicfeel ./cmd/feel
```

### Running
```bash
# Run the player (from build directory)
./build/musicplayer <path_to_audio_file_or_directory>

# Run the feel command (from build directory)
./build/musicfeel
```

### Testing
```bash
# Run all tests (currently no tests exist in main project)
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test package
go test ./internal/audio

# Run tests with coverage
go test -cover ./...
```

### Linting & Formatting
```bash
# Format code with gofmt
go fmt ./...

# Check for formatting issues
gofmt -d .

# Install and run golangci-lint (recommended)
# go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run

# Vet for suspicious constructs
go vet ./...

# Check for security issues
# go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Code Style Guidelines

### General Formatting
- Use `gofmt` as the canonical formatter (run `go fmt ./...` before committing)
- Line length: Aim for 80-100 characters, but prioritize readability
- Indentation: Use tabs for indentation (Go standard)
- Blank lines: Use blank lines to separate logical sections in functions

### Imports
- Group imports: Standard library first, then third-party, then local
- Use alphabetical ordering within each group
- Avoid dot imports
- Example:
```go
import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/charmbracelet/bubbletea"
    "github.com/dhowden/tag"

    "choinhaccli/internal/audio"
)
```

### Naming Conventions
- Use MixedCaps for exported names (PascalCase)
- Use camelCase for unexported names
- Package names should be lowercase and single-word
- Interface names ending in "-er" (Reader, Writer, Player)
- Constants in mixedCaps or ALL_CAPS for enum-like values
- Acronyms: Use consistent case (e.g., "id" or "ID" but be consistent)

### Types & Structures
- Define structs with clear, descriptive names
- Use embedded fields when appropriate for composition
- Keep zero values meaningful
- Use constructor functions (New*) for complex initialization
- Example:
```go
type Player struct {
    ctrl     *beep.Ctrl
    volume   *effects.Volume
    streamer beep.StreamSeekCloser
    // ... other fields
}
```

### Error Handling
- Handle errors explicitly; don't ignore them
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Return early on errors to avoid deep nesting
- Use sentinel errors for predictable error conditions
- Example:
```go
if err := audio.InitSpeaker(); err != nil {
    fmt.Printf("Failed to initialize audio speaker: %v\n", err)
    os.Exit(1)
}
```

### Concurrency
- Prefer channels over shared state for goroutine communication
- Use context package for cancellation and timeouts
- Avoid exposing channel internals in APIs when possible
- Properly clean up resources (close channels, etc.)

### Comments
- Write clear, concise comments
- Comment exported functions, types, and constants
- Use sentence-style comments that start with a capital letter and end with a period
- Avoid obvious comments; explain why, not what
- Example:
```go
// InitSpeaker initializes the global audio speaker. Must be called once.
func InitSpeaker() error {
    // ... implementation
}
```

### Logging & Output
- Use fmt.Printf for user-facing output in CLI applications
- Format error messages clearly and actionably
- For logging, consider using a proper logger in future enhancements
- Keep user messages helpful and concise

## Specific Patterns in This Codebase

### Audio Package Patterns
- Use `speaker.Lock()`/`speaker.Unlock()` around speaker operations
- Initialize global state once with boolean flags
- Extract metadata separately from decoding
- Use channels for signaling completion (`done chan bool`)

### UI Package Patterns
- Separate model (state) from view (rendering)
- Use BubbleTea's Msg system for event handling
- Helper functions for formatting (duration, progress bars)
- Style definitions centralized in style.go

### Error Messages
- User-facing errors: Clear, actionable messages
- Internal errors: Wrapped with context using `%w`
- Exit with non-zero code on failure in main()

## Dependency Management
- Use Go modules (`go.mod` tracks dependencies)
- Vendoring not used; rely on module proxy
- Keep dependencies updated: `go get -u ./...`
- Check for vulnerabilities: `govulncheck ./...`

## AI Agent Guidelines
When working with this codebase:
1. Always run `go fmt ./...` before considering code complete
2. Test changes by building and running the affected binaries
3. Follow existing patterns in the codebase for consistency
4. Pay attention to error handling patterns
5. Keep user experience in mind for CLI interactions
6. When adding features, consider both player and feel commands
7. Respect the separation between internal packages

## Future Considerations
- Add comprehensive tests as features are added
- Consider implementing configuration files
- Explore additional audio format support
- Add more sophisticated UI elements
- Consider cross-compilation for different platforms
