package entities

import (
	"math"

	"github.com/anpiex12/kebab-ali/internal/physics"
)

// EnemyKind identifies an enemy type for rendering and spawn parsing.
type EnemyKind int

const (
	// KindTomato is a rolling tomato that patrols back and forth.
	KindTomato EnemyKind = iota
	// KindOnion is an onion ring that hops in arcs.
	KindOnion
	// KindPeperoni is a hot peperoni that shoots chili and cannot be stomped.
	KindPeperoni
	// KindCucumber is a cucumber glider that flies in a sine wave.
	KindCucumber
)

// Enemy is the behaviour every foe implements. Update returns any projectiles
// spawned this frame (usually none). All of it is head-less; the game renders
// enemies from their exposed state.
type Enemy interface {
	Update(w World) []*Projectile
	Rect() physics.Rect
	Alive() bool
	Dying() bool
	Stompable() bool
	Stomp()
	Damage()
	Kind() EnemyKind
	Facing() int
	AnimTick() int
}

// enemyBase carries the state and helpers shared by all enemies.
type enemyBase struct {
	body   physics.Rect
	vx, vy float64
	facing int
	alive  bool
	tick   int
	dying  int // >0 while playing the squish animation before removal
}

func (e *enemyBase) Rect() physics.Rect { return e.body }
func (e *enemyBase) Alive() bool         { return e.alive }
func (e *enemyBase) Dying() bool         { return e.dying > 0 }
func (e *enemyBase) Facing() int         { return e.facing }
func (e *enemyBase) AnimTick() int       { return e.tick }

func (e *enemyBase) kill() {
	if e.dying == 0 {
		e.dying = 16
		e.vx = 0
	}
}

// stepBase advances shared timers. It returns true when the concrete enemy
// should skip its AI this frame (because it is dying or already dead).
func (e *enemyBase) stepBase() bool {
	e.tick++
	if e.dying > 0 {
		e.dying--
		if e.dying == 0 {
			e.alive = false
		}
		return true
	}
	return !e.alive
}

// walkGravity applies gravity and vertical tile collision, returning whether
// the enemy is standing on ground.
func (e *enemyBase) walkGravity(w World) bool {
	e.vy += playerGravity
	if e.vy > playerMaxFall {
		e.vy = playerMaxFall
	}
	var hit bool
	e.body, hit = physics.ResolveY(e.body, e.vy, w.TileSize(), w.Solid)
	grounded := false
	if hit {
		if e.vy > 0 {
			grounded = true
		}
		e.vy = 0
	}
	return grounded
}

// patrolX moves horizontally, reversing on walls and (when grounded) at ledges
// so ground walkers stay on their platform.
func (e *enemyBase) patrolX(w World, grounded bool) {
	var hit bool
	e.body, hit = physics.ResolveX(e.body, e.vx, w.TileSize(), w.Solid)
	if hit {
		e.reverse()
		return
	}
	if grounded && e.ledgeAhead(w) {
		e.reverse()
	}
}

func (e *enemyBase) reverse() {
	e.facing = -e.facing
	e.vx = -e.vx
}

// ledgeAhead reports whether the tile just beyond and below the leading foot is
// empty (a drop-off).
func (e *enemyBase) ledgeAhead(w World) bool {
	ts := float64(w.TileSize())
	probeX := e.body.X - 1
	if e.facing > 0 {
		probeX = e.body.Right() + 1
	}
	tx := int(math.Floor(probeX / ts))
	ty := int(math.Floor((e.body.Bottom() + 2) / ts))
	return !w.Solid(tx, ty)
}

// --- Tomato -----------------------------------------------------------------

// Tomato is the basic patrolling enemy.
type Tomato struct{ enemyBase }

// NewTomato creates a tomato walking left from (x,y).
func NewTomato(x, y float64) *Tomato {
	return &Tomato{enemyBase{
		body:   physics.Rect{X: x, Y: y, W: 14, H: 14},
		vx:     -0.55,
		facing: -1,
		alive:  true,
	}}
}

func (t *Tomato) Kind() EnemyKind { return KindTomato }
func (t *Tomato) Stompable() bool { return true }
func (t *Tomato) Stomp()          { t.kill() }
func (t *Tomato) Damage()         { t.kill() }

