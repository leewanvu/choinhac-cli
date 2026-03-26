// Package agent provides the AI music appreciation agent
// with multi-provider LLM support.
package agent

import (
	"context"
	"fmt"

	"choinhaccli/internal/analyzer"
	"choinhaccli/internal/audio"
)

// LLMProvider defines the interface for LLM backends.
type LLMProvider interface {
	// Name returns the provider name (e.g. "openai", "gemini", "claude").
	Name() string
	// Chat sends a system prompt and user prompt, returns the model response.
	Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// Agent orchestrates the music appreciation flow.
type Agent struct {
	provider LLMProvider
	lang     string
}

// NewAgent creates a new Agent with the given LLM provider and output language.
func NewAgent(provider LLMProvider, lang string) *Agent {
	if lang == "" {
		lang = "vi"
	}
	return &Agent{
		provider: provider,
		lang:     lang,
	}
}

// Feel generates a music appreciation response from audio features and metadata.
func (a *Agent) Feel(ctx context.Context, features *analyzer.AudioFeatures, metadata *audio.TrackMetadata) (string, error) {
	systemPrompt := buildSystemPrompt(a.lang)
	userPrompt := buildUserPrompt(features, metadata)

	response, err := a.provider.Chat(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("LLM (%s) failed: %w", a.provider.Name(), err)
	}

	return response, nil
}
