package main

// Falling words game mode with multi-row ASCII art aliens:
//
//   Small:       Medium:       Large:          XLarge:
//     .-.          ___           ._.              \_/
//    (o o)       (o o)         (o o)            (o o)
//    |the|       |quick|     /|afternoon|\    <|accomplishment|>
//    /| |\        /| |\       / | \              / | | \
//                 /   \      /  |  \              /   \
//
// - Turret on the shield slides to track the targeted word
// - Laser beam + explosion on word destroy
// - Overlap-aware spawning prevents aliens from stacking

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	edgePadding     = 3
	turretSpeed     = 3
	laserDuration   = 3
	explodeDuration = 4
)

type fallingWord struct {
	word   string
	x      int     // left edge of the alien art
	y      float64 // row of the WORD LINE (always row index 2 of the alien)
	typed  int
	active bool
}

type explosion struct {
	x     int
	y     int
	ticks int
}

type laserBeam struct {
	x     int
	fromY int
	toY   int
	ticks int
}

type fallingTickMsg time.Time

func fallingTickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return fallingTickMsg(t)
	})
}

// --- Multi-row ASCII Art Alien Builder ---
//
// Each alien is built to exactly fit its word — no padding.
// The body row is always " |word| " and the head/legs are centered
// to match that width. Three visual styles based on word length:
//
// Short (1-3):    Medium (4-6):     Long (7+):
//    .-.             ___               ._.
//   (o o)          (o o)             (o o)
//   |no|           |like|          |between|
//   /| |\          /| |\            / | \
//                  /   \           /  |  \

type builtAlien struct {
	lines   []string
	wordRow int
	wordCol int
	wordLen int
	width   int
}

