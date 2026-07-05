package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/audio"
	"github.com/anpiex12/kebab-ali/internal/entities"
	"github.com/anpiex12/kebab-ali/internal/gfx"
)

type storyKind int

const (
	storyIntro storyKind = iota
	storyWin
)

// GameResult is a completed run's outcome, carried into the name-entry screen.
type GameResult struct {
	Score   int
	Seconds float64
	Won     bool
}

// storyScene shows a sequence of full-screen narration pages (intro and the
// final rescue cutscene), then runs onDone.
type storyScene struct {
	pages  []string
	music  string
	win    bool
	page   int
	tick   int
	onDone func(*Game) Scene
}

func charName(g *Game) string {
	if g.Character == entities.Mehmet {
		return g.T("char.mehmet")
	}
	return g.T("char.ali")
}

func newStoryScene(g *Game, kind storyKind) *storyScene {
	name := charName(g)
	if kind == storyIntro {
		return &storyScene{
			pages: []string{
				g.T("story.intro.1"),
				g.T("story.intro.2"),
				g.T("story.intro.3"),
				g.T("story.intro.4", name),
			},
			music:  audio.MusicMenu,
			onDone: func(g *Game) Scene { return newPlayScene(g, 0, nil) },
		}
	}
	return nil
}

// newVictoryScene shows the rescue cutscene, then goes to name entry with the
// finished run's result.
func newVictoryScene(g *Game, res GameResult) *storyScene {
	name := charName(g)
	return &storyScene{
		pages: []string{
			g.T("story.win.1"),
			g.T("story.win.2", name),
			g.T("story.win.3"),
			g.T("story.win.4"),
		},
		music:  audio.MusicVictory,
		win:    true,
		onDone: func(g *Game) Scene { return newGameOverScene(res) },
	}
}

func (s *storyScene) Enter(g *Game) { g.Audio.PlayMusic(s.music) }

func (s *storyScene) Update(g *Game) error {
	s.tick++
	if keyPressed(keysConfirm...) {
		g.Audio.Play(audio.SFXSelect)
		s.page++
		if s.page >= len(s.pages) {
			g.SwitchTo(s.onDone(g))
		}
	}
	if keyPressed(ebiten.KeyM) {
		g.ToggleMute()
	}
	return nil
}

func (s *storyScene) Draw(g *Game, canvas *ebiten.Image) {
	top, bottom := gfx.SkyTop, gfx.Darken(gfx.SkyBottom, 0.4)
	if s.win {
		top, bottom = gfx.Darken(gfx.Gold, 0.3), gfx.SkyBottom
	}
	fillGradient(canvas, top, bottom)

	if s.page < len(s.pages) {
		// Show all pages up to the current one, fading in the newest.
		for i := 0; i <= s.page && i < len(s.pages); i++ {
			y := 80 + float64(i)*30
			clr := gfx.Cream
			if i == s.page {
				clr = gfx.Gold
			}
			gfx.DrawTextShadow(canvas, s.pages[i], screenW/2, y, gfx.TextOptions{
				Size: 14, Color: clr, Align: gfx.AlignCenter,
			})
		}
	}
	if (s.tick/30)%2 == 0 {
		drawHint(canvas, g.T("common.continue"))
	}
}
