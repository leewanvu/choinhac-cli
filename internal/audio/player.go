package audio

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"

	"github.com/dhowden/tag"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

// TrackMetadata holds info about the current track
type TrackMetadata struct {
	Title      string
	Artist     string
	Album      string
	SampleRate int
	Duration   time.Duration
}

// State represents the playback state
type State int

const (
	StateStopped State = iota
	StatePlaying
	StatePaused
)

// Player manages audio playback, stream formatting, and volume.
type Player struct {
	ctrl     *beep.Ctrl
	volume   *effects.Volume
	streamer beep.StreamSeekCloser
	format   beep.Format
	state       State
	Metadata    TrackMetadata
	Playlist    []string
	PlaylistIdx int

	// Channels for UI updates
	done chan bool
}

var (
	baseSampleRate = beep.SampleRate(44100)
	speakerInit    bool
)

// InitSpeaker initializes the global audio speaker. Must be called once.
func InitSpeaker() error {
	if speakerInit {
		return nil
	}
	err := speaker.Init(baseSampleRate, baseSampleRate.N(time.Second/10))
	if err != nil {
		return fmt.Errorf("failed to init speaker: %w", err)
	}
	speakerInit = true
	return nil
}

// NewPlayer creates a new player instance
func NewPlayer() *Player {
	return &Player{
		state: StateStopped,
		done:  make(chan bool, 1),
	}
}

// LoadPlaylist sets the playlist and plays the first valid track
func (p *Player) LoadPlaylist(paths []string, startIndex int) error {
	if len(paths) == 0 {
		return fmt.Errorf("empty playlist")
	}
	p.Playlist = paths
	if startIndex < 0 || startIndex >= len(paths) {
		startIndex = 0
	}
	p.PlaylistIdx = startIndex
	return p.LoadAndPlay(p.Playlist[p.PlaylistIdx])
}

// Next plays the next track in the playlist
func (p *Player) Next() error {
	if len(p.Playlist) <= 1 {
		return nil
	}
	p.PlaylistIdx = (p.PlaylistIdx + 1) % len(p.Playlist)
	return p.LoadAndPlay(p.Playlist[p.PlaylistIdx])
}

// Prev plays the previous track in the playlist
func (p *Player) Prev() error {
	if len(p.Playlist) <= 1 {
		return nil
	}
	p.PlaylistIdx--
	if p.PlaylistIdx < 0 {
		p.PlaylistIdx = len(p.Playlist) - 1
	}
	return p.LoadAndPlay(p.Playlist[p.PlaylistIdx])
}

// Random plays a random track in the playlist
func (p *Player) Random() error {
	if len(p.Playlist) <= 1 {
		return nil
	}
	p.PlaylistIdx = rand.IntN(len(p.Playlist))
	return p.LoadAndPlay(p.Playlist[p.PlaylistIdx])
}

// LoadAndPlay loads an audio file (WAV or FLAC), extracts metadata, and starts playback
func (p *Player) LoadAndPlay(filename string) error {
	p.Stop() // Stop any current playback

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", filename, err)
	}

	ext := filepath.Ext(filename)
	switch ext {
	case ".wav":
		p.streamer, p.format, err = wav.Decode(f)
	case ".flac":
		p.streamer, p.format, err = flac.Decode(f)
	default:
		f.Close()
		return fmt.Errorf("unsupported file format: %s", ext)
	}
	if err != nil {
		f.Close()
		return fmt.Errorf("could not decode file: %w", err)
	}

	// Extract Metadata
	p.extractMetadata(f, filename) // This might read tags so we reposition decoder

	// Loop first, then resample. beep.Loop requires a StreamSeeker.
	var looped beep.Streamer = beep.Loop(1, p.streamer)

	// Resample if necessary
	var finalStream beep.Streamer = looped
	if p.format.SampleRate != baseSampleRate {
		finalStream = beep.Resample(4, p.format.SampleRate, baseSampleRate, looped)
	}

	p.ctrl = &beep.Ctrl{Streamer: finalStream, Paused: false}
	p.volume = &effects.Volume{
		Streamer: p.ctrl,
		Base:     2,
		Volume:   0, // Normal volume (0 means 1.0x multiplier)
		Silent:   false,
	}

	// Play sequence that notifies when track is done
	// We wrap in beep.Seq to append a callback
	speaker.Play(beep.Seq(p.volume, beep.Callback(func() {
		p.state = StateStopped
		// non-blocking send
		select {
		case p.done <- true:
		default:
		}
	})))

	p.state = StatePlaying
	return nil
}

func (p *Player) extractMetadata(f *os.File, filename string) {
	// Need a new file handle because tag.ReadFrom might consume bytes differently
	// or we can just seek back to 0. But for safety, we open a new descriptor.
	tf, err := os.Open(filename)
	if err == nil {
		defer tf.Close()
		m, err := tag.ReadFrom(tf)
		if err == nil {
			p.Metadata = TrackMetadata{
				Title:      m.Title(),
				Artist:     m.Artist(),
				Album:      m.Album(),
				SampleRate: int(p.format.SampleRate),
				Duration:   p.format.SampleRate.D(p.streamer.Len()),
			}
			if p.Metadata.Title == "" {
				p.Metadata.Title = filepath.Base(filename)
			}
			return
		}
	}

	// fallback if no tags
	p.Metadata = TrackMetadata{
		Title:      filepath.Base(filename),
		Artist:     "Unknown Artist",
		Album:      "Unknown Album",
		SampleRate: int(p.format.SampleRate),
		Duration:   p.format.SampleRate.D(p.streamer.Len()),
	}
}

// TogglePause flips the pause state
func (p *Player) TogglePause() {
	if p.state == StateStopped || p.ctrl == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = !p.ctrl.Paused
	speaker.Unlock()

	if p.ctrl.Paused {
		p.state = StatePaused
	} else {
		p.state = StatePlaying
	}
}

// Stop stops playback entirely
func (p *Player) Stop() {
	if p.state == StateStopped {
		return
	}
	speaker.Clear()

	speaker.Lock()
	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}
	speaker.Unlock()

	p.state = StateStopped
}

// GetPosition returns current playback position safely
func (p *Player) GetPosition() time.Duration {
	if p.streamer == nil {
		return 0
	}
	speaker.Lock()
	pos := p.streamer.Position()
	speaker.Unlock()
	return p.format.SampleRate.D(pos)
}

// VolumeUp increases the volume smoothly
func (p *Player) VolumeUp() {
	if p.volume == nil {
		return
	}
	speaker.Lock()
	if p.volume.Volume < 2.0 {
		p.volume.Volume += 0.2
	}
	speaker.Unlock()
}

// VolumeDown decreases the volume smoothly
func (p *Player) VolumeDown() {
	if p.volume == nil {
		return
	}
	speaker.Lock()
	if p.volume.Volume > -5.0 {
		p.volume.Volume -= 0.2
	}
	speaker.Unlock()
}

// GetVolume returns the current volume multiplier (-5 to 2 log scale approx)
func (p *Player) GetVolume() float64 {
	if p.volume == nil {
		return 0
	}
	speaker.Lock()
	vol := p.volume.Volume
	speaker.Unlock()
	return vol
}

// GetState returns current player state (playing, paused, stopped)
func (p *Player) GetState() State {
	return p.state
}

// Done returns a channel that receives when the track finishes
func (p *Player) Done() <-chan bool {
	return p.done
}
