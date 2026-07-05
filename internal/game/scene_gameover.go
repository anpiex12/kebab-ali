package game

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/audio"
	"github.com/anpiex12/kebab-ali/internal/gfx"
)

// gameOverScene handles both the "game over" and "you won" end states, prompting
// for a name when the score reaches the leaderboard.
type gameOverScene struct {
	res       GameResult
	name      []rune
	entering  bool
	submitted bool
	rank      int
	tick      int
}

func newGameOverScene(res GameResult) *gameOverScene { return &gameOverScene{res: res} }

func (s *gameOverScene) Enter(g *Game) {
	s.entering = g.Leaderboard.Qualifies(s.res.Score)
	if s.res.Won {
		g.Audio.PlayMusic(audio.MusicVictory)
	} else {
		g.Audio.PlayMusic(audio.MusicMenu)
	}
}

func (s *gameOverScene) Update(g *Game) error {
	s.tick++
	if s.entering && !s.submitted {
		for _, r := range ebiten.AppendInputChars(nil) {
			if len(s.name) < 10 && r >= 32 && r < 127 {
				s.name = append(s.name, unicode.ToUpper(r))
			}
		}
		if keyPressed(ebiten.KeyBackspace) && len(s.name) > 0 {
			s.name = s.name[:len(s.name)-1]
		}
		if keyPressed(ebiten.KeyEnter, ebiten.KeyNumpadEnter) {
			name := strings.TrimSpace(string(s.name))
			if name == "" {
				name = "ALI"
			}
			s.rank = g.SubmitScore(name, s.res.Score, s.res.Seconds)
			s.submitted = true
			g.Audio.Play(audio.SFXOneUp)
		}
		return nil
	}
	if keyPressed(keysConfirm...) || keyPressed(ebiten.KeyEscape) {
		g.Audio.Play(audio.SFXSelect)
		if s.submitted {
			g.SwitchTo(newLeaderboardSceneHighlight(s.rank))
		} else {
			g.SwitchTo(newMenuScene())
		}
	}
	return nil
}

func (s *gameOverScene) Draw(g *Game, canvas *ebiten.Image) {
	top := gfx.Darken(gfx.SkyTop, 0.3)
	if s.res.Won {
		top = gfx.Darken(gfx.Gold, 0.2)
	}
	fillGradient(canvas, top, gfx.SkyBottom)

	title := g.T("gameover.title")
	titleColor := gfx.Red
	if s.res.Won {
		title = g.T("story.win.4")
		titleColor = gfx.Gold
	}
	gfx.DrawTextShadow(canvas, title, screenW/2, 44,
		gfx.TextOptions{Size: 26, Color: titleColor, Align: gfx.AlignCenter, Bold: true})

	gfx.DrawText(canvas, fmt.Sprintf("%s %d", g.T("hud.score"), s.res.Score), screenW/2, 92,
		gfx.TextOptions{Size: 14, Color: gfx.Cream, Align: gfx.AlignCenter})
	gfx.DrawText(canvas, fmt.Sprintf("%s %s", g.T("hud.time"), formatTime(s.res.Seconds)), screenW/2, 112,
		gfx.TextOptions{Size: 12, Color: gfx.Cream, Align: gfx.AlignCenter})

	switch {
	case s.entering && !s.submitted:
		gfx.DrawText(canvas, g.T("name.prompt"), screenW/2, 150,
			gfx.TextOptions{Size: 12, Color: gfx.Gold, Align: gfx.AlignCenter})
		field := string(s.name)
		if (s.tick/20)%2 == 0 {
			field += "_"
		}
		drawPanel(canvas, screenW/2-90, 170, 180, 26)
		gfx.DrawText(canvas, field, screenW/2, 176,
			gfx.TextOptions{Size: 16, Color: gfx.White, Align: gfx.AlignCenter, Bold: true})
		drawHint(canvas, g.T("name.hint"))
	case s.submitted:
		gfx.DrawTextShadow(canvas, fmt.Sprintf("#%d", s.rank), screenW/2, 158,
			gfx.TextOptions{Size: 20, Color: gfx.Gold, Align: gfx.AlignCenter, Bold: true})
		drawHint(canvas, g.T("common.continue"))
	default:
		gfx.DrawText(canvas, g.T("gameover.retry"), screenW/2, 158,
			gfx.TextOptions{Size: 11, Color: gfx.Cream, Align: gfx.AlignCenter})
		drawHint(canvas, g.T("gameover.menu"))
	}
}
