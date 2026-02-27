package main

// The typing test screen. This is where the core gameplay happens.
//
// Input tracking:
//   - Each regular keypress appends a rune to input[wordIndex]
//   - Space advances to the next word
//   - Backspace removes the last rune from the current word
//   - You can't backspace into a previous word (matches monkeytype)
//
// Timer:
//   - Created in initTypingState but NOT started
//   - Started on the very first keypress (via timer.Init())
//   - Ticks every second, sending timer.TickMsg which triggers a re-render
//   - When it hits zero, sends timer.TimeoutMsg → transition to results

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxWordOverflow = 5

func updateTyping(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case timer.TickMsg:
		// The timer sends TickMsg every second. We forward it to timer.Update()
		// which decrements the remaining time and returns a cmd to schedule
		// the next tick. This is the "command" pattern in Elm architecture —
		// side effects (like scheduling a future tick) are returned as commands,
		// never executed directly.
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		// Time's up! Calculate results and switch screens.
		m = calculateResults(m)
		m.state = stateResults
		return m, nil

	case tea.KeyMsg:
		// Start the timer on the very first keypress.
		// timer.Init() returns a Cmd that kicks off the first tick.
		if !m.timerStarted {
			m.timerStarted = true
			m.startTime = time.Now()
			cmd := m.timer.Init()
			// Process this keypress AND start the timer simultaneously
			m, _ = processKeypress(m, msg)
			return m, cmd
		}

		return processKeypress(m, msg)
	}

	return m, nil
}

// processKeypress handles a single keypress during the typing test.
// Separated from updateTyping so we can call it alongside timer.Init()
// on the first keypress without duplicating logic.
func processKeypress(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.Type {

	case tea.KeyEsc:
		m.state = stateMenu
		return m, nil

	case tea.KeyTab:
		m = initTypingState(m)
		return m, nil

	case tea.KeyBackspace:
		if m.charIndex > 0 {
			m.charIndex--
			m.input[m.wordIndex] = m.input[m.wordIndex][:m.charIndex]
		}
		return m, nil

	case tea.KeySpace:
		// Only advance if the user has typed something for this word.
		// Prevents accidental double-space from skipping words.
		if len(m.input[m.wordIndex]) > 0 && m.wordIndex < len(m.words)-1 {
			m.wordIndex++
			m.charIndex = 0
		}
		return m, nil

	case tea.KeyRunes:
		char := msg.Runes[0]
		targetLen := len([]rune(m.words[m.wordIndex]))
		if m.charIndex < targetLen+maxWordOverflow {
			m.input[m.wordIndex] = append(m.input[m.wordIndex], char)
			m.charIndex++
		}
		return m, nil
	}

	return m, nil
}

func viewTyping(m model) string {
	const containerWidth = 70

	lines := wrapWords(m.words, containerWidth)

	// Find which line the current word is on
	currentLine := 0
	for i, line := range lines {
		for _, wIdx := range line {
			if wIdx == m.wordIndex {
				currentLine = i
			}
		}
	}

	// Show 3 lines: one above current, current, one below
	startLine := currentLine - 1
	if startLine < 0 {
		startLine = 0
	}
	endLine := startLine + 3
	if endLine > len(lines) {
		endLine = len(lines)
	}

	var renderedLines []string
	for _, line := range lines[startLine:endLine] {
		var lineStr strings.Builder
		for j, wIdx := range line {
			if j > 0 {
				lineStr.WriteString(styleUntyped.Render(" "))
			}
			lineStr.WriteString(renderWord(m, wIdx))
		}
		renderedLines = append(renderedLines, lineStr.String())
	}

	textBlock := strings.Join(renderedLines, "\n")

	// Timer display
	var timerText string
	if !m.timerStarted {
		timerText = styleTimer.Render(fmt.Sprintf("%d", int(m.duration.Seconds())))
	} else {
		remaining := m.timer.Timeout.Seconds()
		timerText = styleTimer.Render(fmt.Sprintf("%d", int(remaining)))
	}

	hint := styleHint.Render("tab restart  esc menu")

	content := lipgloss.JoinVertical(lipgloss.Left,
		timerText,
		"",
		textBlock,
		"",
		hint,
	)

	return content
}

// renderWord renders a single word with character-by-character styling.
func renderWord(m model, wordIdx int) string {
	target := []rune(m.words[wordIdx])
	typed := m.input[wordIdx]
	var result strings.Builder

	for i, targetChar := range target {
		if wordIdx < m.wordIndex {
			if i < len(typed) && typed[i] == targetChar {
				result.WriteString(styleCorrect.Render(string(targetChar)))
			} else {
				result.WriteString(styleIncorrect.Render(string(targetChar)))
			}
		} else if wordIdx == m.wordIndex {
			if i < len(typed) {
				if typed[i] == targetChar {
					result.WriteString(styleCorrect.Render(string(targetChar)))
				} else {
					result.WriteString(styleIncorrect.Render(string(targetChar)))
				}
			} else if i == len(typed) {
				result.WriteString(styleCursor.Render(string(targetChar)))
			} else {
				result.WriteString(styleUntyped.Render(string(targetChar)))
			}
		} else {
			result.WriteString(styleUntyped.Render(string(targetChar)))
		}
	}

	// Overflow characters (typed more than the word length)
	if wordIdx <= m.wordIndex && len(typed) > len(target) {
		for i := len(target); i < len(typed); i++ {
			result.WriteString(styleIncorrect.Render(string(typed[i])))
		}
	}

	return result.String()
}

// wrapWords groups word indices into lines that fit within maxWidth.
func wrapWords(words []string, maxWidth int) [][]int {
	var lines [][]int
	var currentLine []int
	lineWidth := 0

	for i, word := range words {
		wordWidth := len([]rune(word))
		spaceNeeded := wordWidth
		if len(currentLine) > 0 {
			spaceNeeded++
		}

		if lineWidth+spaceNeeded > maxWidth && len(currentLine) > 0 {
			lines = append(lines, currentLine)
			currentLine = []int{i}
			lineWidth = wordWidth
		} else {
			currentLine = append(currentLine, i)
			lineWidth += spaceNeeded
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}
	return lines
}