func buildAlienArt(word string) builtAlien {
	n := len(word)
	bodyRow := " |" + word + "| "
	totalWidth := len(bodyRow)

	// center pads a string to totalWidth
	center := func(s string) string {
		pad := totalWidth - len(s)
		if pad <= 0 {
			return s
		}
		lp := pad / 2
		rp := pad - lp
		return strings.Repeat(" ", lp) + s + strings.Repeat(" ", rp)
	}

	var lines []string
	if n <= 3 {
		lines = []string{
			center(".-."),
			center("(o o)"),
			bodyRow,
			center(`/| |\`),
		}
	} else if n <= 6 {
		lines = []string{
			center("___"),
			center("(o o)"),
			bodyRow,
			center(`/| |\`),
			center(`/   \`),
		}
	} else {
		lines = []string{
			center("._."),
			center("(o o)"),
			bodyRow,
			center(`/ | \`),
			center(`/  |  \`),
		}
	}

	return builtAlien{
		lines:   lines,
		wordRow: 2,
		wordCol: 2, // " |" = 2 chars before word starts
		wordLen: n,
		width:   totalWidth,
	}
}

// --- Game state management ---

func initFallingState(m model) model {
	m.state = stateFalling
	m.fallingWords = nil
	m.fallingInput = nil
	m.fallingTarget = -1
	m.fallingLives = 3
	m.fallingScore = 0
	m.fallingSpeed = 0.3
	m.fallingSpawnCD = 0
	m.fallingTicks = 0
	m.fallingGameOver = false
	m.fallingStartTime = time.Now()
	m.fallingCharsTyped = 0
	m.turretX = m.width / 2
	m.explosions = nil
	m.laser = nil
	return m
}

func updateFalling(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fallingTickMsg:
		if m.fallingGameOver {
			return m, nil
		}
		livesBefore := m.fallingLives
		m = fallingTick(m)
		var cmds []tea.Cmd
		if m.fallingLives < livesBefore {
			cmds = append(cmds, playSound(soundHit))
		}
		if m.fallingGameOver {
			cmds = append(cmds, playSound(soundGameOver))
			return m, tea.Batch(cmds...)
		}
		cmds = append(cmds, fallingTickCmd())
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if m.fallingGameOver {
			return handleGameOverKey(m, msg)
		}
		return handleFallingKey(m, msg)
	}

	return m, nil
}

func fallingTick(m model) model {
	m.fallingTicks++

	for i := range m.fallingWords {
		m.fallingWords[i].y += m.fallingSpeed
	}

	// Tick down explosions
	var activeExplosions []explosion
	for _, e := range m.explosions {
		e.ticks--
		if e.ticks > 0 {
			activeExplosions = append(activeExplosions, e)
		}
	}
	m.explosions = activeExplosions

	if m.laser != nil {
		m.laser.ticks--
		if m.laser.ticks <= 0 {
			m.laser = nil
		}
	}

	// Check for words hitting the shield
	playHeight := m.height - 6
	if playHeight < 5 {
		playHeight = 5
	}

	var survived []fallingWord
	targetWord := ""
	if m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
		targetWord = m.fallingWords[m.fallingTarget].word
	}

	for _, fw := range m.fallingWords {
		if int(fw.y) >= playHeight {
			m.fallingLives--
			if fw.active {
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

	m.fallingTarget = -1
	if targetWord != "" {
		for i, fw := range m.fallingWords {
			if fw.active && fw.word == targetWord {
				m.fallingTarget = i
				break
			}
		}
		if m.fallingTarget == -1 {
			m.fallingInput = nil
		}
	}

	m.fallingSpawnCD--
	if m.fallingSpawnCD <= 0 {
		m = spawnFallingWord(m)
		m.fallingSpawnCD = fallingSpawnInterval(m.fallingTicks)
	}

	m.fallingSpeed = fallingSpeedForTick(m.fallingTicks)

	return m
}

// wordCenter returns the screen column of the word's center for turret targeting.
func wordCenter(fw fallingWord) int {
	art := buildAlienArt(fw.word)
	return fw.x + art.wordCol + art.wordLen/2
}

func overlapsExisting(m model, art builtAlien, x int) bool {
	newLeft := x
	newRight := x + art.width

	for _, fw := range m.fallingWords {
		if fw.y > 5 {
			continue
		}
		existArt := buildAlienArt(fw.word)
		existLeft := fw.x
		existRight := fw.x + existArt.width

		if newLeft < existRight+1 && newRight > existLeft-1 {
			return true
		}
	}
	return false
}

func spawnFallingWord(m model) model {
	var word string
	if m.contentMode == modeQuotes {
		allWords := getQuoteWords(50)
		word = allWords[rand.Intn(len(allWords))]
	} else {
		word = commonWords[rand.Intn(len(commonWords))]
	}

	art := buildAlienArt(word)
	minX := edgePadding
	maxX := m.width - art.width - edgePadding
	if maxX <= minX {
		maxX = minX + 1
	}

	var x int
	placed := false
	for attempt := 0; attempt < 10; attempt++ {
		x = rand.Intn(maxX-minX) + minX
		if !overlapsExisting(m, art, x) {
			placed = true
			break
		}
	}

	if !placed {
		m.fallingSpawnCD = 3
		return m
	}

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
			if len(m.fallingInput) == 0 && m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
				m.fallingWords[m.fallingTarget].active = false
				m.fallingWords[m.fallingTarget].typed = 0
				m.fallingTarget = -1
			}
		}
		return m, nil

	case tea.KeySpace:
		return m, nil

	case tea.KeyRunes:
		char := msg.Runes[0]
		m.fallingInput = append(m.fallingInput, char)

		if m.fallingTarget == -1 {
			m.fallingTarget = findTarget(m, char)
			if m.fallingTarget >= 0 {
				m.fallingWords[m.fallingTarget].active = true
				m.fallingWords[m.fallingTarget].typed = 1
				m.turretStartX = m.turretX
			}
		} else if m.fallingTarget < len(m.fallingWords) {
			m.fallingWords[m.fallingTarget].typed = len(m.fallingInput)
		}

		// Move turret proportionally toward target center
		if m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
			fw := m.fallingWords[m.fallingTarget]
			targetX := wordCenter(fw)
			wordLen := len([]rune(fw.word))
			if wordLen > 0 {
				progress := float64(len(m.fallingInput)) / float64(wordLen)
				m.turretX = m.turretStartX + int(progress*float64(targetX-m.turretStartX))
			}
		}

		if m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
			fw := m.fallingWords[m.fallingTarget]
			if string(m.fallingInput) == fw.word {
				centerX := wordCenter(fw)
				wordRowY := int(fw.y)

				playHeight := m.height - 6
				if playHeight < 5 {
					playHeight = 5
				}

				m.laser = &laserBeam{
					x:     centerX,
					fromY: playHeight,
					toY:   wordRowY - 2, // laser reaches the top of the alien
					ticks: laserDuration,
				}
				m.explosions = append(m.explosions, explosion{
					x:     centerX,
					y:     wordRowY,
					ticks: explodeDuration,
				})

				m.turretX = centerX
				m.fallingScore++
				m.fallingCharsTyped += len(fw.word)
				m.fallingWords = append(m.fallingWords[:m.fallingTarget], m.fallingWords[m.fallingTarget+1:]...)
				m.fallingTarget = -1
				m.fallingInput = nil
				return m, playRandomDestroy()
			}
		}

		return m, nil
	}

	return m, nil
}

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
	m.correctWords = m.fallingScore
	return m
}

// --- Difficulty scaling ---

func fallingSpeedForTick(ticks int) float64 {
	base := 0.3
	increments := float64(ticks / 67)
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

func renderShieldWithStyle(width int, lives int, turretX int, sShield, sShieldDmg, sHint lipgloss.Style) string {
	if width < 4 {
		width = 4
	}

	var shield []rune
	switch lives {
	case 3:
		shield = []rune(strings.Repeat("█", width))
	case 2:
		shield = []rune(strings.Repeat("█", width))
		for _, pos := range []int{width / 4, width / 2, width * 3 / 4} {
			if pos < len(shield) {
				shield[pos] = '░'
			}
		}
	case 1:
		shield = make([]rune, width)
		for i := range shield {
			if i%3 == 0 {
				shield[i] = '░'
			} else if i%5 == 0 {
				shield[i] = ' '
			} else {
				shield[i] = '▒'
			}
		}
	default:
		shield = make([]rune, width)
		for i := range shield {
			if i%2 == 0 {
				shield[i] = '░'
			} else {
				shield[i] = ' '
			}
		}
	}

	turretPos := turretX
	if turretPos < 1 {
		turretPos = 1
	}
	if turretPos >= width-1 {
		turretPos = width - 2
	}
	if turretPos-1 >= 0 && turretPos-1 < len(shield) {
		shield[turretPos-1] = '/'
	}
	if turretPos < len(shield) {
		shield[turretPos] = '▲'
	}
	if turretPos+1 < len(shield) {
		shield[turretPos+1] = '\\'
	}

	var result strings.Builder
	for i, ch := range shield {
		s := string(ch)
		if i >= turretPos-1 && i <= turretPos+1 {
			result.WriteString(sShield.Render(s))
		} else if lives >= 2 {
			result.WriteString(sShield.Render(s))
		} else if lives == 1 {
			result.WriteString(sShieldDmg.Render(s))
		} else {
			result.WriteString(sHint.Render(s))
		}
	}
	return result.String()
}

func viewFalling(m model) string {
	playHeight := m.height - 6
	if playHeight < 5 {
		playHeight = 5
	}
	playWidth := m.width
	if playWidth < 20 {
		playWidth = 20
	}

	// Compute styles — either dynamic (cycle) or static (default)
	sUntyped := styleUntyped
	sAlien := styleAlien
	sAlienActive := styleAlienActive
	sShield := styleShield
	sShieldDmg := styleShieldDamaged
	sHint := styleHint
	sStatLabel := styleStatLabel
	sStatValue := styleStatValue
	sHighlight := styleHighlight
	hasCycle := m.dayCycle
	var cycleBg lipgloss.Color

	if hasCycle {
		pal := cycleColors(m.fallingTicks)
		cycleBg = pal.bg
		sUntyped = lipgloss.NewStyle().Foreground(pal.dim)
		sAlien = lipgloss.NewStyle().Foreground(pal.alien)
		sAlienActive = lipgloss.NewStyle().Foreground(pal.accent).Bold(true)
		sShield = lipgloss.NewStyle().Foreground(pal.shield).Bold(true)
		sShieldDmg = lipgloss.NewStyle().Foreground(pal.dim)
		sHint = lipgloss.NewStyle().Foreground(pal.hint)
		sStatLabel = lipgloss.NewStyle().Foreground(pal.hint)
		sStatValue = lipgloss.NewStyle().Foreground(pal.accent).Bold(true)
		sHighlight = lipgloss.NewStyle().Foreground(pal.accent)
	}

	// Build 2D grid
	grid := make([][]string, playHeight)
	for row := range grid {
		grid[row] = make([]string, playWidth)
		for col := range grid[row] {
			grid[row][col] = " "
		}
	}

	// Draw celestial body (sun or moon)
	if m.dayCycle {
		body := getCelestialBody(m.fallingTicks, playWidth, playHeight)
		renderCelestialOnGrid(grid, body, playWidth, playHeight)
	}

	// Draw laser beam
	if m.laser != nil {
		col := m.laser.x
		if col >= 0 && col < playWidth {
			for row := m.laser.toY; row < m.laser.fromY && row < playHeight; row++ {
				if row >= 0 {
					grid[row][col] = styleLaser.Render("│")
				}
			}
		}
	}

	// Draw explosions
	for _, e := range m.explosions {
		phase := explodeDuration - e.ticks
		particles := explosionParticles(phase)
		for _, p := range particles {
			px := e.x + p.dx
			py := e.y + p.dy
			if py >= 0 && py < playHeight && px >= 0 && px < playWidth {
				grid[py][px] = styleExplosion.Render(p.ch)
			}
		}
	}

	// Place multi-row alien sprites
	for _, fw := range m.fallingWords {
		art := buildAlienArt(fw.word)
		wordRowY := int(fw.y) // the word row on the grid

		aStyle := sAlien
		if fw.active {
			aStyle = sAlienActive
		}

		for rowIdx, line := range art.lines {
			gridRow := wordRowY - art.wordRow + rowIdx
			if gridRow < 0 || gridRow >= playHeight {
				continue
			}

			for colIdx, ch := range []rune(line) {
				if ch == ' ' {
					continue // don't overwrite grid background with spaces
				}
				gridCol := fw.x + colIdx
				if gridCol < 0 || gridCol >= playWidth {
					continue
				}

				// Is this character part of the word text?
				if rowIdx == art.wordRow && colIdx >= art.wordCol && colIdx < art.wordCol+art.wordLen {
					charIdx := colIdx - art.wordCol
					if fw.active && charIdx < fw.typed {
						grid[gridRow][gridCol] = styleCorrect.Render(string(ch))
					} else if fw.active {
						grid[gridRow][gridCol] = styleCursor.Render(string(ch))
					} else {
						grid[gridRow][gridCol] = sUntyped.Render(string(ch))
					}
				} else {
					// Alien decoration character
					grid[gridRow][gridCol] = aStyle.Render(string(ch))
				}
			}
		}
	}

	// Render grid
	var lines []string
	for _, row := range grid {
		lines = append(lines, strings.Join(row, ""))
	}
	playField := strings.Join(lines, "\n")

	// Shield with dynamic colors
	shield := renderShieldWithStyle(playWidth, m.fallingLives, m.turretX, sShield, sShieldDmg, sHint)

	hearts := styleLife.Render(strings.Repeat("♥ ", m.fallingLives))
	if m.fallingLives == 0 {
		hearts = sHint.Render("♥ ♥ ♥")
	}
	scoreText := sStatLabel.Render("score ") + sStatValue.Render(fmt.Sprintf("%d", m.fallingScore))
	elapsed := time.Since(m.fallingStartTime).Seconds()
	timeText := sStatLabel.Render("time ") + sStatValue.Render(fmt.Sprintf("%.0fs", elapsed))
	statusBar := hearts + "  " + scoreText + "  " + timeText

	inputStr := string(m.fallingInput)
	inputDisplay := sHighlight.Render("> ") + styleCorrect.Render(inputStr) + styleCursor.Render("_")

	hint := sHint.Render("tab restart  esc menu")

	if m.fallingGameOver {
		return viewFallingGameOver(m)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		statusBar,
		playField,
		shield,
		inputDisplay,
		hint,
	)

	if hasCycle {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Left, lipgloss.Top,
			content,
			lipgloss.WithWhitespaceBackground(cycleBg),
		)
	}

	return content
}

type particle struct {
	dx, dy int
	ch     string
}

func explosionParticles(phase int) []particle {
	switch phase {
	case 0:
		return []particle{
			{0, 0, "✦"},
		}
	case 1:
		return []particle{
			{0, 0, "◇"},
			{-1, 0, "✧"}, {1, 0, "✧"},
			{0, -1, "·"},
		}
	case 2:
		return []particle{
			{-2, 0, "·"}, {2, 0, "·"},
			{-1, -1, "✧"}, {1, -1, "✧"},
			{-1, 1, "*"}, {1, 1, "*"},
			{0, 0, " "},
		}
	default:
		return []particle{
			{-3, 0, "."}, {3, 0, "."},
			{-2, -1, "."}, {2, -1, "."},
			{0, 0, " "},
		}
	}
}

func viewFallingGameOver(m model) string {
	gameOver := styleLife.Render("GAME OVER")

	scoreNum := styleBigWPM.Render(fmt.Sprintf("%d", m.fallingScore))
	scoreLabel := styleHint.Render(" words destroyed")

	elapsed := time.Since(m.fallingStartTime).Seconds()
	timeStat := styleStatLabel.Render("survived     ") + styleStatValue.Render(fmt.Sprintf("%.0fs", elapsed))

	hint := styleHint.Render("tab/enter restart  esc menu")

	content := lipgloss.JoinVertical(lipgloss.Left,
		gameOver,
		"",
		scoreNum+scoreLabel,
		"",
		timeStat,
		"",
		hint,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
