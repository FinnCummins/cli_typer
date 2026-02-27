package main

// The menu screen. Rows depend on the selected game mode:
//
// Classic mode (3 rows):
//   Row 0: game mode  — classic / falling
//   Row 1: content    — words / quotes
//   Row 2: duration   — 15s / 30s / 60s
//
// Falling mode (2 rows):
//   Row 0: game mode  — classic / falling
//   Row 1: content    — words / quotes
//   (no duration row — falling mode ends when you lose all lives)

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func updateMenu(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	maxRow := 2
	if m.gameMode == gameModeFalling {
		maxRow = 1
	}

	switch keyMsg.String() {
	case "up", "k":
		if m.menuRow > 0 {
			m.menuRow--
		}
	case "down", "j":
		if m.menuRow < maxRow {
			m.menuRow++
		}
	case "left", "h":
		handleMenuLeft(&m)
	case "right", "l":
		handleMenuRight(&m)
	case "enter":
		if m.gameMode == gameModeFalling {
			m = initFallingState(m)
			return m, fallingTickCmd()
		}
		m = initTypingState(m)
		return m, nil
	case "q":
		return m, tea.Quit
	}

	return m, nil
}

// handleMenuLeft processes left arrow on the current menu row.
// We use a pointer here so mutations are visible to the caller.
// (An alternative Go pattern — sometimes simpler than returning the model.)
func handleMenuLeft(m *model) {
	switch m.menuRow {
	case 0: // game mode
		if m.gameMode == gameModeClassic {
			m.gameMode = gameModeFalling
		} else {
			m.gameMode = gameModeClassic
		}
		// Clamp menuRow if we switched to a mode with fewer rows
		maxRow := 2
		if m.gameMode == gameModeFalling {
			maxRow = 1
		}
		if m.menuRow > maxRow {
			m.menuRow = maxRow
		}
	case 1: // content mode
		if m.contentMode == modeWords {
			m.contentMode = modeQuotes
		} else {
			m.contentMode = modeWords
		}
	case 2: // duration (classic only)
		m.duration = cycleDuration(m.duration, -1)
	}
}

func handleMenuRight(m *model) {
	switch m.menuRow {
	case 0:
		if m.gameMode == gameModeClassic {
			m.gameMode = gameModeFalling
		} else {
			m.gameMode = gameModeClassic
		}
		maxRow := 2
		if m.gameMode == gameModeFalling {
			maxRow = 1
		}
		if m.menuRow > maxRow {
			m.menuRow = maxRow
		}
	case 1:
		if m.contentMode == modeWords {
			m.contentMode = modeQuotes
		} else {
			m.contentMode = modeWords
		}
	case 2:
		m.duration = cycleDuration(m.duration, 1)
	}
}

func viewMenu(m model) string {
	title := styleTitle.Render("cli_typer")

	// Row 0: Game mode
	gameModeLabel := styleStatLabel.Render("game      ")
	var classicText, fallingText string
	if m.gameMode == gameModeClassic {
		classicText = styleHighlight.Render("[ classic ]")
		fallingText = styleUntyped.Render("  falling ")
	} else {
		classicText = styleUntyped.Render("  classic  ")
		fallingText = styleHighlight.Render("[ falling ]")
	}
	gameModeRow := gameModeLabel + classicText + " " + fallingText

	// Row 1: Content mode
	modeLabel := styleStatLabel.Render("words     ")
	var wordsText, quotesText string
	if m.contentMode == modeWords {
		wordsText = styleHighlight.Render("[ words ]")
		quotesText = styleUntyped.Render("  quotes ")
	} else {
		wordsText = styleUntyped.Render("  words  ")
		quotesText = styleHighlight.Render("[ quotes ]")
	}
	modeRow := modeLabel + wordsText + "  " + quotesText

	// Build the list of rows
	rows := []string{gameModeRow, modeRow}

	// Row 2: Duration (classic mode only)
	if m.gameMode == gameModeClassic {
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
		rows = append(rows, durRow)
	}

	// Add arrow indicator for selected row
	var renderedRows []string
	for i, row := range rows {
		if i == m.menuRow {
			renderedRows = append(renderedRows, styleHighlight.Render("▸ ")+row)
		} else {
			renderedRows = append(renderedRows, "  "+row)
		}
	}

	hint := styleHint.Render("↑↓ navigate  ←→ change  enter start  q quit")

	parts := []string{title, ""}
	parts = append(parts, renderedRows...)
	parts = append(parts, "", hint)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

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
