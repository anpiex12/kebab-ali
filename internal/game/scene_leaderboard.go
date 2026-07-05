package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/audio"
	"github.com/anpiex12/kebab-ali/internal/gfx"
)

type leaderboardScene struct {
	tick      int
	highlight int // 1-based rank to highlight (just-added entry), 0 for none
}

func newLeaderboardScene() *leaderboardScene { return &leaderboardScene{} }

func newLeaderboardSceneHighlight(rank int) *leaderboardScene {
	return &leaderboardScene{highlight: rank}
}

func (s *leaderboardScene) Enter(g *Game) { g.Audio.PlayMusic(audio.MusicMenu) }

func (s *leaderboardScene) Update(g *Game) error {
	s.tick++
	if keyPressed(keysBack...) || keyPressed(keysConfirm...) {
		g.Audio.Play(audio.SFXSelect)
		g.SwitchTo(newMenuScene())
	}
	return nil
}

// formatTime renders seconds as M:SS.
func formatTime(seconds float64) string {
	total := int(seconds)
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

func (s *leaderboardScene) Draw(g *Game, canvas *ebiten.Image) {
	fillGradient(canvas, gfx.SkyTop, gfx.SkyBottom)
	gfx.DrawTextShadow(canvas, g.T("lb.title"), screenW/2, 24,
		gfx.TextOptions{Size: 24, Color: gfx.Gold, Align: gfx.AlignCenter, Bold: true})

	entries := g.Leaderboard.Entries
	if len(entries) == 0 {
		gfx.DrawText(canvas, g.T("lb.empty"), screenW/2, screenH/2,
			gfx.TextOptions{Size: 12, Color: gfx.Cream, Align: gfx.AlignCenter})
		drawHint(canvas, g.T("lb.back"))
		return
	}

	// Column header.
	y := 62.0
	head := gfx.TextOptions{Size: 9, Color: gfx.Grey}
	gfx.DrawText(canvas, g.T("lb.rank"), 60, y, head)
	gfx.DrawText(canvas, g.T("lb.name"), 96, y, head)
	gfx.DrawText(canvas, g.T("lb.score"), 300, y, gfx.TextOptions{Size: 9, Color: gfx.Grey, Align: gfx.AlignRight})
	gfx.DrawText(canvas, g.T("lb.time"), 400, y, gfx.TextOptions{Size: 9, Color: gfx.Grey, Align: gfx.AlignRight})

	for i, e := range entries {
		ry := 78 + float64(i)*17
		clr := gfx.Cream
		if i+1 == s.highlight {
			clr = gfx.Gold
			gfx.FillRect(canvas, 50, ry-2, 360, 15, gfx.Blend(gfx.Gold, gfx.Black, 0.7))
		}
		gfx.DrawText(canvas, fmt.Sprintf("%d.", i+1), 60, ry, gfx.TextOptions{Size: 11, Color: clr, Align: gfx.AlignRight})
		gfx.DrawText(canvas, e.Name, 96, ry, gfx.TextOptions{Size: 11, Color: clr})
		gfx.DrawText(canvas, fmt.Sprintf("%d", e.Score), 300, ry, gfx.TextOptions{Size: 11, Color: clr, Align: gfx.AlignRight})
		gfx.DrawText(canvas, formatTime(e.Seconds), 400, ry, gfx.TextOptions{Size: 11, Color: clr, Align: gfx.AlignRight})
	}
	drawHint(canvas, g.T("lb.back"))
}
