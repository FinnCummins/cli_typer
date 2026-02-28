# cli_typer

A terminal typing game built with Go and [Bubbletea](https://github.com/charmbracelet/bubbletea). Two game modes: a monkeytype-style typing test and a falling words arcade game with ASCII art aliens, sound effects, and a day/night cycle.

## Install

Requires [Go](https://go.dev/dl/) 1.25+.

```bash
go install github.com/FinnCummins/cli_typer@latest
```

Or clone and build from source:

```bash
git clone https://github.com/FinnCummins/cli_typer.git
cd cli_typer
go build -o cli_typer .
./cli_typer
```

## Game Modes

### Classic Typing Test

A clean, monkeytype-inspired typing test. Words appear in a 3-line scrolling window. Type them as fast and accurately as you can before the timer runs out.

- Choose between **random words** or **famous quotes**
- Timed: **15s**, **30s**, or **60s**
- Live WPM counter while you type
- Results: net WPM, accuracy, characters, words

**Controls:**
- Type normally to begin (timer starts on first keypress)
- `space` — advance to next word
- `backspace` — delete within current word
- `tab` — restart
- `esc` — back to menu

### Falling Words

Words wrapped in ASCII art aliens fall from the sky toward your shield. Type each word to lock on and destroy it with a laser before it breaks through. You have 3 lives.

```
  ╱◉‾‾‾◉╲       ¤◉───◉¤
  >{the}<        ◀[quick]▶
```

- **Turret** on the shield tracks your target and slides toward it as you type
- **Laser beam** fires from the turret to the alien on word completion
- **Explosion** particles burst where the alien was
- **Difficulty ramps** — words fall faster and spawn more frequently over time
- **Sound effects** — destroy, shield hit, game over (embedded OGG via [beep](https://github.com/gopxl/beep))
- **Day/night cycle** (optional) — sun and moon arc across the sky, terminal background shifts from white (day) to black (night)

**Controls:**
- Start typing to target the lowest matching word
- Complete the word to destroy it (no space needed)
- `backspace` — fix mistakes or release target
- `tab` — restart
- `esc` — back to menu

## Menu

Navigate with arrow keys (or `hjkl`), change options with left/right, press `enter` to start.

```
cli_typer

▸ game      [ classic ]  falling
  words     [ words ]    quotes
  duration  [ 30s ]
```

When falling mode is selected, the duration row is replaced with a day/night cycle toggle.

## Sound Effects

Sounds are from [Kenney's Interface Sounds](https://kenney.nl/assets/interface-sounds) (CC0 public domain) and are embedded in the binary at compile time — no external files needed.

## Credits

Built with [Bubbletea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss), [Bubbles](https://github.com/charmbracelet/bubbles), and [Beep](https://github.com/gopxl/beep).

Sound effects by [Kenney](https://kenney.nl/).
