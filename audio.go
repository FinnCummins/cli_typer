package main

// Cross-platform audio playback using gopxl/beep.
//
// Go's embed package (//go:embed directive) bakes files into the binary at
// compile time. This means the sounds ship inside the executable — no external
// files needed at runtime.
//
// beep/speaker is initialized once at startup. To play a sound, we decode
// the embedded OGG data into a buffer (in-memory), then play a copy of it
// each time. Buffering avoids re-decoding on every play.

import (
	"bytes"
	"embed"
	"fmt"
	"math/rand"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"

	tea "github.com/charmbracelet/bubbletea"
)

//go:embed sounds/destroy1.ogg sounds/destroy2.ogg sounds/destroy3.ogg sounds/destroy4.ogg sounds/hit.ogg sounds/gameover.ogg sounds/click.ogg
var soundFiles embed.FS

// Pre-decoded sound buffers.
var (
	soundDestroy  [4]*beep.Buffer // 4 variations, picked randomly
	soundHit      *beep.Buffer
	soundGameOver *beep.Buffer
	soundClick    *beep.Buffer
	audioReady    bool
)

func initAudio() {
	// Decode the first destroy sound to get the sample rate for speaker init
	firstData, err := soundFiles.ReadFile("sounds/destroy1.ogg")
	if err != nil {
		return
	}

	streamer, format, err := vorbis.Decode(nopCloser(firstData))
	if err != nil {
		return
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return
	}

	// Buffer the first destroy sound
	soundDestroy[0] = beep.NewBuffer(format)
	soundDestroy[0].Append(streamer)

	// Buffer the remaining 3 destroy sounds
	for i := 1; i < 4; i++ {
		data, err := soundFiles.ReadFile(fmt.Sprintf("sounds/destroy%d.ogg", i+1))
		if err == nil {
			if s, _, err := vorbis.Decode(nopCloser(data)); err == nil {
				soundDestroy[i] = beep.NewBuffer(format)
				soundDestroy[i].Append(s)
			}
		}
	}

	// Buffer the hit sound
	hitData, err := soundFiles.ReadFile("sounds/hit.ogg")
	if err == nil {
		if s, _, err := vorbis.Decode(nopCloser(hitData)); err == nil {
			soundHit = beep.NewBuffer(format)
			soundHit.Append(s)
		}
	}

	// Buffer the game over sound
	goData, err := soundFiles.ReadFile("sounds/gameover.ogg")
	if err == nil {
		if s, _, err := vorbis.Decode(nopCloser(goData)); err == nil {
			soundGameOver = beep.NewBuffer(format)
			soundGameOver.Append(s)
		}
	}

	// Buffer the click sound
	clickData, err := soundFiles.ReadFile("sounds/click.ogg")
	if err == nil {
		if s, _, err := vorbis.Decode(nopCloser(clickData)); err == nil {
			soundClick = beep.NewBuffer(format)
			soundClick.Append(s)
		}
	}

	audioReady = true
}

// playSound returns a tea.Cmd that plays a buffered sound.
func playSound(buf *beep.Buffer) tea.Cmd {
	if !audioReady || buf == nil {
		return nil
	}
	return func() tea.Msg {
		speaker.Play(buf.Streamer(0, buf.Len()))
		return nil
	}
}

// playRandomDestroy returns a tea.Cmd that plays one of the 4 destroy sounds at random.
func playRandomDestroy() tea.Cmd {
	if !audioReady {
		return nil
	}
	buf := soundDestroy[rand.Intn(4)]
	if buf == nil {
		return nil
	}
	return playSound(buf)
}

type readCloser struct {
	*bytes.Reader
}

func (readCloser) Close() error { return nil }

func nopCloser(data []byte) *readCloser {
	return &readCloser{bytes.NewReader(data)}
}
