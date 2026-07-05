package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/audio"
	"github.com/anpiex12/kebab-ali/internal/gfx"
)

type creditsScene struct{ tick int }

func newCreditsScene() *creditsScene { return &creditsScene{} }

func (s *creditsScene) Enter(g *Game) { g.Audio.PlayMusic(audio.MusicMenu) }

func (s *creditsScene) Update(g *Game) error {
	s.tick++
	if keyPressed(ebiten.KeyO) {
		openBrowser(RepoURL)
	}
	if keyPressed(keysBack...) || keyPressed(keysConfirm...) {
		g.Audio.Play(audio.SFXSelect)
		g.SwitchTo(newMenuScene())
	}
	return nil
}

func (s *creditsScene) Draw(g *Game, canvas *ebiten.Image) {
	fillGradient(canvas, gfx.SkyTop, gfx.SkyBottom)
	drawLogo(canvas, g.Assets, g, screenW/2, 44, s.tick)

	gfx.DrawText(canvas, g.T("credits.made_by"), screenW/2, 128,
		gfx.TextOptions{Size: 11, Color: gfx.Cream, Align: gfx.AlignCenter})
	gfx.DrawText(canvas, g.T("credits.assets"), screenW/2, 144,
		gfx.TextOptions{Size: 10, Color: gfx.Cream, Align: gfx.AlignCenter})

	// The prominent GitHub star call-to-action, flanked by drawn sparkles.
	pulse := 0.6 + 0.4*float64((s.tick/20)%2)
	starColor := gfx.Blend(gfx.Gold, gfx.White, pulse)
	star := g.T("credits.star")
	gfx.DrawTextShadow(canvas, star, screenW/2, 176,
		gfx.TextOptions{Size: 13, Color: starColor, Align: gfx.AlignCenter, Bold: true})
	sw := measure(star, 13)
	drawSparkle(canvas, screenW/2-sw/2-12, 182, 5, gfx.Gold)
	drawSparkle(canvas, screenW/2+sw/2+12, 182, 5, gfx.Gold)

	gfx.DrawText(canvas, RepoURL, screenW/2, 198,
		gfx.TextOptions{Size: 10, Color: gfx.NeonBlue, Align: gfx.AlignCenter})
	gfx.DrawText(canvas, g.T("credits.open"), screenW/2, 214,
		gfx.TextOptions{Size: 10, Color: gfx.Cream, Align: gfx.AlignCenter})

	drawHint(canvas, g.T("credits.thanks"))
}
