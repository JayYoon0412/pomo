package audio

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

// Player manages looping ambient audio playback for a session.
type Player struct{}

// NewPlayer returns a new Player.
func NewPlayer() *Player {
	return &Player{}
}

// PlayLoop starts playing the sound at path (an embedded FS path) in an
// infinite loop. Playback runs in beep's background goroutine and does not
// block the caller.
func (p *Player) PlayLoop(path string) error {
	data, err := soundData(path)
	if err != nil {
		return err
	}

	streamer, format, err := wav.Decode(io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return fmt.Errorf("audio: decode %q: %w", path, err)
	}

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return fmt.Errorf("audio: init speaker: %w", err)
	}

	speaker.Play(beep.Loop(-1, streamer))
	return nil
}

// Stop halts all active playback immediately.
func (p *Player) Stop() {
	speaker.Clear()
}
