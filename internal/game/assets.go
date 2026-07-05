package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/entities"
	"github.com/anpiex12/kebab-ali/internal/gfx"
)

// CharFrames is the GPU-side animation set for a character.
type CharFrames struct {
	Idle *ebiten.Image
	Jump *ebiten.Image
	Run  []*ebiten.Image
}

// Assets holds all sprites once uploaded to the GPU. It must be built after the
// graphics driver is live (i.e. from within the running game loop), so the
// splash scene constructs it on its first frame.
type Assets struct {
	Ali    CharFrames
	Mehmet CharFrames

	Taler     *ebiten.Image
	Spit      *ebiten.Image
	Ayran     *ebiten.Image
	MeatSlice *ebiten.Image
}

// LoadAssets rasterises every procedural sprite and uploads it.
func LoadAssets() *Assets {
	buildChar := func(cs gfx.CharacterSprites) CharFrames {
		run := make([]*ebiten.Image, len(cs.Run))
		for i, r := range cs.Run {
			run[i] = gfx.ToEbiten(r)
		}
		return CharFrames{
			Idle: gfx.ToEbiten(cs.Idle),
			Jump: gfx.ToEbiten(cs.Jump),
			Run:  run,
		}
	}
	return &Assets{
		Ali:       buildChar(gfx.BuildCharacter(gfx.Red, gfx.RedDark)),
		Mehmet:    buildChar(gfx.BuildCharacter(gfx.Green, gfx.GreenDark)),
		Taler:     gfx.ToEbiten(gfx.Taler()),
		Spit:      gfx.ToEbiten(gfx.DoenerSpit()),
		Ayran:     gfx.ToEbiten(gfx.Ayran()),
		MeatSlice: gfx.ToEbiten(gfx.MeatSlice()),
	}
}

// Frames returns the animation set for the given character.
func (a *Assets) Frames(k entities.CharKind) CharFrames {
	if k == entities.Mehmet {
		return a.Mehmet
	}
	return a.Ali
}
