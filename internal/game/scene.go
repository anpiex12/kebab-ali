// Package game wires the pure simulation, graphics, audio, i18n and save
// packages into an Ebitengine game: a stack-free scene manager, the low-res
// canvas, and every screen from the splash to the credits.
package game

import "github.com/hajimehoshi/ebiten/v2"

// Scene is one screen of the game (splash, menu, play, …). Update advances it a
// frame and may switch scenes via g.SwitchTo; Draw paints into the shared
// low-resolution canvas.
type Scene interface {
	Update(g *Game) error
	Draw(g *Game, canvas *ebiten.Image)
}

// entering is an optional hook a scene can implement to run setup the first
// time it becomes active (after construction, once it owns the game).
type entering interface {
	Enter(g *Game)
}
