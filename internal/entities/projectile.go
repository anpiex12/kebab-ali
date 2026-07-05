package entities

import "github.com/anpiex12/kebab-ali/internal/physics"

// ProjectileKind distinguishes the player's meat slice from enemy chili shots.
type ProjectileKind int

const (
	// MeatSliceProj is thrown by Master Ali; it bounces and kills enemies.
	MeatSliceProj ProjectileKind = iota
	// ChiliProj is fired by peperoni enemies and hurts the player.
	ChiliProj
	// GarlicProj is Captain Garlic's floaty stink cloud.
	GarlicProj
	// RingProj is a thrown onion ring (arcs through the air).
	RingProj
)

// Projectile tuning.
const (
	meatSliceLife = 150
	meatGravity   = 0.22
	meatBounce    = -3.2
	chiliLife     = 170
)

// Projectile is a small moving hazard/weapon.
type Projectile struct {
	Kind     ProjectileKind
	Body     physics.Rect
	VX, VY   float64
	AnimTick int

	life  int
	alive bool
}

// NewMeatSlice spawns a spinning meat slice moving in dir (+1/-1) from (x,y).
func NewMeatSlice(x, y, dir float64) *Projectile {
	return &Projectile{
		Kind:  MeatSliceProj,
		Body:  physics.Rect{X: x, Y: y, W: 6, H: 6},
		VX:    3.2 * sign(dir),
		VY:    -1,
		life:  meatSliceLife,
		alive: true,
	}
}

// NewChili spawns a chili shot moving in dir from (x,y).
func NewChili(x, y, dir float64) *Projectile {
	return &Projectile{
		Kind:  ChiliProj,
		Body:  physics.Rect{X: x, Y: y, W: 6, H: 4},
		VX:    1.9 * sign(dir),
		life:  chiliLife,
		alive: true,
	}
}

// NewGarlicCloud spawns a slow, gently rising stink cloud toward dir.
func NewGarlicCloud(x, y, dir float64) *Projectile {
	return &Projectile{
		Kind:  GarlicProj,
		Body:  physics.Rect{X: x, Y: y, W: 10, H: 10},
		VX:    1.1 * sign(dir),
		VY:    -0.4,
		life:  200,
		alive: true,
	}
}

// NewOnionRing spawns an arcing onion ring toward dir.
func NewOnionRing(x, y, dir float64) *Projectile {
	return &Projectile{
		Kind:  RingProj,
		Body:  physics.Rect{X: x, Y: y, W: 10, H: 8},
		VX:    2.0 * sign(dir),
		VY:    -3.4,
		life:  200,
		alive: true,
	}
}

// arcs reports whether this projectile kind is affected by gravity.
func (pr *Projectile) arcs() bool {
	return pr.Kind == MeatSliceProj || pr.Kind == RingProj
}

// Alive reports whether the projectile is still active.
func (pr *Projectile) Alive() bool { return pr.alive }

// Rect returns the collision box.
func (pr *Projectile) Rect() physics.Rect { return pr.Body }

// Kill deactivates the projectile (on hitting a target).
func (pr *Projectile) Kill() { pr.alive = false }

// FromPlayer reports whether this projectile damages enemies (vs the player).
func (pr *Projectile) FromPlayer() bool { return pr.Kind == MeatSliceProj }

// Update advances the projectile, handling walls, ground bounce and lifetime.
func (pr *Projectile) Update(w World) {
	if !pr.alive {
		return
	}
	pr.AnimTick++
	pr.life--
	if pr.life <= 0 {
		pr.alive = false
		return
	}
	ts := w.TileSize()
	var hit bool
	pr.Body, hit = physics.ResolveX(pr.Body, pr.VX, ts, w.Solid)
	if hit {
		pr.alive = false // both kinds die against a wall
		return
	}
	switch {
	case pr.arcs():
		pr.VY += meatGravity
		var groundHit bool
		pr.Body, groundHit = physics.ResolveY(pr.Body, pr.VY, ts, w.Solid)
		if groundHit {
			if pr.Kind == MeatSliceProj && pr.VY > 0 {
				pr.VY = meatBounce // meat slices bounce
			} else {
				pr.alive = false // onion rings splat on the floor
			}
		}
	case pr.Kind == GarlicProj:
		// A floaty cloud that drifts and ignores terrain.
		pr.Body.Y += pr.VY
	}
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}
