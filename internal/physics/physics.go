// Package physics implements the small amount of geometry the game needs:
// axis-aligned bounding boxes (AABB) and tile-based collision resolution.
//
// It is deliberately free of any Ebitengine or rendering dependency so that
// the whole thing can be unit-tested head-less (no window required).
package physics

import "math"

// epsilon nudges the leading edge inward so a box sitting exactly on a tile
// boundary is not counted as being inside the next tile.
const epsilon = 1e-6

// Rect is an axis-aligned bounding box. X/Y is the top-left corner, with the
// Y axis pointing down (screen coordinates), and W/H are width/height.
type Rect struct {
	X, Y, W, H float64
}

// Right returns the X coordinate of the right edge.
func (r Rect) Right() float64 { return r.X + r.W }

// Bottom returns the Y coordinate of the bottom edge.
func (r Rect) Bottom() float64 { return r.Y + r.H }

// CenterX returns the horizontal center.
func (r Rect) CenterX() float64 { return r.X + r.W/2 }

// CenterY returns the vertical center.
func (r Rect) CenterY() float64 { return r.Y + r.H/2 }

// Intersects reports whether two rectangles overlap. Touching edges do not
// count as an intersection.
func (r Rect) Intersects(o Rect) bool {
	return r.X < o.Right() && r.Right() > o.X &&
		r.Y < o.Bottom() && r.Bottom() > o.Y
}

// ContainsPoint reports whether the point (px,py) lies inside the rectangle.
func (r Rect) ContainsPoint(px, py float64) bool {
	return px >= r.X && px < r.Right() && py >= r.Y && py < r.Bottom()
}

// SolidFunc reports whether the tile at grid coordinates (tx,ty) blocks
// movement. Coordinates outside the map should generally be reported as solid
// (walls) or non-solid (open sky) depending on the caller's preference.
type SolidFunc func(tx, ty int) bool

// ResolveX moves r horizontally by dx and pushes it back out of any solid tile
// it would penetrate. It returns the corrected rectangle and whether a
// collision happened. The caller can infer the direction from the sign of dx.
//
// The columns swept between the old and new leading edge are all tested, so a
// fast-moving box cannot tunnel straight through a thin wall.
func ResolveX(r Rect, dx float64, tileSize int, solid SolidFunc) (Rect, bool) {
	if dx == 0 || tileSize <= 0 {
		r.X += dx
		return r, false
	}
	ts := float64(tileSize)
	oldLeft, oldRight := r.X, r.Right()
	r.X += dx
	top := int(math.Floor(r.Y / ts))
	bottom := int(math.Floor((r.Bottom() - epsilon) / ts))
	if dx > 0 {
		startCol := int(math.Floor((oldRight - epsilon) / ts))
		endCol := int(math.Floor((r.Right() - epsilon) / ts))
		for col := startCol; col <= endCol; col++ {
			if columnBlocked(col, top, bottom, solid) {
				r.X = float64(col)*ts - r.W
				return r, true
			}
		}
	} else {
		startCol := int(math.Floor(oldLeft / ts))
		endCol := int(math.Floor(r.X / ts))
		for col := startCol; col >= endCol; col-- {
			if columnBlocked(col, top, bottom, solid) {
				r.X = float64(col+1) * ts
				return r, true
			}
		}
	}
	return r, false
}

// ResolveY moves r vertically by dy and pushes it back out of any solid tile.
// It returns the corrected rectangle and whether a collision happened. A
// collision while moving down (dy>0) means the box landed on ground; while
// moving up (dy<0) it means the box bonked its head on a ceiling. Rows swept in
// between are all tested so the box cannot tunnel through the floor at speed.
func ResolveY(r Rect, dy float64, tileSize int, solid SolidFunc) (Rect, bool) {
	if dy == 0 || tileSize <= 0 {
		r.Y += dy
		return r, false
	}
	ts := float64(tileSize)
	oldTop, oldBottom := r.Y, r.Bottom()
	r.Y += dy
	left := int(math.Floor(r.X / ts))
	right := int(math.Floor((r.Right() - epsilon) / ts))
	if dy > 0 {
		startRow := int(math.Floor((oldBottom - epsilon) / ts))
		endRow := int(math.Floor((r.Bottom() - epsilon) / ts))
		for row := startRow; row <= endRow; row++ {
			if rowBlocked(row, left, right, solid) {
				r.Y = float64(row)*ts - r.H
				return r, true
			}
		}
	} else {
		startRow := int(math.Floor(oldTop / ts))
		endRow := int(math.Floor(r.Y / ts))
		for row := startRow; row >= endRow; row-- {
			if rowBlocked(row, left, right, solid) {
				r.Y = float64(row+1) * ts
				return r, true
			}
		}
	}
	return r, false
}

func columnBlocked(col, top, bottom int, solid SolidFunc) bool {
	for ty := top; ty <= bottom; ty++ {
		if solid(col, ty) {
			return true
		}
	}
	return false
}

func rowBlocked(row, left, right int, solid SolidFunc) bool {
	for tx := left; tx <= right; tx++ {
		if solid(tx, row) {
			return true
		}
	}
	return false
}

// Clamp constrains v to the inclusive range [lo, hi].
func Clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// StompedFrom reports whether attacker (typically the player) is landing on top
// of victim (an enemy) this frame: it must be moving downward, its feet must be
// in the upper band of the victim, and the two boxes must overlap horizontally.
// This is the classic "jump on the enemy's head" test.
func StompedFrom(attacker, victim Rect, attackerVelY float64) bool {
	if attackerVelY <= 0 {
		return false
	}
	if !attacker.Intersects(victim) {
		return false
	}
	// Feet must be in the top third of the victim to count as a stomp rather
	// than a side bump.
	feet := attacker.Bottom()
	return feet <= victim.Y+victim.H*0.6 &&
		attacker.Right() > victim.X && attacker.X < victim.Right()
}
