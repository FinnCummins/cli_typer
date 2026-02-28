package main

// The menu screen. Rows depend on the selected game mode:
//
// Classic mode (3 rows):
//   Row 0: game mode  — classic / falling
//   Row 1: content    — words / quotes
//   Row 2: duration   — 15s / 30s / 60s
//
// Falling mode (3 rows):
//   Row 0: game mode  — classic / falling
//   Row 1: content    — words / quotes
//   Row 2: cycle      — off / on

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

	maxRow := 2 // both modes have 3 rows now

	switch keyMsg.String() {
	case "up", "k":
		if m.menuRow > 0 {
			m.menuRow--
			return m, playSound(soundClick)
		}
	case "down", "j":
		if m.menuRow < maxRow {
			m.menuRow++
			return m, playSound(soundClick)
		}
	case "left", "h":
		handleMenuLeft(&m)
		return m, playSound(soundClick)
	case "right", "l":
		handleMenuRight(&m)
		return m, playSound(soundClick)
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

func handleMenuLeft(m *model) {
	switch m.menuRow {
	case 0: // game mode
		if m.gameMode == gameModeClassic {
			m.gameMode = gameModeFalling
		} else {
			m.gameMode = gameModeClassic
		}
	case 1: // content mode
		if m.contentMode == modeWords {
			m.contentMode = modeQuotes
		} else {
			m.contentMode = modeWords
		}
	case 2: // duration (classic) or cycle (falling)
		if m.gameMode == gameModeClassic {
			m.duration = cycleDuration(m.duration, -1)
		} else {
			m.dayCycle = !m.dayCycle
		}
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
	case 1:
		if m.contentMode == modeWords {
			m.contentMode = modeQuotes
		} else {
			m.contentMode = modeWords
		}
	case 2:
		if m.gameMode == gameModeClassic {
			m.duration = cycleDuration(m.duration, 1)
		} else {
			m.dayCycle = !m.dayCycle
		}
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

	// Row 2: depends on game mode
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
	} else {
		cycleLabel := styleStatLabel.Render("cycle     ")
		var offText, onText string
		if m.dayCycle {
			offText = styleUntyped.Render("  off  ")
			onText = styleHighlight.Render("[ on ]")
		} else {
			offText = styleHighlight.Render("[ off ]")
			onText = styleUntyped.Render("  on  ")
		}
		cycleRow := cycleLabel + offText + "  " + onText
		rows = append(rows, cycleRow)
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
