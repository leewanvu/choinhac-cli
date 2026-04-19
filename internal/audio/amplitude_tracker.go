package audio

import (
	"math"
	"sync/atomic"

	"github.com/gopxl/beep"
)

// amplitudeTracker wraps a Streamer and captures peak amplitude per chunk.
type amplitudeTracker struct {
	beep.Streamer
	peak atomic.Uint64 // stores float64 bits via math.Float64bits
}

func newAmplitudeTracker(s beep.Streamer) *amplitudeTracker {
	return &amplitudeTracker{Streamer: s}
}

func (a *amplitudeTracker) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = a.Streamer.Stream(samples)
	peak := 0.0
	for _, s := range samples[:n] {
		if v := math.Abs(s[0]); v > peak {
			peak = v
		}
		if v := math.Abs(s[1]); v > peak {
			peak = v
		}
	}
	a.peak.Store(math.Float64bits(peak))
	return
}

func (a *amplitudeTracker) amplitude() float64 {
	return math.Float64frombits(a.peak.Load())
}
