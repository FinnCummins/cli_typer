package main

// Falling words game mode.
//
// Words spawn at the top of the screen and fall downward. Type a word to
// destroy it before it hits the bottom. You get 3 lives — lose one each
// time a word reaches the bottom. Difficulty ramps over time: words fall
// faster and spawn more frequently.
//
// Word targeting: when you type a character with no active target, we
// find the lowest falling word whose first letter matches. That word
// becomes your target (highlighted). Finish typing it to destroy it.
//
// The animation loop uses tea.Tick — a bubbletea function that fires
// a message after a delay. We return a new tick command each time to
// keep the loop going (tea.Tick is one-shot, not repeating).

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// fallingWord represents one word on the play field.
type fallingWord struct {
	word   string
	x      int     // column position (left edge)
	y      float64 // row position (fractional for smooth movement)
	typed  int     // how many leading characters have been matched
	active bool    // true if this is the user's current target
}

// fallingTickMsg is our custom message for the animation loop.
// In bubbletea, you define your own message types as simple types.
// The Update function matches on them with a type switch.
type fallingTickMsg time.Time

// fallingTickCmd returns a command that fires a fallingTickMsg after 150ms.
// tea.Tick(duration, func) schedules a one-shot timer. The func wraps the
// time.Time into our custom message type.
func fallingTickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return fallingTickMsg(t)
	})
}

// initFallingState resets all falling-mode state for a new game.
func initFallingState(m model) model {
	m.state = stateFalling
	m.fallingWords = nil
	m.fallingInput = nil
	m.fallingTarget = -1
	m.fallingLives = 3
	m.fallingScore = 0
	m.fallingSpeed = 0.3
	m.fallingSpawnCD = 0 // spawn immediately
	m.fallingTicks = 0
	m.fallingGameOver = false
	m.fallingStartTime = time.Now()
	m.fallingCharsTyped = 0
	return m
}

func updateFalling(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fallingTickMsg:
		if m.fallingGameOver {
			return m, nil // stop ticking
		}
		m = fallingTick(m)
		if m.fallingGameOver {
			return m, nil // just became game over
		}
		return m, fallingTickCmd() // schedule next tick

	case tea.KeyMsg:
		if m.fallingGameOver {
			return handleGameOverKey(m, msg)
		}
		return handleFallingKey(m, msg)
	}

	return m, nil
}

// fallingTick runs every 150ms: move words, check bottom, spawn, scale difficulty.
func fallingTick(m model) model {
	m.fallingTicks++

	// Move all words down
	for i := range m.fallingWords {
		m.fallingWords[i].y += m.fallingSpeed
	}

	// Check for words hitting the bottom
	playHeight := m.height - 5
	if playHeight < 5 {
		playHeight = 5
	}

	// Collect surviving words (can't remove from slice while iterating)
	var survived []fallingWord
	targetWord := ""
	if m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
		targetWord = m.fallingWords[m.fallingTarget].word
	}

	for _, fw := range m.fallingWords {
		if int(fw.y) >= playHeight {
			// This word hit the bottom
			m.fallingLives--
			if fw.active {
				// Lost our target
				m.fallingInput = nil
				targetWord = ""
			}
			if m.fallingLives <= 0 {
				m.fallingLives = 0
				m.fallingGameOver = true
				m = calculateFallingResults(m)
				return m
			}
		} else {
			survived = append(survived, fw)
		}
	}
	m.fallingWords = survived

	// Re-find target index after slice rebuild
	m.fallingTarget = -1
	if targetWord != "" {
		for i, fw := range m.fallingWords {
			if fw.active && fw.word == targetWord {
				m.fallingTarget = i
				break
			}
		}
		if m.fallingTarget == -1 {
			// Target was lost (hit bottom)
			m.fallingInput = nil
		}
	}

	// Spawn new words
	m.fallingSpawnCD--
	if m.fallingSpawnCD <= 0 {
		m = spawnFallingWord(m)
		m.fallingSpawnCD = fallingSpawnInterval(m.fallingTicks)
	}

	// Scale difficulty
	m.fallingSpeed = fallingSpeedForTick(m.fallingTicks)

	return m
}

func spawnFallingWord(m model) model {
	var word string
	if m.contentMode == modeQuotes {
		allWords := getQuoteWords(50)
		word = allWords[rand.Intn(len(allWords))]
	} else {
		word = commonWords[rand.Intn(len(commonWords))]
	}

	maxX := m.width - len(word) - 2
	if maxX < 1 {
		maxX = 1
	}
	x := rand.Intn(maxX) + 1

	m.fallingWords = append(m.fallingWords, fallingWord{
		word: word,
		x:    x,
		y:    0,
	})
	return m
}

func handleFallingKey(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.state = stateMenu
		return m, nil

	case tea.KeyTab:
		m = initFallingState(m)
		return m, fallingTickCmd()

	case tea.KeyBackspace:
		if len(m.fallingInput) > 0 {
			m.fallingInput = m.fallingInput[:len(m.fallingInput)-1]
			if m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
				m.fallingWords[m.fallingTarget].typed = len(m.fallingInput)
			}
			// If input is now empty, release the target
			if len(m.fallingInput) == 0 && m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
				m.fallingWords[m.fallingTarget].active = false
				m.fallingWords[m.fallingTarget].typed = 0
				m.fallingTarget = -1
			}
		}
		return m, nil

	case tea.KeySpace:
		// Ignore space in falling mode (words don't have spaces)
		return m, nil

	case tea.KeyRunes:
		char := msg.Runes[0]
		m.fallingInput = append(m.fallingInput, char)

		if m.fallingTarget == -1 {
			// No target — find one by first character
			m.fallingTarget = findTarget(m, char)
			if m.fallingTarget >= 0 {
				m.fallingWords[m.fallingTarget].active = true
				m.fallingWords[m.fallingTarget].typed = 1
			}
		} else if m.fallingTarget < len(m.fallingWords) {
			m.fallingWords[m.fallingTarget].typed = len(m.fallingInput)
		}

		// Check if the word is complete
		if m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
			fw := m.fallingWords[m.fallingTarget]
			if string(m.fallingInput) == fw.word {
				// Word destroyed!
				m.fallingScore++
				m.fallingCharsTyped += len(fw.word)
				m.fallingWords = append(m.fallingWords[:m.fallingTarget], m.fallingWords[m.fallingTarget+1:]...)
				m.fallingTarget = -1
				m.fallingInput = nil
			}
		}

		return m, nil
	}

	return m, nil
}

