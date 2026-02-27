package main

// Falling words game mode with 2-row alien sprites:
//
//   ╱‾‾‾╲      <- head row (y-1)
//   >{the}<    <- body row (y, where the word text lives)
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
	edgePadding     = 7
	turretSpeed     = 3
	laserDuration   = 3
	explodeDuration = 4
	wingWidth       = 2 // characters per wing on each side
)

type fallingWord struct {
	word   string
	x      int     // column of the word text (not the wings)
	y      float64 // row of the body (head is at y-1)
	typed  int
	active bool
}

// totalWidth returns the full width of this alien including wings.
func (fw fallingWord) totalWidth() int {
	return wingWidth + len(fw.word) + wingWidth
}

// leftEdge returns the leftmost column this alien occupies.
func (fw fallingWord) leftEdge() int {
	return fw.x - wingWidth
}

// rightEdge returns one past the rightmost column (exclusive).
func (fw fallingWord) rightEdge() int {
	return fw.x + len(fw.word) + wingWidth
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

// overlapsExisting checks if a new word at position x would overlap
// with any existing word that's still near the top of the screen.
// Accounts for the full 2-row alien width (wings included).
func overlapsExisting(m model, word string, x int) bool {
	newLeft := x - wingWidth
	newRight := x + len(word) + wingWidth

	for _, fw := range m.fallingWords {
		// Only check words near the top (within 3 rows of spawn point)
		if fw.y > 3 {
			continue
		}
		existLeft := fw.leftEdge()
		existRight := fw.rightEdge()

		// Check X overlap with 1 char gap
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

	minX := edgePadding
	maxX := m.width - len(word) - edgePadding
	if maxX <= minX {
		maxX = minX + 1
	}

	// Try up to 10 random positions to find one that doesn't overlap
	var x int
	placed := false
	for attempt := 0; attempt < 10; attempt++ {
		x = rand.Intn(maxX-minX) + minX
		if !overlapsExisting(m, word, x) {
			placed = true
			break
		}
	}

	if !placed {
		// All positions overlap — skip this spawn, try next tick
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
				m.turretStartX = m.turretX // remember where turret was
			}
		} else if m.fallingTarget < len(m.fallingWords) {
			m.fallingWords[m.fallingTarget].typed = len(m.fallingInput)
		}

		// Move turret proportionally — each keypress covers an equal fraction
		// of the distance from where the turret started to the target center.
		if m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
			fw := m.fallingWords[m.fallingTarget]
			targetX := fw.x + len(fw.word)/2
			wordLen := len([]rune(fw.word))
			if wordLen > 0 {
				progress := float64(len(m.fallingInput)) / float64(wordLen)
				m.turretX = m.turretStartX + int(progress*float64(targetX-m.turretStartX))
			}
		}

		if m.fallingTarget >= 0 && m.fallingTarget < len(m.fallingWords) {
			fw := m.fallingWords[m.fallingTarget]
			if string(m.fallingInput) == fw.word {
				wordCenterX := fw.x + len(fw.word)/2
				wordRow := int(fw.y)

				playHeight := m.height - 6
				if playHeight < 5 {
					playHeight = 5
				}

				m.laser = &laserBeam{
					x:     wordCenterX,
					fromY: playHeight,
					toY:   wordRow - 1, // laser reaches the head row
					ticks: laserDuration,
				}
				m.explosions = append(m.explosions, explosion{
					x:     wordCenterX,
					y:     wordRow,
					ticks: explodeDuration,
				})

				m.turretX = wordCenterX
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

// --- 2-Row Alien Sprites ---
//
// Each alien has a head row and a body row.
// Head:  ╱‾‾‾╲   (width matches body)
// Body:  >{the}<  (wings + word)
//
// The head is dynamically generated to match the word's total width.

type alienSprite struct {
	bodyLeft  string // left wing on body row
	bodyRight string // right wing on body row
	headLeft  string // left frame of head (includes eye)
	headFill  rune   // character repeated between eyes
	headRight string // right frame of head (includes eye)
}

// 4 alien designs — each has a distinct silhouette.
// The head row stretches to match the word width.
// Head caps are 2 chars each (matching wing width), so fill = word length.
var alienSprites = []alienSprite{
	// Classic invader — ridge with peeking eyes
	{">{", "}<", "╱◉", '‾', "◉╲"},
	// Antenna bug — antennae with big eyes
	{"◀[", "]▶", "¤◉", '─', "◉¤"},
	// Jellyfish — wavy with round eyes
	{"({", "})", "~◎", '~', "◎~"},
	// Robot — boxy with diamond eyes
	{"╞{", "}╡", "[◈", '·', "◈]"},
}

func spriteForWord(word string) int {
	if len(word) == 0 {
		return 0
	}
	return int(word[0]) % len(alienSprites)
}

// buildHead generates the head row string for a given word and sprite.
// It matches the total width of the body row (wings + word).
func buildHead(word string, sprite alienSprite) string {
	// Body width: len(bodyLeft) + len(word) + len(bodyRight)
	// Head width: len(headLeft) + fill + len(headRight) should equal body width
	bodyWidth := len([]rune(sprite.bodyLeft)) + len([]rune(word)) + len([]rune(sprite.bodyRight))
	headCapWidth := len([]rune(sprite.headLeft)) + len([]rune(sprite.headRight))
	fillCount := bodyWidth - headCapWidth
	if fillCount < 0 {
		fillCount = 0
	}
	return sprite.headLeft + strings.Repeat(string(sprite.headFill), fillCount) + sprite.headRight
}

// --- Rendering ---

func renderShield(width int, lives int, turretX int) string {
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
			result.WriteString(styleShield.Render(s))
		} else if lives >= 2 {
			result.WriteString(styleShield.Render(s))
		} else if lives == 1 {
			result.WriteString(styleShieldDamaged.Render(s))
		} else {
			result.WriteString(styleHint.Render(s))
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

	// Build 2D grid
	grid := make([][]string, playHeight)
	for row := range grid {
		grid[row] = make([]string, playWidth)
		for col := range grid[row] {
			grid[row][col] = " "
		}
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

	// Place 2-row alien sprites
	for _, fw := range m.fallingWords {
		bodyRow := int(fw.y)
		headRow := bodyRow - 1
		sprite := alienSprites[spriteForWord(fw.word)]
		headStr := buildHead(fw.word, sprite)
		headRunes := []rune(headStr)

		alienStyle := styleAlien
		if fw.active {
			alienStyle = styleAlienActive
		}

		// --- Head row ---
		if headRow >= 0 && headRow < playHeight {
			for i, ch := range headRunes {
				col := fw.x - wingWidth + i
				if col >= 0 && col < playWidth {
					grid[headRow][col] = alienStyle.Render(string(ch))
				}
			}
		}

		// --- Body row ---
		if bodyRow >= 0 && bodyRow < playHeight {
			// Left wing
			bodyLeftRunes := []rune(sprite.bodyLeft)
			for i, ch := range bodyLeftRunes {
				col := fw.x - len(bodyLeftRunes) + i
				if col >= 0 && col < playWidth {
					grid[bodyRow][col] = alienStyle.Render(string(ch))
				}
			}

			// Word text
			for j, ch := range []rune(fw.word) {
				col := fw.x + j
				if col < 0 || col >= playWidth {
					continue
				}
				if fw.active && j < fw.typed {
					grid[bodyRow][col] = styleCorrect.Render(string(ch))
				} else if fw.active {
					grid[bodyRow][col] = styleCursor.Render(string(ch))
				} else {
					grid[bodyRow][col] = styleUntyped.Render(string(ch))
				}
			}

			// Right wing
			bodyRightRunes := []rune(sprite.bodyRight)
			for i, ch := range bodyRightRunes {
				col := fw.x + len([]rune(fw.word)) + i
				if col >= 0 && col < playWidth {
					grid[bodyRow][col] = alienStyle.Render(string(ch))
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

	shield := renderShield(playWidth, m.fallingLives, m.turretX)

	hearts := styleLife.Render(strings.Repeat("♥ ", m.fallingLives))
	if m.fallingLives == 0 {
		hearts = styleHint.Render("♥ ♥ ♥")
	}
	scoreText := styleStatLabel.Render("score ") + styleStatValue.Render(fmt.Sprintf("%d", m.fallingScore))
	elapsed := time.Since(m.fallingStartTime).Seconds()
	timeText := styleStatLabel.Render("time ") + styleStatValue.Render(fmt.Sprintf("%.0fs", elapsed))
	statusBar := hearts + "  " + scoreText + "  " + timeText

	inputStr := string(m.fallingInput)
	inputDisplay := styleHighlight.Render("> ") + styleCorrect.Render(inputStr) + styleCursor.Render("_")

	hint := styleHint.Render("tab restart  esc menu")

	if m.fallingGameOver {
		return viewFallingGameOver(m)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		statusBar,
		playField,
		shield,
		inputDisplay,
		hint,
	)
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
