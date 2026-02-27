package main

// The results screen shown after the timer expires.
//
// WPM calculation uses the standard convention: 1 "word" = 5 characters.
// This is the same formula monkeytype, typeracer, and every major typing
// test uses. We calculate "net WPM" which only counts correct characters.
//
//   Net WPM = (correct characters / 5) / minutes elapsed

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// calculateResults computes WPM and accuracy from the typing session.
func calculateResults(m model) model {
	elapsed := time.Since(m.startTime).Seconds()
	if elapsed < 1 {
		elapsed = 1
	}

	correctChars := 0
	totalChars := 0
	correctWords := 0

	for i := 0; i < len(m.words); i++ {
		if i > m.wordIndex {
			break // don't count words the user never reached
		}

		typed := m.input[i]
		target := []rune(m.words[i])

		wordCorrect := true
		for j := 0; j < len(target); j++ {
			totalChars++
			if j < len(typed) && typed[j] == target[j] {
				correctChars++
			} else {
				wordCorrect = false
			}
		}

		// Overflow characters count as errors
		if len(typed) > len(target) {
			totalChars += len(typed) - len(target)
			wordCorrect = false
		}

		// Spaces between words (implicitly correct if user pressed space)
		if i < m.wordIndex {
			totalChars++
			correctChars++
		}

		if wordCorrect && len(typed) == len(target) {
			correctWords++
		}
	}

	minutes := elapsed / 60.0
	netWPM := (float64(correctChars) / 5.0) / minutes
	if netWPM < 0 {
		netWPM = 0
	}

	accuracy := 0.0
	if totalChars > 0 {
		accuracy = float64(correctChars) / float64(totalChars) * 100
	}

	m.finalWPM = netWPM
	m.finalAccuracy = accuracy
	m.correctChars = correctChars
	m.totalChars = totalChars
	m.correctWords = correctWords
	m.totalWords = m.wordIndex + 1
	return m
}

func updateResults(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.Type {
	case tea.KeyTab, tea.KeyEnter:
		// Restart with same settings
		m = initTypingState(m)
		return m, nil
	case tea.KeyEsc:
		m.state = stateMenu
		return m, nil
	}

	return m, nil
}

func viewResults(m model) string {
	// Big WPM number
	wpm := styleStatValue.Copy().Bold(true).Render(fmt.Sprintf("%.0f wpm", m.finalWPM))

	// Stats
	acc := styleStatLabel.Render("accuracy     ") + styleStatValue.Render(fmt.Sprintf("%.1f%%", m.finalAccuracy))
	chars := styleStatLabel.Render("characters   ") + styleStatValue.Render(fmt.Sprintf("%d/%d", m.correctChars, m.totalChars))
	words := styleStatLabel.Render("words        ") + styleStatValue.Render(fmt.Sprintf("%d/%d", m.correctWords, m.totalWords))

	hint := styleHint.Render("tab/enter restart  esc menu")

	return lipgloss.JoinVertical(lipgloss.Left,
		wpm,
		"",
		acc,
		chars,
		words,
		"",
		hint,
	)
}
