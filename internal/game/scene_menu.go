package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/audio"
	"github.com/anpiex12/kebab-ali/internal/entities"
	"github.com/anpiex12/kebab-ali/internal/gfx"
)

const (
	miStart = iota
	miCharacter
	miLeaderboard
	miLanguage
	miCredits
	miQuit
	miCount
)

type menuScene struct {
	sel  int
	tick int
}

func newMenuScene() *menuScene { return &menuScene{} }

func (m *menuScene) Enter(g *Game) { g.Audio.PlayMusic(audio.MusicMenu) }

func (m *menuScene) Update(g *Game) error {
	m.tick++
	switch {
	case keyPressed(keysDown...):
		m.sel = (m.sel + 1) % miCount
		g.Audio.Play(audio.SFXSelect)
	case keyPressed(keysUp...):
		m.sel = (m.sel + miCount - 1) % miCount
		g.Audio.Play(audio.SFXSelect)
	}

	if keyPressed(keysLeft...) || keyPressed(keysRight...) {
		switch m.sel {
		case miCharacter:
			m.toggleCharacter(g)
		case miLanguage:
			g.CycleLanguage()
			g.Audio.Play(audio.SFXSelect)
		}
	}

	if keyPressed(keysConfirm...) {
		m.activate(g)
	}
	if keyPressed(ebiten.KeyM) {
		g.ToggleMute()
	}
	if keyPressed(ebiten.KeyF) {
		g.ToggleFullscreen()
	}
	return nil
}

func (m *menuScene) toggleCharacter(g *Game) {
	if g.Character == entities.Ali {
		g.SetCharacter(entities.Mehmet)
	} else {
		g.SetCharacter(entities.Ali)
	}
	g.Audio.Play(audio.SFXSelect)
}

func (m *menuScene) activate(g *Game) {
	g.Audio.Play(audio.SFXSelect)
	switch m.sel {
	case miStart:
		g.SwitchTo(newStoryScene(g, storyIntro))
	case miCharacter:
		m.toggleCharacter(g)
	case miLeaderboard:
		g.SwitchTo(newLeaderboardScene())
	case miLanguage:
		g.CycleLanguage()
	case miCredits:
		g.SwitchTo(newCreditsScene())
	case miQuit:
		g.Quit()
	}
}

func (m *menuScene) options(g *Game) []string {
	charName := g.T("char.ali")
	if g.Character == entities.Mehmet {
		charName = g.T("char.mehmet")
	}
	langName := g.T("lang." + g.Lang.Language())
	return []string{
		g.T("menu.start"),
		g.T("menu.character", charName),
		g.T("menu.leaderboard"),
		g.T("menu.language", langName),
		g.T("menu.credits"),
		g.T("menu.quit"),
	}
}

func (m *menuScene) Draw(g *Game, canvas *ebiten.Image) {
	fillGradient(canvas, gfx.SkyTop, gfx.SkyBottom)
	drawCitySilhouette(canvas, gfx.Darken(gfx.SkyTop, 0.3), m.tick)
	drawLogo(canvas, g.Assets, g, screenW/2, 50, m.tick)

	drawMenuList(canvas, screenW/2, 138, m.options(g), m.sel, m.tick)

	// Character preview and blurb.
	if g.Assets != nil {
		frames := g.Assets.Frames(g.Character)
		gfx.DrawSprite(canvas, frames.Idle, 40, 150, 3, false)
	}
	if m.sel == miCharacter {
		desc := g.T("char.ali.desc")
		if g.Character == entities.Mehmet {
			desc = g.T("char.mehmet.desc")
		}
		gfx.DrawText(canvas, desc, screenW/2, screenH-32, gfx.TextOptions{Size: 9, Color: gfx.Cream, Align: gfx.AlignCenter})
	}

	// Mute indicator.
	if g.Audio.Muted() {
		gfx.DrawText(canvas, "M "+g.T("common.off"), screenW-6, 6, gfx.TextOptions{Size: 9, Color: gfx.Grey, Align: gfx.AlignRight})
	}
	drawHint(canvas, g.T("menu.hint")+"   M/F")
}
