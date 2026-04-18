package ui

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	artWidth  = 20 // terminal columns
	artHeight = 10 // terminal rows (each row = 2 image rows via half-block)
)

var artPlaceholder = []string{
	"┌──────────────────┐",
	"│                  │",
	"│   ♪  ♫  ♪  ♫   │",
	"│                  │",
	"│   ♫  ♬  ♫  ♬   │",
	"│                  │",
	"│   ♪  ♫  ♪  ♫   │",
	"│                  │",
	"│   ♫  ♬  ♫  ♬   │",
	"└──────────────────┘",
}

// renderArt converts cover art bytes to a half-block terminal string.
// Falls back to placeholder on any error.
func renderArt(data []byte) string {
	if len(data) == 0 {
		return renderPlaceholder()
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return renderPlaceholder()
	}
	return renderImage(img)
}

func renderPlaceholder() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	var sb strings.Builder
	for _, line := range artPlaceholder {
		sb.WriteString(style.Render(line))
		sb.WriteString("\n")
	}
	return sb.String()
}

func renderImage(img image.Image) string {
	bounds := img.Bounds()
	srcW := bounds.Max.X - bounds.Min.X
	srcH := bounds.Max.Y - bounds.Min.Y

	scaleX := float64(srcW) / float64(artWidth)
	scaleY := float64(srcH) / float64(artHeight*2) // *2: half-block uses 2 rows per terminal row

	var sb strings.Builder
	for row := 0; row < artHeight; row++ {
		for col := 0; col < artWidth; col++ {
			ux := bounds.Min.X + int(float64(col)*scaleX)
			uy := bounds.Min.Y + int(float64(row*2)*scaleY)
			ly := bounds.Min.Y + int(float64(row*2+1)*scaleY)

			ur, ug, ub, _ := img.At(ux, uy).RGBA()
			lr, lg, lb, _ := img.At(ux, ly).RGBA()

			upper := lipgloss.Color(rgbHex(uint8(ur>>8), uint8(ug>>8), uint8(ub>>8)))
			lower := lipgloss.Color(rgbHex(uint8(lr>>8), uint8(lg>>8), uint8(lb>>8)))

			// '▀' = upper half block; fg = upper color, bg = lower color
			cell := lipgloss.NewStyle().Foreground(upper).Background(lower).Render("▀")
			sb.WriteString(cell)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func rgbHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
