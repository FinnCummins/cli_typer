package main

import (
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type gameState int

const (
	stateMenu    gameState = iota
	stateTyping
	stateResults
	stateFalling
)

type contentMode int

const (
	modeWords  contentMode = iota
	modeQuotes
)

type gameMode int

const (
	gameModeClassic gameMode = iota
	gameModeFalling
)

// model holds ALL application state.
type model struct {
	// Global
	state  gameState
	width  int
	height int

	// Menu
	menuRow     int
	gameMode    gameMode
	contentMode contentMode
	duration    time.Duration

	// Classic typing test
	words     []string
	input     [][]rune
	wordIndex int
	charIndex int

	// Classic timer
	timer        timer.Model
	timerStarted bool
	startTime    time.Time

	// Results (shared between modes)
	finalWPM      float64
	finalAccuracy float64
	correctChars  int
	totalChars    int
	correctWords  int
	totalWords    int

	// Falling words mode
	fallingWords     []fallingWord // active words on screen
	fallingInput     []rune        // what the user is currently typing
	fallingTarget    int           // index of targeted word, or -1
	fallingLives     int           // starts at 3, game over at 0
	fallingScore     int           // words destroyed
	fallingSpeed     float64       // rows per tick (increases over time)
	fallingSpawnCD   int           // ticks until next word spawns
	fallingTicks     int           // total ticks elapsed
	fallingStartTime time.Time     // for "time survived"
	fallingGameOver  bool
	fallingCharsTyped int          // total chars in destroyed words (for WPM)
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

// initTypingState sets up a fresh classic typing session.
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
	m.timer = timer.NewWithInterval(m.duration, time.Second)
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
	case stateFalling:
		return updateFalling(m, msg)
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.state {
	case stateFalling:
		// Falling mode manages its own full-screen layout
		return viewFalling(m)
	default:
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
}
