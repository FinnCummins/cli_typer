package main

// This file defines the central model and implements bubbletea's tea.Model interface.
//
// Bubbletea uses the Elm architecture:
//   1. Model  — a struct that holds ALL application state
//   2. Update — a function that receives messages (keypresses, timer ticks, etc.)
//               and returns an updated model
//   3. View   — a function that takes the model and returns a string to render
//
// The framework calls Update whenever something happens, then calls View to
// re-render. You never mutate state directly — you return a new model from Update.

import (
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// gameState represents which screen we're on.
type gameState int

const (
	stateMenu    gameState = iota
	stateTyping
	stateResults
)

// contentMode is what kind of text the user types.
type contentMode int

const (
	modeWords  contentMode = iota
	modeQuotes
)

// model holds ALL application state.
type model struct {
	// Global
	state  gameState
	width  int
	height int

	// Menu
	menuRow     int
	contentMode contentMode
	duration    time.Duration

	// Typing
	//
	// The key insight: words[] is the target text, input[][] is what the user typed.
	// They're parallel arrays — input[i] holds the runes typed for words[i].
	//
	// Example:
	//   words: ["the", "quick", "brown"]
	//   input: [['t','h','e'], ['q','i','c','k'], []]
	//                                              ^ wordIndex=2, charIndex=0
	words     []string  // target words to type
	input     [][]rune  // what the user has typed for each word
	wordIndex int       // which word the cursor is on
	charIndex int       // cursor position within current word's input

	// Timer
	// timer.Model is from the bubbles library — it handles tick scheduling
	// and sends timer.TickMsg every interval, plus timer.TimeoutMsg when done.
	// We create it in initTypingState but don't start it until the first keypress.
	timer        timer.Model
	timerStarted bool
	startTime    time.Time

	// Results (will be populated in step 7)
	finalWPM      float64
	finalAccuracy float64
	correctChars  int
	totalChars    int
	correctWords  int
	totalWords    int
}

var durations = []time.Duration{
	15 * time.Second,
	30 * time.Second,
	60 * time.Second,
}

func initialModel() model {
	return model{
		state:    stateMenu,
		duration: 30 * time.Second,
	}
}

// initTypingState sets up a fresh typing session based on current menu settings.
func initTypingState(m model) model {
	var words []string
	if m.contentMode == modeQuotes {
		words = getQuoteWords(200)
	} else {
		words = generateWords(200)
	}

	m.state = stateTyping
	m.words = words
	m.input = make([][]rune, len(words))
	m.wordIndex = 0
	m.charIndex = 0
	m.timerStarted = false
	m.timer = timer.NewWithInterval(m.duration, time.Second) // ticks every 1s
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	switch m.state {
	case stateMenu:
		return updateMenu(m, msg)
	case stateTyping:
		return updateTyping(m, msg)
	case stateResults:
		return updateResults(m, msg)
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	var content string
	switch m.state {
	case stateMenu:
		content = viewMenu(m)
	case stateTyping:
		content = viewTyping(m)
	case stateResults:
		content = viewResults(m)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
