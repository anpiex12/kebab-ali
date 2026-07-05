package entities

import "github.com/anpiex12/kebab-ali/internal/physics"

// ContactKind describes what happens when the player touches a boss's body
// (as opposed to stomping its head).
type ContactKind int

const (
	// ContactNone is harmless — the boss is stunned/vulnerable.
	ContactNone ContactKind = iota
	// ContactHurt deals normal damage (down a power tier / lose a life).
	ContactHurt
	// ContactEnroll is the Giant Dürüm's grab: wrap the player in fladenbread.
	ContactEnroll
)

// Boss is the behaviour every end-of-level boss implements. Like the rest of
// entities it is head-less; the game renders it from the exposed state.
type Boss interface {
	Update(w World) []*Projectile
	Rect() physics.Rect
	Alive() bool
	Stompable() bool
	Stomp()
	HitByProjectile()
	Contact() ContactKind
	HP() int
	MaxHP() int
	Phase() int
	Kind() string
	Facing() int
	AnimTick() int
	NameKey() string
	TauntKey() string
}

// bossBase carries state shared by all bosses.
type bossBase struct {
	body            physics.Rect
	vx, vy          float64
	facing          int
	hp, maxhp       int
	tick            int
	hitCD           int
	originX, rangeX float64
	kind            string
	nameKey         string
	tauntKey        string
}

func (b *bossBase) Rect() physics.Rect { return b.body }
func (b *bossBase) Alive() bool         { return b.hp > 0 }
func (b *bossBase) HP() int             { return b.hp }
func (b *bossBase) MaxHP() int          { return b.maxhp }
func (b *bossBase) Facing() int         { return b.facing }
func (b *bossBase) AnimTick() int       { return b.tick }
func (b *bossBase) Kind() string        { return b.kind }
func (b *bossBase) NameKey() string     { return b.nameKey }
func (b *bossBase) TauntKey() string    { return b.tauntKey }
func (b *bossBase) Phase() int          { return 1 }

// hurt applies one point of damage unless the boss is in its post-hit
// invulnerability window. It returns whether damage landed.
func (b *bossBase) hurt() bool {
	if b.hitCD > 0 || b.hp <= 0 {
		return false
	}
	b.hp--
	b.hitCD = 40
	return true
}

func (b *bossBase) tickTimers() {
	b.tick++
	if b.hitCD > 0 {
		b.hitCD--
	}
}

// gravityMove applies gravity and vertical collision, returning grounded state.
func (b *bossBase) gravityMove(w World) bool {
	b.vy += playerGravity
	if b.vy > playerMaxFall {
		b.vy = playerMaxFall
	}
	var hit bool
	b.body, hit = physics.ResolveY(b.body, b.vy, w.TileSize(), w.Solid)
	grounded := false
	if hit {
		if b.vy > 0 {
			grounded = true
		}
		b.vy = 0
	}
	return grounded
}

// moveX walks horizontally, turning at walls and at the arena patrol bounds.
func (b *bossBase) moveX(w World, vx float64) {
	b.vx = vx
	var hit bool
	b.body, hit = physics.ResolveX(b.body, b.vx, w.TileSize(), w.Solid)
	lo, hi := b.originX-b.rangeX, b.originX+b.rangeX
	if hit || b.body.X < lo || b.body.X > hi {
		b.facing = -b.facing
		b.body.X = physics.Clamp(b.body.X, lo, hi)
	}
}

func (b *bossBase) dirToPlayer(w World) int {
	if w.PlayerRect().CenterX() < b.body.CenterX() {
		return -1
	}
	return 1
}

// spawnY places a boss of height h so its feet rest on the floor tile just
// below the spawn tile at pixel-top y.
func spawnY(y, h float64) float64 { return y + float64(TileSizePx) - h }

// TileSizePx mirrors level.TileSize without importing the level package (which
// would create an import cycle, since level is data and entities is simulation).
const TileSizePx = 16

// --- Captain Garlic ---------------------------------------------------------

type garlicBoss struct {
	bossBase
	hopCD   int
	shootCD int
}

