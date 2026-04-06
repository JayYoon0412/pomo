package audio

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

var speakerOnce sync.Once

type readSeekCloser struct{ *bytes.Reader }
func (readSeekCloser) Close() error { return nil }

type Player struct{}
func NewPlayer() *Player {
	return &Player{}
}

func (p *Player) PlayLoop(path string) error {
	data, err := soundData(path)
	if err != nil {
		return err
	}

	streamer, format, err := wav.Decode(readSeekCloser{bytes.NewReader(data)})
	if err != nil {
		return fmt.Errorf("audio: decode %q: %w", path, err)
	}

	var initErr error
	speakerOnce.Do(func() {
		initErr = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	})
	if initErr != nil {
		return fmt.Errorf("audio: init speaker: %w", initErr)
	}

	speaker.Play(beep.Loop(-1, streamer))
	return nil
}

func (p *Player) Stop() {
	speaker.Clear()
}
