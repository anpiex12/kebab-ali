package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/audio"
	"github.com/anpiex12/kebab-ali/internal/gfx"
)

// splashScene shows the loading screen with the spinning spit logo and builds
// the GPU asset atlas on its first frame (safe now the game loop is running).
type splashScene struct {
	timer  int
	loaded bool
}

func newSplashScene() *splashScene { return &splashScene{} }

func (s *splashScene) Update(g *Game) error {
	if !s.loaded {
		g.Assets = LoadAssets()
		g.Audio.PlayMusic(audio.MusicMenu)
		s.loaded = true
	}
	s.timer++
	ready := s.timer > 60
	if ready && (keyPressed(keysConfirm...) || keyPressed(keysBack...)) || s.timer > 300 {
		g.SwitchTo(newMenuScene())
	}
	return nil
}

func (s *splashScene) Draw(g *Game, canvas *ebiten.Image) {
	fillGradient(canvas, gfx.SkyTop, gfx.SkyBottom)
	drawLogo(canvas, g.Assets, g, screenW/2, 90, s.timer)
	if s.timer > 60 && (s.timer/30)%2 == 0 {
		drawHint(canvas, g.T("splash.press_any"))
	} else if s.timer <= 60 {
		gfx.DrawText(canvas, g.T("splash.loading"), screenW/2, screenH-40,
			gfx.TextOptions{Size: 12, Color: gfx.Cream, Align: gfx.AlignCenter})
	}
}
