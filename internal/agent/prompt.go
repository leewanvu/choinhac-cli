package agent

import (
	"fmt"
	"strings"

	"choinhaccli/internal/analyzer"
	"choinhaccli/internal/audio"
)

func buildSystemPrompt(lang string) string {
	langInstruction := "Hãy viết cảm nhận bằng tiếng Việt."
	if lang == "en" {
		langInstruction = "Write your appreciation in English."
	}

	return fmt.Sprintf(`Bạn là một nhà phê bình âm nhạc đầy cảm xúc và tinh tế. Bạn có khả năng cảm nhận sâu sắc về âm nhạc dựa trên các đặc điểm kỹ thuật và metadata của bài hát.

Nhiệm vụ của bạn:
- Nhận thông tin phân tích kỹ thuật (BPM, key, spectral features, energy profile, v.v.) và metadata (title, artist, album) của một bài hát.
- Viết một bài cảm nhận NGẮN GỌN, TỰ DO, TỐI ĐA 10 CÂU.
- Cảm nhận nên tự nhiên, có chiều sâu, như đang trò chuyện với người nghe nhạc.
- Kết hợp cả phân tích kỹ thuật (tempo, key, năng lượng) lẫn cảm xúc cá nhân.
- Không liệt kê số liệu khô khan, hãy diễn đạt thông qua ngôn ngữ cảm xúc.

%s`, langInstruction)
}

func buildUserPrompt(features *analyzer.AudioFeatures, metadata *audio.TrackMetadata) string {
	var sb strings.Builder

	// Metadata section
	sb.WriteString("## Thông tin bài hát\n")
	sb.WriteString(fmt.Sprintf("- Title: %s\n", metadata.Title))
	sb.WriteString(fmt.Sprintf("- Artist: %s\n", metadata.Artist))
	sb.WriteString(fmt.Sprintf("- Album: %s\n", metadata.Album))
	sb.WriteString(fmt.Sprintf("- Duration: %s\n", metadata.Duration.String()))
	sb.WriteString(fmt.Sprintf("- Sample Rate: %d Hz\n", metadata.SampleRate))

	// DSP Features section
	sb.WriteString("\n## Phân tích kỹ thuật\n")
	sb.WriteString(fmt.Sprintf("- BPM (Tempo): %.1f\n", features.BPM))
	sb.WriteString(fmt.Sprintf("- Key: %s\n", features.Key))
	sb.WriteString(fmt.Sprintf("- Spectral Centroid: %.1f Hz (%s)\n", features.SpectralCentroidMean, describeBrightness(features.SpectralCentroidMean)))
	sb.WriteString(fmt.Sprintf("- Spectral Bandwidth: %.1f Hz\n", features.SpectralBandwidthMean))
	sb.WriteString(fmt.Sprintf("- RMS Energy: %.4f (%s)\n", features.RMSEnergyMean, describeEnergy(features.RMSEnergyMean)))
	sb.WriteString(fmt.Sprintf("- Zero Crossing Rate: %.4f (%s)\n", features.ZeroCrossingRateMean, describeZCR(features.ZeroCrossingRateMean)))
	sb.WriteString(fmt.Sprintf("- Onset Strength: %.2f\n", features.OnsetStrengthMean))

	// Energy profile
	sb.WriteString(fmt.Sprintf("\n## Energy Profile (chia %d đoạn theo thời gian)\n", len(features.EnergyProfile)))
	sb.WriteString(describeEnergyProfile(features.EnergyProfile))

	// Mood
	if len(features.MoodKeywords) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Mood Keywords: %s\n", strings.Join(features.MoodKeywords, ", ")))
	}

	sb.WriteString("\nHãy viết cảm nhận của bạn về bài hát này.\n")
	return sb.String()
}

func describeBrightness(centroid float64) string {
	if centroid > 3000 {
		return "âm thanh sáng, nhiều treble"
	} else if centroid > 2000 {
		return "âm thanh cân bằng"
	} else if centroid > 1000 {
		return "âm thanh ấm áp"
	}
	return "âm thanh trầm, đầy đặn"
}

func describeEnergy(rms float64) string {
	if rms > 0.2 {
		return "rất mạnh mẽ"
	} else if rms > 0.1 {
		return "mạnh mẽ"
	} else if rms > 0.05 {
		return "vừa phải"
	}
	return "nhẹ nhàng"
}

func describeZCR(zcr float64) string {
	if zcr > 0.15 {
		return "nhiều nhịp điệu, percussive"
	} else if zcr > 0.08 {
		return "có nhịp điệu"
	}
	return "mượt mà, tonal"
}

func describeEnergyProfile(profile []float64) string {
	if len(profile) == 0 {
		return "Không có dữ liệu.\n"
	}

	var sb strings.Builder
	maxE := profile[0]
	minE := profile[0]
	maxIdx := 0
	for i, e := range profile {
		if e > maxE {
			maxE = e
			maxIdx = i
		}
		if e < minE {
			minE = e
		}
	}

	// Describe the arc
	startE := profile[0]
	endE := profile[len(profile)-1]
	midE := profile[len(profile)/2]

	sb.WriteString(fmt.Sprintf("- Năng lượng mở đầu: %.4f\n", startE))
	sb.WriteString(fmt.Sprintf("- Năng lượng giữa bài: %.4f\n", midE))
	sb.WriteString(fmt.Sprintf("- Năng lượng kết thúc: %.4f\n", endE))
	sb.WriteString(fmt.Sprintf("- Đỉnh năng lượng ở đoạn %d/%d (%.4f)\n", maxIdx+1, len(profile), maxE))

	if maxE-minE > 0.05 {
		sb.WriteString("- Dynamic range rộng: có cao trào rõ rệt\n")
	} else {
		sb.WriteString("- Dynamic range hẹp: năng lượng đều đặn\n")
	}

	return sb.String()
}
