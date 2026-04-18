package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dhowden/tag"
	"github.com/spf13/cobra"

	"choinhaccli/internal/agent"
	"choinhaccli/internal/agent/providers"
	"choinhaccli/internal/analyzer"
	"choinhaccli/internal/audio"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6B9D")).PaddingLeft(1)
	sectionStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#C792EA")).PaddingLeft(1)
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#82AAFF")).Width(22).PaddingLeft(3)
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#C3E88D"))
	reviewStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#EEFFFF")).PaddingLeft(3).PaddingRight(3)
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B4252"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5370")).Bold(true)
)

var feelCmd = &cobra.Command{
	Use:   "feel <audio_file>",
	Short: "Let an AI agent listen to your music and share its feelings",
	Args:  cobra.ExactArgs(1),
	RunE:  runFeel,
}

var feelFlags struct {
	provider    string
	model       string
	lang        string
	analyzerURL string
}

func init() {
	feelCmd.Flags().StringVar(&feelFlags.provider, "provider", "openrouter", "LLM provider: openai, gemini, claude, openrouter")
	feelCmd.Flags().StringVar(&feelFlags.model, "model", "", "Model name (uses provider default if empty)")
	feelCmd.Flags().StringVar(&feelFlags.lang, "lang", "vi", "Output language: vi (Vietnamese), en (English)")
	feelCmd.Flags().StringVar(&feelFlags.analyzerURL, "analyzer-url", analyzer.DefaultURL, "Analyzer service URL")
}

func runFeel(_ *cobra.Command, args []string) error {
	filePath := args[0]

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ File not found: %s", filePath)))
		os.Exit(1)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".flac" && ext != ".wav" {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Unsupported format: %s (only .flac and .wav)", ext)))
		os.Exit(1)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	divider := dividerStyle.Render(strings.Repeat("─", 60))

	fmt.Println()
	fmt.Println(headerStyle.Render("🎵 AI Music Appreciation"))
	fmt.Println(divider)

	metadata := extractMetadata(absPath)
	printMetadata(metadata)
	fmt.Println(divider)

	fmt.Println(sectionStyle.Render("🔬 Analyzing audio..."))
	fmt.Println()

	client := analyzer.NewClient(feelFlags.analyzerURL)
	if err := client.HealthCheck(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Analyzer service not available at %s", feelFlags.analyzerURL)))
		fmt.Println(errorStyle.Render("  Start it with: cd analyzer && uvicorn main:app --port 8000"))
		os.Exit(1)
	}

	features, err := client.Analyze(absPath)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Analysis failed: %v", err)))
		os.Exit(1)
	}

	printFeatures(features)
	fmt.Println(divider)

	provider, err := createProvider(feelFlags.provider, feelFlags.model)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Provider error: %v", err)))
		os.Exit(1)
	}

	agentInstance := agent.NewAgent(provider, feelFlags.lang)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	spinnerCtx, spinnerCancel := context.WithCancel(context.Background())
	msg := sectionStyle.Render(fmt.Sprintf("🤖 Agent (%s) đang cảm nhận...", feelFlags.provider))

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5370")).Bold(true)
		i := 0
		for {
			select {
			case <-spinnerCtx.Done():
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r%s %s", msg, spinnerStyle.Render(frames[i]))
				i = (i + 1) % len(frames)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	response, err := agentInstance.Feel(ctx, features, metadata)
	spinnerCancel()
	time.Sleep(50 * time.Millisecond)

	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Agent error: %v", err)))
		os.Exit(1)
	}

	fmt.Println(msg + lipgloss.NewStyle().Foreground(lipgloss.Color("#C3E88D")).Bold(true).Render(" ✨ Đã xong!"))
	fmt.Println()
	fmt.Println(reviewStyle.Render(response))
	fmt.Println()
	fmt.Println(divider)
	fmt.Println()
	return nil
}

func extractMetadata(filePath string) *audio.TrackMetadata {
	meta := &audio.TrackMetadata{
		Title:  filepath.Base(filePath),
		Artist: "Unknown Artist",
		Album:  "Unknown Album",
	}

	f, err := os.Open(filePath)
	if err != nil {
		return meta
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return meta
	}

	if m.Title() != "" {
		meta.Title = m.Title()
	}
	if m.Artist() != "" {
		meta.Artist = m.Artist()
	}
	if m.Album() != "" {
		meta.Album = m.Album()
	}
	return meta
}

func printMetadata(m *audio.TrackMetadata) {
	fmt.Println(sectionStyle.Render("📋 Track Info"))
	fmt.Println(labelStyle.Render("Title:") + valueStyle.Render(m.Title))
	fmt.Println(labelStyle.Render("Artist:") + valueStyle.Render(m.Artist))
	fmt.Println(labelStyle.Render("Album:") + valueStyle.Render(m.Album))
}

func printFeatures(f *analyzer.AudioFeatures) {
	fmt.Println(sectionStyle.Render("📊 Audio Features"))
	fmt.Println(labelStyle.Render("BPM:") + valueStyle.Render(fmt.Sprintf("%.1f", f.BPM)))
	fmt.Println(labelStyle.Render("Key:") + valueStyle.Render(f.Key))
	fmt.Println(labelStyle.Render("Duration:") + valueStyle.Render(fmt.Sprintf("%.1fs", f.DurationSeconds)))
	fmt.Println(labelStyle.Render("Spectral Centroid:") + valueStyle.Render(fmt.Sprintf("%.0f Hz", f.SpectralCentroidMean)))
	fmt.Println(labelStyle.Render("RMS Energy:") + valueStyle.Render(fmt.Sprintf("%.4f", f.RMSEnergyMean)))
	fmt.Println(labelStyle.Render("Mood:") + valueStyle.Render(strings.Join(f.MoodKeywords, ", ")))
	fmt.Println()
}

func createProvider(name, model string) (agent.LLMProvider, error) {
	switch strings.ToLower(name) {
	case "openai":
		return providers.NewOpenAI(model)
	case "gemini":
		return providers.NewGemini(model)
	case "claude":
		return providers.NewClaude(model)
	case "openrouter":
		return providers.NewOpenRouter(model)
	default:
		return nil, fmt.Errorf("unknown provider %q (available: openai, gemini, claude, openrouter)", name)
	}
}