// findTarget finds the lowest (highest Y) word whose first character matches.
func findTarget(m model, firstChar rune) int {
	bestIdx := -1
	bestY := -1.0

	for i, fw := range m.fallingWords {
		if fw.active {
			continue
		}
		runes := []rune(fw.word)
		if len(runes) > 0 && runes[0] == firstChar && fw.y > bestY {
			bestY = fw.y
			bestIdx = i
		}
	}
	return bestIdx
}

func handleGameOverKey(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyTab, tea.KeyEnter:
		m = initFallingState(m)
		return m, fallingTickCmd()
	case tea.KeyEsc:
		m.state = stateMenu
		return m, nil
	}
	return m, nil
}

func calculateFallingResults(m model) model {
	elapsed := time.Since(m.fallingStartTime).Seconds()
	if elapsed < 1 {
		elapsed = 1
	}
	minutes := elapsed / 60.0
	m.finalWPM = (float64(m.fallingCharsTyped) / 5.0) / minutes
	m.correctWords = m.fallingScore
	return m
}

// --- Difficulty scaling ---

func fallingSpeedForTick(ticks int) float64 {
	base := 0.3
	increments := float64(ticks / 67) // every ~10 seconds
	speed := base + increments*0.05
	if speed > 1.5 {
		speed = 1.5
	}
	return speed
}

func fallingSpawnInterval(ticks int) int {
	base := 20
	reduction := ticks / 67
	interval := base - reduction*2
	if interval < 7 {
		interval = 7
	}
	return interval
}

// --- Rendering ---

func viewFalling(m model) string {
	playHeight := m.height - 5
	if playHeight < 5 {
		playHeight = 5
	}
	playWidth := m.width
	if playWidth < 20 {
		playWidth = 20
	}

	// Build a 2D grid of styled characters
	grid := make([][]string, playHeight)
	for row := range grid {
		grid[row] = make([]string, playWidth)
		for col := range grid[row] {
			grid[row][col] = " "
		}
	}

	// Place each falling word on the grid
	for _, fw := range m.fallingWords {
		row := int(fw.y)
		if row < 0 || row >= playHeight {
			continue
		}
		for j, ch := range []rune(fw.word) {
			col := fw.x + j
			if col < 0 || col >= playWidth {
				continue
			}
			if fw.active && j < fw.typed {
				// Already typed portion — correct color
				grid[row][col] = styleCorrect.Render(string(ch))
			} else if fw.active {
				// Remaining portion of targeted word — highlighted
				grid[row][col] = styleCursor.Render(string(ch))
			} else {
				// Normal untargeted word
				grid[row][col] = styleUntyped.Render(string(ch))
			}
		}
	}

	// Render grid to lines
	var lines []string
	for _, row := range grid {
		lines = append(lines, strings.Join(row, ""))
	}
	playField := strings.Join(lines, "\n")

	// Status bar: lives, score, time
	hearts := styleLife.Render(strings.Repeat("♥ ", m.fallingLives))
	if m.fallingLives == 0 {
		hearts = styleHint.Render("♥ ♥ ♥")
	}
	scoreText := styleStatLabel.Render("score ") + styleStatValue.Render(fmt.Sprintf("%d", m.fallingScore))
	elapsed := time.Since(m.fallingStartTime).Seconds()
	timeText := styleStatLabel.Render("time ") + styleStatValue.Render(fmt.Sprintf("%.0fs", elapsed))
	statusBar := hearts + "  " + scoreText + "  " + timeText

	// Separator line
	separator := styleHint.Render(strings.Repeat("─", playWidth))

	// Input line
	inputStr := string(m.fallingInput)
	inputDisplay := styleHighlight.Render("> ") + styleCorrect.Render(inputStr) + styleCursor.Render("_")

	hint := styleHint.Render("tab restart  esc menu")

	if m.fallingGameOver {
		return viewFallingGameOver(m)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		statusBar,
		playField,
		separator,
		inputDisplay,
		hint,
	)
}

func viewFallingGameOver(m model) string {
	// Center the game over screen
	gameOver := styleLife.Render("GAME OVER")

	scoreNum := styleBigWPM.Render(fmt.Sprintf("%d", m.fallingScore))
	scoreLabel := styleHint.Render(" words destroyed")

	elapsed := time.Since(m.fallingStartTime).Seconds()
	timeStat := styleStatLabel.Render("survived     ") + styleStatValue.Render(fmt.Sprintf("%.0fs", elapsed))
	wpmStat := styleStatLabel.Render("wpm          ") + styleStatValue.Render(fmt.Sprintf("%.0f", m.finalWPM))

	hint := styleHint.Render("tab/enter restart  esc menu")

	content := lipgloss.JoinVertical(lipgloss.Left,
		gameOver,
		"",
		scoreNum+scoreLabel,
		"",
		timeStat,
		wpmStat,
		"",
		hint,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
