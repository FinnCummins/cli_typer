package main

// All visual styling lives in this one file. If you want to tweak colors
// or spacing, this is the only place you need to look.
//
// lipgloss is like CSS for the terminal — you define styles (foreground color,
// bold, padding, etc.) and then call style.Render("text") to apply them.

import "github.com/charmbracelet/lipgloss"

// Monkeytype-inspired color palette
var (
	colorBg      = lipgloss.Color("#323437") // dark background
	colorDim     = lipgloss.Color("#646669") // untyped text
	colorText    = lipgloss.Color("#d1d0c5") // correctly typed text
	colorError   = lipgloss.Color("#ca4754") // incorrectly typed text
	colorAccent  = lipgloss.Color("#e2b714") // cursor, highlights, accents
	colorSuccess = lipgloss.Color("#98c379") // positive results
)

// Character-level styles (used in the typing view to color individual chars)
var (
	styleUntyped   = lipgloss.NewStyle().Foreground(colorDim)
	styleCorrect   = lipgloss.NewStyle().Foreground(colorText)
	styleIncorrect = lipgloss.NewStyle().Foreground(colorError)
	styleCursor    = lipgloss.NewStyle().Foreground(colorBg).Background(colorAccent)
)

// UI element styles
var (
	styleTitle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleTimer = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleHint = lipgloss.NewStyle().
			Foreground(colorDim)

	styleStatLabel = lipgloss.NewStyle().
			Foreground(colorDim)

	styleStatValue = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleHighlight = lipgloss.NewStyle().
			Foreground(colorAccent)

	// Results screen — large WPM display
	styleBigWPM = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	styleLiveWPM = lipgloss.NewStyle().
			Foreground(colorDim)

	// Falling words mode
	styleLife = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	styleShield = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4fc1ff")).
			Bold(true)

	styleShieldDamaged = lipgloss.NewStyle().
				Foreground(colorError)

	styleAlien = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7c6f9f"))

	styleAlienActive = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	styleLaser = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff6b6b")).
			Bold(true)

	styleExplosion = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffaa44")).
			Bold(true)
)
