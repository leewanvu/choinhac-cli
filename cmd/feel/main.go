package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dhowden/tag"

	"choinhaccli/internal/agent"
	"choinhaccli/internal/agent/providers"
	"choinhaccli/internal/analyzer"
	"choinhaccli/internal/audio"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B9D")).
			PaddingLeft(1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#C792EA")).
			PaddingLeft(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#82AAFF")).
			Width(22).
			PaddingLeft(3)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C3E88D"))

	reviewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EEFFFF")).
			PaddingLeft(3).
			PaddingRight(3)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B4252"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5370")).
			Bold(true)
)

func main() {
	providerName := flag.String("provider", "openrouter", "LLM provider: openai, gemini, claude, openrouter")
	model := flag.String("model", "", "Model name (uses provider default if empty)")
	lang := flag.String("lang", "vi", "Output language: vi (Vietnamese), en (English)")
	analyzerURL := flag.String("analyzer-url", analyzer.DefaultURL, "Analyzer service URL")
	separate := flag.Bool("separate", false, "Separate vocals from accompaniment")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: feel [flags] <audio_file>\n\n")
		fmt.Fprintf(os.Stderr, "Let an AI agent listen to your music and share its feelings.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  feel song.flac\n")
		fmt.Fprintf(os.Stderr, "  feel --separate song.flac\n")
		fmt.Fprintf(os.Stderr, "  feel --provider openai song.wav\n")
		fmt.Fprintf(os.Stderr, "  feel --provider claude --lang en track.flac\n")
		fmt.Fprintf(os.Stderr, "  feel --provider openrouter song.flac\n")
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	filePath := flag.Arg(0)

	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ File not found: %s", filePath)))
		os.Exit(1)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".flac" && ext != ".wav" {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Unsupported format: %s (only .flac and .wav)", ext)))
		os.Exit(1)
	}

	// Absolute path for the analyzer service
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to resolve path: %v", err)))
		os.Exit(1)
	}

	divider := dividerStyle.Render(strings.Repeat("─", 60))

	// ── Step 1: Extract metadata (Go side) ──
	fmt.Println()
	fmt.Println(headerStyle.Render("🎵 AI Music Appreciation"))
	fmt.Println(divider)

	metadata := extractMetadata(absPath)
	printMetadata(metadata)
	fmt.Println(divider)

	// ── Step 2: Analyze audio via Python service ──
	fmt.Println(sectionStyle.Render("🔬 Analyzing audio..."))
	fmt.Println()

	client := analyzer.NewClient(*analyzerURL)
	if err := client.HealthCheck(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Analyzer service not available at %s", *analyzerURL)))
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

	// ── Step 2.5: Extract lyrics ──
	if *separate {
		fmt.Println(sectionStyle.Render("✂️ Extracting lyrics..."))
		fmt.Println()

		sepCtx, sepCancel := context.WithCancel(context.Background())
		sepMsg := sectionStyle.Render("🎵 Đang tách lời bài hát...")

		go func() {
			frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5370")).Bold(true)
			i := 0
			for {
				select {
				case <-sepCtx.Done():
					fmt.Print("\r\033[K")
					return
				default:
					fmt.Printf("\r%s %s", sepMsg, spinnerStyle.Render(frames[i]))
					i = (i + 1) % len(frames)
					time.Sleep(80 * time.Millisecond)
				}
			}
		}()

		lyrics, err := client.Separate(absPath)
		sepCancel()
		time.Sleep(50 * time.Millisecond)

		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Lyrics extraction failed: %v", err)))
		} else {
			fmt.Println(sepMsg + lipgloss.NewStyle().Foreground(lipgloss.Color("#C3E88D")).Bold(true).Render(" ✨ Đã xong!"))
			fmt.Println()
			fmt.Println(sectionStyle.Render("📝 Lyrics"))
			fmt.Println(reviewStyle.Render(lyrics))
			fmt.Println()
		}
		fmt.Println(divider)
	}

	// ── Step 3: Get AI appreciation ──
	provider, err := createProvider(*providerName, *model)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Provider error: %v", err)))
		os.Exit(1)
	}

	agentInstance := agent.NewAgent(provider, *lang)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Run spinner in background
	spinnerCtx, spinnerCancel := context.WithCancel(context.Background())
	msg := sectionStyle.Render(fmt.Sprintf("🤖 Agent (%s) đang cảm nhận...", *providerName))
	
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5370")).Bold(true)
		i := 0
		for {
			select {
			case <-spinnerCtx.Done():
				fmt.Print("\r\033[K") // Clear the spinner line
				return
			default:
				fmt.Printf("\r%s %s", msg, spinnerStyle.Render(frames[i]))
				i = (i + 1) % len(frames)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	response, err := agentInstance.Feel(ctx, features, metadata)
	
	// Stop spinner
	spinnerCancel()
	time.Sleep(50 * time.Millisecond) // Give goroutine a moment to print clear line

	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Agent error: %v", err)))
		os.Exit(1)
	}

	// Print a nice completion status
	fmt.Println(msg + lipgloss.NewStyle().Foreground(lipgloss.Color("#C3E88D")).Bold(true).Render(" ✨ Đã xong!"))
	fmt.Println()

	fmt.Println(reviewStyle.Render(response))
	fmt.Println()
	fmt.Println(divider)
	fmt.Println()
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
