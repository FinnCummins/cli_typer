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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// gameState represents which screen we're on.
// In Go, this is an "iota enum" — iota auto-increments from 0.
type gameState int

const (
	stateMenu    gameState = iota // 0
	stateTyping                   // 1
	stateResults                  // 2
)

// contentMode is what kind of text the user types.
type contentMode int

const (
	modeWords  contentMode = iota // 0
	modeQuotes                    // 1
)

// model holds ALL application state. Every field lives here.
// In Go, structs are value types (like a C struct, not a reference).
// When bubbletea calls Update, it passes a COPY of the model.
type model struct {
	// Global
	state  gameState
	width  int // terminal width  (updated via WindowSizeMsg)
	height int // terminal height (updated via WindowSizeMsg)

	// Menu
	menuRow     int         // which row is selected (0=mode, 1=duration)
	contentMode contentMode // words or quotes
	duration    time.Duration
}

// The three duration options the user can cycle through.
var durations = []time.Duration{
	15 * time.Second,
	30 * time.Second,
	60 * time.Second,
}

// initialModel returns the starting state of the app.
func initialModel() model {
	return model{
		state:    stateMenu,
		duration: 30 * time.Second,
	}
}

// Init is called once when the program starts.
func (m model) Init() tea.Cmd {
	return nil
}

// Update is the central message handler. It dispatches to screen-specific
// update functions based on the current state.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resize globally (applies to all screens).
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	// Handle ctrl+c globally.
	if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// Dispatch to the current screen's update handler.
	switch m.state {
	case stateMenu:
		return updateMenu(m, msg)
	case stateTyping:
		// TODO: will be added in step 4
	case stateResults:
		// TODO: will be added in step 7
	}

	return m, nil
}

// View dispatches to the current screen's view function.
func (m model) View() string {
	if m.width == 0 {
		return "" // waiting for initial WindowSizeMsg
	}

	var content string
	switch m.state {
	case stateMenu:
		content = viewMenu(m)
	case stateTyping:
		content = "typing screen (TODO)"
	case stateResults:
		content = "results screen (TODO)"
	}

	// Center everything on screen. lipgloss.Place is like CSS flexbox centering.
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