func (t *Tomato) Update(w World) []*Projectile {
	if t.stepBase() {
		return nil
	}
	grounded := t.walkGravity(w)
	t.patrolX(w, grounded)
	return nil
}

// --- Onion ring -------------------------------------------------------------

// Onion hops toward wherever it is facing in gentle arcs.
type Onion struct {
	enemyBase
	hopTimer int
}

// NewOnion creates a hopping onion ring at (x,y).
func NewOnion(x, y float64) *Onion {
	return &Onion{enemyBase: enemyBase{
		body:   physics.Rect{X: x, Y: y, W: 14, H: 12},
		facing: -1,
		alive:  true,
	}, hopTimer: 40}
}

func (o *Onion) Kind() EnemyKind { return KindOnion }
func (o *Onion) Stompable() bool { return true }
func (o *Onion) Stomp()          { o.kill() }
func (o *Onion) Damage()         { o.kill() }

func (o *Onion) Update(w World) []*Projectile {
	if o.stepBase() {
		return nil
	}
	grounded := o.walkGravity(w)
	if grounded {
		if o.hopTimer <= 0 {
			o.vy = -4.2
			o.vx = 0.9 * float64(o.facing)
			o.hopTimer = 46
		} else {
			o.hopTimer--
			o.vx = 0
		}
	}
	var hit bool
	o.body, hit = physics.ResolveX(o.body, o.vx, w.TileSize(), w.Solid)
	if hit {
		o.reverse()
	}
	return nil
}

// --- Peperoni (shooter) -----------------------------------------------------

// Peperoni stands its ground, faces the player and fires chili shots. It cannot
// be stomped — only a meat slice defeats it.
type Peperoni struct {
	enemyBase
	shootTimer int
}

// NewPeperoni creates a chili-shooting peperoni at (x,y).
func NewPeperoni(x, y float64) *Peperoni {
	return &Peperoni{enemyBase: enemyBase{
		body:   physics.Rect{X: x, Y: y, W: 12, H: 16},
		facing: -1,
		alive:  true,
	}, shootTimer: 90}
}

func (p *Peperoni) Kind() EnemyKind { return KindPeperoni }
func (p *Peperoni) Stompable() bool { return false }
func (p *Peperoni) Stomp()          {} // stomping does nothing; it's spiky
func (p *Peperoni) Damage()         { p.kill() }

func (p *Peperoni) Update(w World) []*Projectile {
	if p.stepBase() {
		return nil
	}
	p.walkGravity(w)
	// Face the player.
	pr := w.PlayerRect()
	if pr.CenterX() < p.body.CenterX() {
		p.facing = -1
	} else {
		p.facing = 1
	}
	p.shootTimer--
	if p.shootTimer <= 0 {
		p.shootTimer = 120
		// Only shoot if the player is roughly on the same level and in range.
		if math.Abs(pr.CenterY()-p.body.CenterY()) < 40 && math.Abs(pr.CenterX()-p.body.CenterX()) < 160 {
			shot := NewChili(p.body.CenterX(), p.body.CenterY(), float64(p.facing))
			return []*Projectile{shot}
		}
	}
	return nil
}

// --- Cucumber glider --------------------------------------------------------

// Cucumber flies in a horizontal patrol with a vertical sine bob, ignoring
// gravity and terrain.
type Cucumber struct {
	enemyBase
	originX, originY float64
	rangeX           float64
	phase            float64
}

// NewCucumber creates a gliding cucumber centred at (x,y).
func NewCucumber(x, y float64) *Cucumber {
	return &Cucumber{
		enemyBase: enemyBase{
			body:   physics.Rect{X: x, Y: y, W: 16, H: 12},
			vx:     -0.8,
			facing: -1,
			alive:  true,
		},
		originX: x,
		originY: y,
		rangeX:  48,
	}
}

func (c *Cucumber) Kind() EnemyKind { return KindCucumber }
func (c *Cucumber) Stompable() bool { return true }
func (c *Cucumber) Stomp()          { c.kill() }
func (c *Cucumber) Damage()         { c.kill() }

func (c *Cucumber) Update(w World) []*Projectile {
	if c.stepBase() {
		return nil
	}
	c.body.X += c.vx
	if c.body.X < c.originX-c.rangeX || c.body.X > c.originX+c.rangeX {
		c.reverse()
	}
	c.phase += 0.12
	c.body.Y = c.originY + math.Sin(c.phase)*6
	return nil
}
