package ui

import (
	"math/rand/v2"
	"strings"
)

const (
	vizBars        = 24
	vizDecay       = 0.75
	vizSensitivity = 8.0
)

// blocks ordered low→high amplitude
var blocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

type visualizer struct {
	bars [vizBars]float64
}

// update feeds a new amplitude sample into the visualizer state.
func (v *visualizer) update(amplitude float64) {
	scaled := amplitude * vizSensitivity
	if scaled > 1.0 {
		scaled = 1.0
	}

	for i := range v.bars {
		spread := (rand.Float64()*0.4 - 0.2) * scaled
		target := scaled + spread
		if target < 0 {
			target = 0
		}
		if target > 1.0 {
			target = 1.0
		}
		// Bars rise fast, fall slow via exponential decay
		decayed := v.bars[i] * vizDecay
		if target > decayed {
			v.bars[i] = target
		} else {
			v.bars[i] = decayed
		}
	}
}

// render returns a single-line string of block characters for the visualizer.
func (v *visualizer) render() string {
	var sb strings.Builder
	for _, h := range v.bars {
		idx := int(h * float64(len(blocks)-1))
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		sb.WriteRune(blocks[idx])
		sb.WriteRune(' ')
	}
	return vizStyle.Render(strings.TrimRight(sb.String(), " "))
}
