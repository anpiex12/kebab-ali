package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/gfx"
)

// drawHUD paints the always-on status bar: lives, taler, power, score and time.
func (s *playScene) drawHUD(g *Game, canvas *ebiten.Image) {
	p := s.player

	// A subtle darkened strip so text stays readable over bright skies.
	gfx.FillRect(canvas, 0, 0, screenW, 22, color.RGBA{0, 0, 0, 0x55})

	left := func(txt string, y float64, c color.RGBA) {
		gfx.DrawTextShadow(canvas, txt, 6, y, gfx.TextOptions{Size: 10, Color: c})
	}
	right := func(txt string, y float64, c color.RGBA) {
		gfx.DrawTextShadow(canvas, txt, screenW-6, y, gfx.TextOptions{Size: 10, Color: c, Align: gfx.AlignRight})
	}

	// Little icons.
	gfx.FillCircle(canvas, 10, 8, 3, gfx.Red) // heart-ish
	left(fmt.Sprintf("  x%d", p.Lives), 3, gfx.Cream)
	gfx.FillCircle(canvas, 10, 18, 3, gfx.Gold) // taler
	left(fmt.Sprintf("  %s %d", g.T("hud.coins"), p.Coins), 13, gfx.Gold)

	right(fmt.Sprintf("%s %06d", g.T("hud.score"), p.Score), 3, gfx.Cream)
	right(fmt.Sprintf("%s %d", g.T("hud.time"), int(s.timeLeft)), 13, gfx.Cream)

	// Power state, centred.
	power := g.T("hud.power." + p.Power.String())
	if p.AyranActive() {
		power = g.T("hud.power.ayran")
	}
	gfx.DrawTextShadow(canvas, power, screenW/2, 6, gfx.TextOptions{Size: 11, Color: gfx.Gold, Align: gfx.AlignCenter, Bold: true})

	if s.bossActive {
		s.drawBossBar(g, canvas)
	}
}

// drawBossBar shows the current boss's combined health and, briefly, its taunt.
func (s *playScene) drawBossBar(g *Game, canvas *ebiten.Image) {
	hp, max := 0, 0
	name := ""
	phase := 0
	for _, b := range s.bosses {
		hp += b.HP()
		max += b.MaxHP()
		name = g.T(b.NameKey())
		phase = b.Phase()
	}
	if max == 0 {
		return
	}
	barW := 180.0
	x := (screenW - barW) / 2
	y := 30.0
	gfx.DrawTextShadow(canvas, name, screenW/2, y-12, gfx.TextOptions{Size: 10, Color: gfx.Red, Align: gfx.AlignCenter, Bold: true})
	gfx.FillRect(canvas, x, y, barW, 6, gfx.GreyDark)
	frac := float64(hp) / float64(max)
	gfx.FillRect(canvas, x, y, barW*frac, 6, gfx.Red)
	gfx.StrokeRect(canvas, x, y, barW, 6, 1, gfx.Black)
	if phase > 1 {
		gfx.DrawText(canvas, fmt.Sprintf("%d", phase), x+barW+8, y-2, gfx.TextOptions{Size: 9, Color: gfx.Gold})
	}

	// Taunt speech bubble over the first living boss.
	if s.bossTaunt > 0 {
		for _, b := range s.bosses {
			if b.Alive() {
				r := b.Rect()
				drawSpeechBubble(canvas, r.CenterX()-s.camX, r.Y-s.camY, g.T(b.NameKey()), g.T(b.TauntKey()))
				break
			}
		}
	}
}

// drawOverlays paints the state-specific full-screen overlays.
func (s *playScene) drawOverlays(g *Game, canvas *ebiten.Image) {
	switch s.state {
	case psBanner:
		dim(canvas, 0x88)
		gfx.DrawTextShadow(canvas, g.T("level.banner", s.levelIdx+1), screenW/2, 90,
			gfx.TextOptions{Size: 26, Color: gfx.Gold, Align: gfx.AlignCenter, Bold: true})
		gfx.DrawTextShadow(canvas, g.T(s.lvl.Name), screenW/2, 124,
			gfx.TextOptions{Size: 14, Color: gfx.Cream, Align: gfx.AlignCenter})
		gfx.DrawText(canvas, g.T("level.get_ready"), screenW/2, 150,
			gfx.TextOptions{Size: 11, Color: gfx.White, Align: gfx.AlignCenter})
	case psCleared:
		dim(canvas, 0x99)
		gfx.DrawTextShadow(canvas, g.T("level.cleared"), screenW/2, 100,
			gfx.TextOptions{Size: 22, Color: gfx.Gold, Align: gfx.AlignCenter, Bold: true})
		gfx.DrawText(canvas, g.T("level.time_bonus", int(s.timeLeft)*10), screenW/2, 134,
			gfx.TextOptions{Size: 12, Color: gfx.Cream, Align: gfx.AlignCenter})
	case psPaused:
		s.drawPauseMenu(g, canvas)
	}
}

func (s *playScene) drawPauseMenu(g *Game, canvas *ebiten.Image) {
	dim(canvas, 0xB0)
	gfx.DrawTextShadow(canvas, g.T("pause.title"), screenW/2, 70,
		gfx.TextOptions{Size: 24, Color: gfx.Gold, Align: gfx.AlignCenter, Bold: true})
	sound := g.T("common.on")
	if g.Audio.Muted() {
		sound = g.T("common.off")
	}
	opts := []string{g.T("pause.resume"), g.T("pause.restart"), g.T("pause.menu")}
	drawMenuList(canvas, screenW/2, 120, opts, s.pauseSel, s.tick)
	gfx.DrawText(canvas, g.T("pause.sound", sound), screenW/2, screenH-30,
		gfx.TextOptions{Size: 10, Color: gfx.Grey, Align: gfx.AlignCenter})
	drawHint(canvas, "M   ESC")
}

// dim overlays a translucent black veil of the given alpha.
func dim(canvas *ebiten.Image, alpha uint8) {
	gfx.FillRect(canvas, 0, 0, screenW, screenH, color.RGBA{0, 0, 0, alpha})
}