// NewGarlicBoss creates Captain Garlic (3 head-stomps to defeat).
func NewGarlicBoss(x, y float64) Boss {
	const w, h = 24.0, 24.0
	return &garlicBoss{bossBase: bossBase{
		body:     physics.Rect{X: x, Y: spawnY(y, h), W: w, H: h},
		facing:   -1,
		hp:       3,
		maxhp:    3,
		originX:  x,
		rangeX:   96,
		kind:     "garlic",
		nameKey:  "boss.garlic.name",
		tauntKey: "boss.garlic.taunt",
	}, shootCD: 90}
}

func (g *garlicBoss) Stompable() bool     { return true }
// Contact is harmless: Captain Garlic is defeated purely by stomping his head,
// and his danger is the stink clouds he lobs. Making the body itself hurt used
// to punish simply walking up to him (a grounded overlap at his base counted as
// a side hit), which felt like a bug.
func (g *garlicBoss) Contact() ContactKind { return ContactNone }
func (g *garlicBoss) Stomp()              { g.hurt() }
func (g *garlicBoss) HitByProjectile()    { g.hurt() }

func (g *garlicBoss) Update(w World) []*Projectile {
	g.tickTimers()
	grounded := g.gravityMove(w)
	// The angrier (lower HP) it gets, the faster it moves and shoots.
	speed := 0.7 + float64(g.maxhp-g.hp)*0.35
	g.moveX(w, speed*float64(g.facing))
	if grounded {
		if g.hopCD <= 0 {
			g.vy = -3.6
			g.hopCD = 70
		} else {
			g.hopCD--
		}
	}
	var shots []*Projectile
	g.shootCD--
	if g.shootCD <= 0 {
		g.shootCD = 80 - (g.maxhp-g.hp)*15
		dir := g.dirToPlayer(w)
		g.facing = dir
		shots = append(shots, NewGarlicCloud(g.body.CenterX(), g.body.Y, float64(dir)))
	}
	return shots
}

// --- Onion Twins ------------------------------------------------------------

type onionBoss struct {
	bossBase
	hopCD    int
	throwCD  int
	sibling  *onionBoss
	enraged  bool
}

// NewOnionTwins creates the two onion bosses; both must be defeated, and downing
// one enrages the survivor.
func NewOnionTwins(x, y float64) []Boss {
	a := newOnion(x-28, y)
	b := newOnion(x+28, y)
	a.sibling = b
	b.sibling = a
	return []Boss{a, b}
}

func newOnion(x, y float64) *onionBoss {
	const w, h = 22.0, 20.0
	return &onionBoss{bossBase: bossBase{
		body:     physics.Rect{X: x, Y: spawnY(y, h), W: w, H: h},
		facing:   -1,
		hp:       2,
		maxhp:    2,
		originX:  x,
		rangeX:   70,
		kind:     "onion",
		nameKey:  "boss.onion.name",
		tauntKey: "boss.onion.taunt",
	}, throwCD: 100, hopCD: 30}
}

func (o *onionBoss) Stompable() bool      { return true }
// Contact is harmless for the same reason as Captain Garlic: the Onion Twins'
// threat is the rings they throw, and they are beaten by stomping.
func (o *onionBoss) Contact() ContactKind { return ContactNone }
func (o *onionBoss) HitByProjectile()     { o.damage() }

func (o *onionBoss) Stomp() { o.damage() }

func (o *onionBoss) damage() {
	if o.hurt() && o.hp <= 0 && o.sibling != nil && o.sibling.Alive() {
		o.sibling.enrage() // downing one drives the other into a frenzy
	}
}

func (o *onionBoss) enrage() { o.enraged = true }

func (o *onionBoss) Update(w World) []*Projectile {
	o.tickTimers()
	grounded := o.gravityMove(w)
	hopEvery := 60
	throwEvery := 110
	speed := 0.9
	if o.enraged {
		hopEvery, throwEvery, speed = 34, 70, 1.6
	}
	if grounded {
		if o.hopCD <= 0 {
			o.vy = -4.0
			o.hopCD = hopEvery
			o.facing = o.dirToPlayer(w)
		} else {
			o.hopCD--
		}
	}
	o.moveX(w, speed*float64(o.facing))
	var shots []*Projectile
	o.throwCD--
	if o.throwCD <= 0 {
		o.throwCD = throwEvery
		shots = append(shots, NewOnionRing(o.body.CenterX(), o.body.CenterY(), float64(o.dirToPlayer(w))))
	}
	return shots
}

