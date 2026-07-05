package game

import (
	"io/fs"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/audio"
	"github.com/anpiex12/kebab-ali/internal/entities"
	"github.com/anpiex12/kebab-ali/internal/gfx"
	"github.com/anpiex12/kebab-ali/internal/i18n"
	"github.com/anpiex12/kebab-ali/internal/save"
)

// RepoURL is shown on the credits screen and opened in the browser.
const RepoURL = "https://github.com/anpiex12/kebab-ali"

// Game is the top-level Ebitengine game and holds all shared resources.
type Game struct {
	scene Scene
	next  Scene

	Audio       *audio.Manager
	Lang        *i18n.Catalog
	Store       *save.Store
	Settings    save.Settings
	Leaderboard save.Leaderboard
	Assets      *Assets
	Version     string

	// Character chosen in the menu, carried into play.
	Character entities.CharKind

	canvas *ebiten.Image
	frame  int
	quit   bool
	cap    *captureDirector
}

// New builds the game from the embedded asset file system and version string.
// It loads persisted settings, the leaderboard and the localisation catalog,
// then starts on the splash scene.
func New(assetsFS fs.FS, version string) *Game {
	g := &Game{
		Version: version,
		canvas:  ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight),
	}

	if store, err := save.DefaultStore(); err == nil {
		g.Store = store
		g.Settings = store.LoadSettings()
		g.Leaderboard = store.LoadLeaderboard()
	} else {
		g.Settings = save.DefaultSettings()
	}

	if cat, err := i18n.Load(assetsFS, "assets/lang"); err == nil {
		g.Lang = cat
	} else {
		g.Lang = i18n.New()
	}
	g.Lang.SetLanguage(g.Settings.Language)

	if g.Settings.Character == save.CharMehmet {
		g.Character = entities.Mehmet
	}

	g.Audio = audio.NewManager()
	g.Audio.SetMuted(g.Settings.Muted)

	ebiten.SetFullscreen(g.Settings.Fullscreen)

	g.scene = newSplashScene()
	return g
}

// SwitchTo queues a scene change that takes effect at the end of the frame.
func (g *Game) SwitchTo(s Scene) { g.next = s }

// Frame returns the global frame counter, handy for animations.
func (g *Game) Frame() int { return g.frame }

// T is shorthand for translating a key in the active language.
func (g *Game) T(key string, args ...any) string { return g.Lang.T(key, args...) }

// Quit signals the game loop to terminate cleanly.
func (g *Game) Quit() { g.quit = true }

// Update advances the active scene and applies any queued scene switch.
func (g *Game) Update() error {
	g.frame++
	if g.quit {
		return ebiten.Termination
	}
	if g.scene != nil {
		if err := g.scene.Update(g); err != nil {
			return err
		}
	}
	if g.next != nil {
		g.scene = g.next
		g.next = nil
		if e, ok := g.scene.(entering); ok {
			e.Enter(g)
		}
	}
	g.maybeCapture()
	return nil
}

// Draw renders the active scene into the low-res canvas, then scales it up to
// the window with nearest-neighbour filtering and letterboxing so the pixels
// stay crisp at any size.
func (g *Game) Draw(screen *ebiten.Image) {
	g.canvas.Clear()
	if g.scene != nil {
		g.scene.Draw(g, g.canvas)
	}
	g.maybeSaveShot()
	screen.Fill(gfx.Black)

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	scale := math.Min(float64(sw)/gfx.ScreenWidth, float64(sh)/gfx.ScreenHeight)
	if scale < 1 {
		scale = 1
	}
	ox := (float64(sw) - gfx.ScreenWidth*scale) / 2
	oy := (float64(sh) - gfx.ScreenHeight*scale) / 2

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(ox, oy)
	screen.DrawImage(g.canvas, op)
}

// Layout runs the game at the window's own resolution; the fixed-size canvas is
// scaled manually in Draw.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// --- shared helpers used by multiple scenes --------------------------------

// ToggleMute flips mute, updates audio and persists the setting.
func (g *Game) ToggleMute() {
	g.Settings.Muted = g.Audio.ToggleMute()
	g.persistSettings()
}

// ToggleFullscreen flips fullscreen and persists the setting.
func (g *Game) ToggleFullscreen() {
	g.Settings.Fullscreen = !g.Settings.Fullscreen
	ebiten.SetFullscreen(g.Settings.Fullscreen)
	g.persistSettings()
}

// CycleLanguage switches to the next available language and persists it.
func (g *Game) CycleLanguage() {
	langs := g.Lang.Languages()
	cur := g.Lang.Language()
	idx := 0
	for i, l := range langs {
		if l == cur {
			idx = i
			break
		}
	}
	next := langs[(idx+1)%len(langs)]
	g.Lang.SetLanguage(next)
	g.Settings.Language = next
	g.persistSettings()
}

// SetCharacter records the chosen brother and persists it.
func (g *Game) SetCharacter(c entities.CharKind) {
	g.Character = c
	if c == entities.Mehmet {
		g.Settings.Character = save.CharMehmet
	} else {
		g.Settings.Character = save.CharAli
	}
	g.persistSettings()
}

// SubmitScore inserts a completed run into the leaderboard, persists it and
// returns the rank achieved (0 if it did not place).
func (g *Game) SubmitScore(name string, score int, seconds float64) int {
	rank := g.Leaderboard.Add(save.Entry{Name: name, Score: score, Seconds: seconds})
	if g.Store != nil {
		_ = g.Store.SaveLeaderboard(g.Leaderboard)
	}
	return rank
}

func (g *Game) persistSettings() {
	if g.Store != nil {
		_ = g.Store.SaveSettings(g.Settings)
	}
}
