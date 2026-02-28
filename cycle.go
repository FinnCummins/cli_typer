package main

// Day/night cycle for falling words mode.
//
// The sun and moon follow semicircular arcs across the play field.
// Colors interpolate through 4 phases: dawn → day → sunset → night.
// The terminal background shifts along with foreground colors.
//
// Full cycle = 800 ticks (~2 minutes at 150ms/tick).
// Ticks 0-399:   Day   — sun arcs left to right
// Ticks 400-799: Night — moon arcs left to right

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

const (
	fullCycleTicks = 800
	halfCycleTicks = 400
)

type rgb struct {
	r, g, b float64
}

func (c rgb) toHex() string {
	r := int(math.Round(clamp(c.r, 0, 255)))
	g := int(math.Round(clamp(c.g, 0, 255)))
	b := int(math.Round(clamp(c.b, 0, 255)))
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func lerpRGB(a, b rgb, t float64) rgb {
	t = clamp(t, 0, 1)
	return rgb{
		r: a.r + (b.r-a.r)*t,
		g: a.g + (b.g-a.g)*t,
		b: a.b + (b.b-a.b)*t,
	}
}

// Color keyframes — foreground elements
var (
	dawnDim    = rgb{138, 110, 66}
	dawnText   = rgb{212, 184, 150}
	dawnAlien  = rgb{156, 118, 68}
	dawnShield = rgb{196, 154, 86}
	dawnAccent = rgb{226, 168, 60}
	dawnHint   = rgb{138, 110, 66}

	dayDim    = rgb{140, 140, 155}
	dayText   = rgb{20, 20, 30}
	dayAlien  = rgb{50, 30, 110}
	dayShield = rgb{20, 60, 140}
	dayAccent = rgb{130, 80, 0}
	dayHint   = rgb{140, 140, 155}

	sunsetDim    = rgb{139, 64, 73}
	sunsetText   = rgb{212, 150, 122}
	sunsetAlien  = rgb{160, 72, 88}
	sunsetShield = rgb{196, 90, 62}
	sunsetAccent = rgb{220, 130, 50}
	sunsetHint   = rgb{139, 64, 73}

	nightDim    = rgb{70, 80, 110}
	nightText   = rgb{180, 190, 220}
	nightAlien  = rgb{90, 100, 160}
	nightShield = rgb{100, 130, 190}
	nightAccent = rgb{140, 170, 220}
	nightHint   = rgb{70, 80, 110}
)

// Background color keyframes
var (
	dawnBg   = rgb{180, 140, 80}  // warm golden dawn
	dayBg    = rgb{255, 255, 255} // pure white
	sunsetBg = rgb{180, 100, 50}  // warm orange sunset
	nightBg  = rgb{0, 0, 0}      // pure black
)

type cyclePalette struct {
	dim    lipgloss.Color
	text   lipgloss.Color
	alien  lipgloss.Color
	shield lipgloss.Color
	accent lipgloss.Color
	hint   lipgloss.Color
	bg     lipgloss.Color // background color
}

func cycleColors(tick int) cyclePalette {
	pos := tick % fullCycleTicks
	isDay := pos < halfCycleTicks

	var progress float64
	if isDay {
		progress = float64(pos) / float64(halfCycleTicks)
	} else {
		progress = float64(pos-halfCycleTicks) / float64(halfCycleTicks)
	}

	var dim, text, alien, shield, accent, hint, bg rgb

	// Transition zones are 8% of the arc — rapid shift between phases
	const edge = 0.08

	if isDay {
		if progress < edge {
			t := progress / edge
			dim = lerpRGB(dawnDim, dayDim, t)
			text = lerpRGB(dawnText, dayText, t)
			alien = lerpRGB(dawnAlien, dayAlien, t)
			shield = lerpRGB(dawnShield, dayShield, t)
			accent = lerpRGB(dawnAccent, dayAccent, t)
			hint = lerpRGB(dawnHint, dayHint, t)
			bg = lerpRGB(dawnBg, dayBg, t)
		} else if progress < 1.0-edge {
			dim = dayDim
			text = dayText
			alien = dayAlien
			shield = dayShield
			accent = dayAccent
			hint = dayHint
			bg = dayBg
		} else {
			t := (progress - (1.0 - edge)) / edge
			dim = lerpRGB(dayDim, sunsetDim, t)
			text = lerpRGB(dayText, sunsetText, t)
			alien = lerpRGB(dayAlien, sunsetAlien, t)
			shield = lerpRGB(dayShield, sunsetShield, t)
			accent = lerpRGB(dayAccent, sunsetAccent, t)
			hint = lerpRGB(dayHint, sunsetHint, t)
			bg = lerpRGB(dayBg, sunsetBg, t)
		}
	} else {
		if progress < edge {
			t := progress / edge
			dim = lerpRGB(sunsetDim, nightDim, t)
			text = lerpRGB(sunsetText, nightText, t)
			alien = lerpRGB(sunsetAlien, nightAlien, t)
			shield = lerpRGB(sunsetShield, nightShield, t)
			accent = lerpRGB(sunsetAccent, nightAccent, t)
			hint = lerpRGB(sunsetHint, nightHint, t)
			bg = lerpRGB(sunsetBg, nightBg, t)
		} else if progress < 1.0-edge {
			dim = nightDim
			text = nightText
			alien = nightAlien
			shield = nightShield
			accent = nightAccent
			hint = nightHint
			bg = nightBg
		} else {
			t := (progress - (1.0 - edge)) / edge
			dim = lerpRGB(nightDim, dawnDim, t)
			text = lerpRGB(nightText, dawnText, t)
			alien = lerpRGB(nightAlien, dawnAlien, t)
			shield = lerpRGB(nightShield, dawnShield, t)
			accent = lerpRGB(nightAccent, dawnAccent, t)
			hint = lerpRGB(nightHint, dawnHint, t)
			bg = lerpRGB(nightBg, dawnBg, t)
		}
	}

	return cyclePalette{
		dim:    lipgloss.Color(dim.toHex()),
		text:   lipgloss.Color(text.toHex()),
		alien:  lipgloss.Color(alien.toHex()),
		shield: lipgloss.Color(shield.toHex()),
		accent: lipgloss.Color(accent.toHex()),
		hint:   lipgloss.Color(hint.toHex()),
		bg:     lipgloss.Color(bg.toHex()),
	}
}

// --- Celestial Bodies (multi-character sprites) ---

// celestialSprite is a character placed at an offset from the body's center.
type celestialSprite struct {
	dx, dy int
	ch     string
	bright bool // true = primary color, false = glow/ray color
}

// Sun sprite — 5x3 with rays
//
//	 \|/
//	--O--
//	 /|\
var sunSprites = []celestialSprite{
	// Rays
	{-1, -1, "\\", false}, {0, -1, "|", false}, {1, -1, "/", false},
	{-2, 0, "-", false}, {-1, 0, "-", false},
	{1, 0, "-", false}, {2, 0, "-", false},
	{-1, 1, "/", false}, {0, 1, "|", false}, {1, 1, "\\", false},
	// Core
	{0, 0, "O", true},
}

// Moon sprite — 3x3 crescent
//
//	 ▄█
//	 ██
//	 ▀█
var moonSprites = []celestialSprite{
	{0, -1, "▄", false}, {1, -1, "█", true},
	{0, 0, " ", false}, {1, 0, "█", true},
	{0, 1, "▀", false}, {1, 1, "█", true},
}

// celestialBody holds position info for rendering.
type celestialBody struct {
	x, y     int
	isDay    bool
	coreFg   string // hex color for the bright parts
	glowFg   string // hex color for rays/glow
}

func celestialPosition(progress float64, playWidth, playHeight int) (int, int) {
	angle := math.Pi * (1.0 - progress)

	centerX := float64(playWidth) / 2.0
	groundY := float64(playHeight) - 2 // start just above shield

	// Use an elliptical arc that fits within the play field.
	// Horizontal radius spans most of the width.
	// Vertical radius is capped so the peak stays on screen (row 2+).
	radiusX := float64(playWidth) / 2.5
	radiusY := float64(playHeight) - 4 // leaves room at top
	if radiusY < 3 {
		radiusY = 3
	}

	x := centerX + radiusX*math.Cos(angle)
	y := groundY - radiusY*math.Sin(angle)

	// Clamp to visible area
	if y < 1 {
		y = 1
	}
	if y > float64(playHeight-2) {
		y = float64(playHeight - 2)
	}

	return int(math.Round(x)), int(math.Round(y))
}

func getCelestialBody(tick int, playWidth, playHeight int) celestialBody {
	pos := tick % fullCycleTicks
	isDay := pos < halfCycleTicks

	var progress float64
	if isDay {
		progress = float64(pos) / float64(halfCycleTicks)
	} else {
		progress = float64(pos-halfCycleTicks) / float64(halfCycleTicks)
	}

	x, y := celestialPosition(progress, playWidth, playHeight)

	if isDay {
		if progress < 0.2 || progress > 0.8 {
			return celestialBody{x, y, true, "#e8903a", "#a06020"}
		}
		return celestialBody{x, y, true, "#f5d442", "#c4a030"}
	}

	if progress < 0.2 || progress > 0.8 {
		return celestialBody{x, y, false, "#6677aa", "#334466"}
	}
	return celestialBody{x, y, false, "#ccddef", "#7799bb"}
}

// renderCelestialOnGrid places the sun or moon sprite on the grid.
func renderCelestialOnGrid(grid [][]string, body celestialBody, playWidth, playHeight int) {
	sprites := moonSprites
	if body.isDay {
		sprites = sunSprites
	}

	coreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(body.coreFg)).Bold(true)
	glowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(body.glowFg))

	for _, sp := range sprites {
		px := body.x + sp.dx
		py := body.y + sp.dy
		if py >= 0 && py < playHeight && px >= 0 && px < playWidth {
			if sp.bright {
				grid[py][px] = coreStyle.Render(sp.ch)
			} else {
				grid[py][px] = glowStyle.Render(sp.ch)
			}
		}
	}
}
