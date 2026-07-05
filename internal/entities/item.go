package entities

import "github.com/anpiex12/kebab-ali/internal/physics"

// ItemKind identifies a collectible.
type ItemKind int

const (
	// ItemTaler is a golden coin; 100 grant an extra life.
	ItemTaler ItemKind = iota
	// ItemSpit is the döner-spit power-up.
	ItemSpit
	// ItemAyran is the rare invincibility bottle.
	ItemAyran
)

// Item is a collectible. Coins sit still and bob (in the renderer); power-ups
// pop out of a block and then slide along the ground like the classic mushroom.
type Item struct {
	Kind     ItemKind
	Body     physics.Rect
	VX, VY   float64
	AnimTick int

	moves bool
	alive bool
}

// NewTaler creates a stationary coin at (x,y).
func NewTaler(x, y float64) *Item {
	return &Item{Kind: ItemTaler, Body: physics.Rect{X: x, Y: y, W: 10, H: 10}, alive: true}
}

// NewSpit creates a döner-spit power-up that pops up and then walks.
func NewSpit(x, y float64) *Item {
	return &Item{Kind: ItemSpit, Body: physics.Rect{X: x, Y: y, W: 12, H: 14}, VX: 0.6, VY: -2.5, moves: true, alive: true}
}

// NewAyran creates an Ayran bottle that pops up and then walks.
func NewAyran(x, y float64) *Item {
	return &Item{Kind: ItemAyran, Body: physics.Rect{X: x, Y: y, W: 10, H: 14}, VX: 0.7, VY: -2.5, moves: true, alive: true}
}

// Alive reports whether the item is still uncollected and present.
func (it *Item) Alive() bool { return it.alive }

// Rect returns the collision box.
func (it *Item) Rect() physics.Rect { return it.Body }

// Collect removes the item once the player has taken it.
func (it *Item) Collect() { it.alive = false }

// Update advances the item. Coins only animate; power-ups fall, walk and bounce
// off walls.
func (it *Item) Update(w World) {
	if !it.alive {
		return
	}
	it.AnimTick++
	if !it.moves {
		return
	}
	ts := w.TileSize()
	it.VY += playerGravity
	if it.VY > playerMaxFall {
		it.VY = playerMaxFall
	}
	var hit bool
	it.Body, hit = physics.ResolveX(it.Body, it.VX, ts, w.Solid)
	if hit {
		it.VX = -it.VX
	}
	it.Body, hit = physics.ResolveY(it.Body, it.VY, ts, w.Solid)
	if hit {
		it.VY = 0
	}
}
