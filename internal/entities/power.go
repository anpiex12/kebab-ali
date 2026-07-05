// Package entities holds the game's simulation: the player, enemies, bosses,
// items and projectiles. It is deliberately free of any Ebitengine/rendering
// dependency — every type is pure state plus an Update method driven through a
// small World interface — so the whole simulation can be unit-tested head-less.
// The game package reads entity state to render it.
package entities

import "github.com/anpiex12/kebab-ali/internal/physics"

// World is the view of the level an entity needs in order to update: which
// tiles are solid, how big a tile is, and where the player currently is.
type World interface {
	Solid(tx, ty int) bool
	TileSize() int
	PlayerRect() physics.Rect
}

// Power is the player's döner-spit power tier. Collecting a spit upgrades it;
// taking a hit downgrades it one step, and being hit while Small is fatal.
type Power int

const (
	// Small is "Rookie Ali": one hit and a life is lost.
	Small Power = iota
	// Chef is the first spit tier: bigger, absorbs one extra hit.
	Chef
	// Master is the second tier: can throw spinning meat slices.
	Master
)

// Big reports whether the player is in an enlarged (Chef/Master) form.
func (p Power) Big() bool { return p >= Chef }

// CanThrow reports whether the player can throw meat-slice projectiles.
func (p Power) CanThrow() bool { return p == Master }

// Upgrade advances one tier, capping at Master.
func (p Power) Upgrade() Power {
	if p < Master {
		return p + 1
	}
	return Master
}

// Downgrade steps one tier down. The bool is true when the player was already
// Small, i.e. the hit is fatal.
func (p Power) Downgrade() (Power, bool) {
	if p == Small {
		return Small, true
	}
	return p - 1, false
}

// String returns a stable identifier used for HUD lookup keys.
func (p Power) String() string {
	switch p {
	case Chef:
		return "chef"
	case Master:
		return "master"
	default:
		return "small"
	}
}
