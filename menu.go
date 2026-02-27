package main

// The menu screen. Two rows:
//   Row 0: content mode  — toggle between "words" and "quotes" with left/right
//   Row 1: duration      — cycle between 15s, 30s, 60s with left/right
//
// Up/down (or j/k) to move between rows. Enter to start the test.

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func updateMenu(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	// We only care about keypresses on this screen.
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		if m.menuRow > 0 {
			m.menuRow--
		}
	case "down", "j":
		if m.menuRow < 1 {
			m.menuRow++
		}
	case "left", "h":
		if m.menuRow == 0 {
			// Toggle content mode
			if m.contentMode == modeWords {
				m.contentMode = modeQuotes
			} else {
				m.contentMode = modeWords
			}
		} else {
			// Cycle duration backward
			m.duration = cycleDuration(m.duration, -1)
		}
	case "right", "l":
		if m.menuRow == 0 {
			// Toggle content mode
			if m.contentMode == modeWords {
				m.contentMode = modeQuotes
			} else {
				m.contentMode = modeWords
			}
		} else {
			// Cycle duration forward
			m.duration = cycleDuration(m.duration, 1)
		}
	case "enter":
		// TODO: transition to typing state (step 4)
		return m, nil
	case "q":
		return m, tea.Quit
	}

	return m, nil
}

func viewMenu(m model) string {
	title := styleTitle.Render("fun_cli")

	// Mode row
	modeLabel := styleStatLabel.Render("mode      ")
	wordsText := "words"
	quotesText := "quotes"
	if m.contentMode == modeWords {
		wordsText = styleHighlight.Render("[ words ]")
		quotesText = styleUntyped.Render("  quotes ")
	} else {
		wordsText = styleUntyped.Render("  words  ")
		quotesText = styleHighlight.Render("[ quotes ]")
	}
	modeRow := modeLabel + wordsText + "  " + quotesText

	// Duration row
	durLabel := styleStatLabel.Render("duration  ")
	var durParts []string
	for _, d := range durations {
		text := fmt.Sprintf("%ds", int(d.Seconds()))
		if d == m.duration {
			durParts = append(durParts, styleHighlight.Render(fmt.Sprintf("[ %s ]", text)))
		} else {
			durParts = append(durParts, styleUntyped.Render(fmt.Sprintf("  %s  ", text)))
		}
	}
	durRow := durLabel
	for _, p := range durParts {
		durRow += p + " "
	}

	// Arrow indicator for selected row
	rows := []string{modeRow, durRow}
	var renderedRows []string
	for i, row := range rows {
		if i == m.menuRow {
			renderedRows = append(renderedRows, styleHighlight.Render("▸ ")+row)
		} else {
			renderedRows = append(renderedRows, "  "+row)
		}
	}

	hint := styleHint.Render("↑↓ navigate  ←→ change  enter start  q quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		renderedRows[0],
		renderedRows[1],
		"",
		hint,
	)
}

// cycleDuration moves forward or backward through the durations slice.
func cycleDuration(current time.Duration, direction int) time.Duration {
	for i, d := range durations {
		if d == current {
			next := i + direction
			if next < 0 {
				next = len(durations) - 1
			}
			if next >= len(durations) {
				next = 0
			}
			return durations[next]
		}
	}
	return current
}