// --- Giant Dürüm (final boss, 3 phases) -------------------------------------

const (
	durIdle = iota
	durRoll
	durLunge
	durThrow
)

type durumBoss struct {
	bossBase
	state      int
	stateTimer int
	shootCD    int
}

// NewDurumBoss creates the Giant Dürüm final boss.
func NewDurumBoss(x, y float64) Boss {
	const w, h = 40.0, 30.0
	return &durumBoss{bossBase: bossBase{
		body:     physics.Rect{X: x - w/2, Y: spawnY(y, h), W: w, H: h},
		facing:   -1,
		hp:       9,
		maxhp:    9,
		originX:  x,
		rangeX:   120,
		kind:     "durum",
		nameKey:  "boss.durum.name",
		tauntKey: "boss.durum.taunt",
	}, state: durIdle, stateTimer: 60}
}

// Phase deepens as HP drops: 1 (roll only) -> 2 (+enroll lunge) -> 3 (+peperoni).
func (d *durumBoss) Phase() int {
	switch {
	case d.hp > 6:
		return 1
	case d.hp > 3:
		return 2
	default:
		return 3
	}
}

// Stompable only while idle/recovering — that's the window to hit its head.
func (d *durumBoss) Stompable() bool { return d.state == durIdle }

func (d *durumBoss) Contact() ContactKind {
	if d.hitCD > 0 {
		return ContactNone // harmless during the post-hit flash
	}
	switch d.state {
	case durLunge:
		return ContactEnroll
	case durRoll, durThrow:
		return ContactHurt
	default:
		return ContactNone
	}
}

func (d *durumBoss) Stomp() {
	if d.state == durIdle && d.hurt() {
		d.state = durIdle
		d.stateTimer = 50 // brief stagger before the next attack
	}
}

func (d *durumBoss) HitByProjectile() { d.hurt() }

func (d *durumBoss) Update(w World) []*Projectile {
	d.tickTimers()
	d.gravityMove(w)
	var shots []*Projectile

	switch d.state {
	case durIdle:
		d.moveX(w, 0.4*float64(d.facing)) // shuffle a little
		if d.stateTimer <= 0 {
			d.beginAction(w)
		}
	case durRoll:
		d.moveX(w, 2.6*float64(d.facing))
		if d.stateTimer <= 0 {
			d.recover()
		}
	case durLunge:
		d.moveX(w, 3.6*float64(d.facing))
		if d.stateTimer <= 0 {
			d.recover()
		}
	case durThrow:
		d.shootCD--
		if d.shootCD <= 0 {
			d.shootCD = 26
			dir := float64(d.dirToPlayer(w))
			shots = append(shots,
				NewChili(d.body.CenterX(), d.body.Y+6, dir),
			)
		}
		if d.stateTimer <= 0 {
			d.recover()
		}
	}
	d.stateTimer--
	return shots
}

func (d *durumBoss) recover() {
	d.state = durIdle
	d.stateTimer = 55
}

// beginAction picks the next attack based on the current phase.
func (d *durumBoss) beginAction(w World) {
	d.facing = d.dirToPlayer(w)
	switch d.Phase() {
	case 1:
		d.state = durRoll
		d.stateTimer = 70
	case 2:
		if d.tick%2 == 0 {
			d.state = durLunge
			d.stateTimer = 40
		} else {
			d.state = durRoll
			d.stateTimer = 70
		}
	default: // phase 3
		switch d.tick % 3 {
		case 0:
			d.state = durThrow
			d.stateTimer = 90
			d.shootCD = 10
		case 1:
			d.state = durLunge
			d.stateTimer = 40
		default:
			d.state = durRoll
			d.stateTimer = 60
		}
	}
}

// NewBosses builds the boss (or bosses) for a level's boss kind at (x,y).
func NewBosses(kind string, x, y float64) []Boss {
	switch kind {
	case "garlic":
		return []Boss{NewGarlicBoss(x, y)}
	case "onion":
		return NewOnionTwins(x, y)
	case "durum":
		return []Boss{NewDurumBoss(x, y)}
	default:
		return nil
	}
}
