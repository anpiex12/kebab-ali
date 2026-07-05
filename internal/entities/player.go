package entities

import (
	"math"

	"github.com/anpiex12/kebab-ali/internal/physics"
)

// CharKind selects which brother is being played, which only changes a few
// movement stats and (in the renderer) the apron colour.
type CharKind int

const (
	// Ali is the nimble default: balanced jump and speed.
	Ali CharKind = iota
	// Mehmet jumps higher but runs a touch slower.
	Mehmet
)

// Player tuning constants, in low-resolution pixels per 1/60s frame.
const (
	playerGravity  = 0.42
	playerMaxFall  = 7.5
	groundAccel    = 0.34
	airAccel       = 0.22
	groundFriction = 0.55
	jumpCutSpeed   = -2.6 // upward speed clamped to this when jump is released early
	stompBounce    = -5.0

	coyoteFrames     = 6
	jumpBufferFrames = 6
	hitInvulnFrames  = 90
	throwCooldown    = 22
	powerFlashFrames = 24

	// AyranFrames is 10 seconds of invincibility at 60fps.
	AyranFrames = 600
	// CoinsPerLife is how many taler earn an extra life.
	CoinsPerLife = 100
	// StartingLives is the number of lives a new game begins with.
	StartingLives = 3
)

// body sizes for the small and big forms.
const (
	smallW, smallH = 11, 14
	bigW, bigH     = 12, 22
)

type charStats struct {
	walkMax, runMax, jump float64
}

func statsFor(k CharKind) charStats {
	if k == Mehmet {
		return charStats{walkMax: 1.6, runMax: 2.6, jump: -8.1}
	}
	return charStats{walkMax: 1.8, runMax: 2.9, jump: -7.4}
}

// Input is the per-frame control state handed to the player; it is deliberately
// device-agnostic so the game maps keys onto it and tests drive it directly.
type Input struct {
	Left, Right, Run    bool
	JumpHeld, JumpPress bool
	Throw               bool
}

// Events reports notable things that happened during an Update so the game can
// react (play a sound, spawn a projectile, pop a "?" block).
type Events struct {
	Jumped bool
	Landed bool
	Threw  bool
	// Bonked is set when the player's head hit a solid tile from below; BonkTX
	// and BonkTY are that tile's grid coordinates (used to pop "?"/bread blocks).
	Bonked         bool
	BonkTX, BonkTY int
}

// Player is the controllable character.
type Player struct {
	Kind   CharKind
	Body   physics.Rect
	VX, VY float64
	Facing int // +1 right, -1 left
	Power  Power

	Lives int
	Coins int
	Score int

	stats    charStats
	onGround bool
	coyote   int
	jumpBuf  int
	hitInv   int
	ayran    int
	throwCD  int
	flash    int
	dead     bool

	// AnimTick advances every frame and drives the walk-cycle in the renderer.
	AnimTick int
}

// NewPlayer creates a fresh player of the given character at the top-left
// position (x,y), in Small form with the starting life count.
func NewPlayer(k CharKind, x, y float64) *Player {
	p := &Player{
		Kind:   k,
		Body:   physics.Rect{X: x, Y: y, W: smallW, H: smallH},
		Facing: 1,
		Power:  Small,
		Lives:  StartingLives,
		stats:  statsFor(k),
	}
	return p
}

// PlaceAt moves the player to (x,y) as its top-left, zeroes velocity and grants
// a brief spawn invulnerability. Used at level start and on respawn.
func (p *Player) PlaceAt(x, y float64) {
	p.Body.X, p.Body.Y = x, y
	p.VX, p.VY = 0, 0
	p.onGround = false
	p.dead = false
	p.hitInv = 60
}

// Rect returns the player's collision box (implements a common shape used by
// the game for overlap tests).
func (p *Player) Rect() physics.Rect { return p.Body }

// Dead reports whether the player lost a life this frame and awaits respawn.
func (p *Player) Dead() bool { return p.dead }

// OnGround reports whether the player is standing on solid ground.
func (p *Player) OnGround() bool { return p.onGround }

// Invincible reports whether the player currently ignores enemy damage (from
// the Ayran power or post-hit blinking).
func (p *Player) Invincible() bool { return p.ayran > 0 || p.hitInv > 0 }

// AyranActive reports whether the Ayran boost (speed + invincibility + flicker)
// is running.
func (p *Player) AyranActive() bool { return p.ayran > 0 }

// Flashing reports whether the player just powered up (for a brief sparkle).
func (p *Player) Flashing() bool { return p.flash > 0 }

// Update advances the player one frame given input and the surrounding world,
// returning notable events.
func (p *Player) Update(in Input, w World) Events {
	var ev Events
	ts := w.TileSize()
	solid := w.Solid

	// --- horizontal movement -------------------------------------------------
	maxSpeed := p.stats.walkMax
	if in.Run {
		maxSpeed = p.stats.runMax
	}
	if p.ayran > 0 {
		maxSpeed *= 1.3 // Ayran tempo boost
	}
	accel := groundAccel
	if !p.onGround {
		accel = airAccel
	}
	switch {
	case in.Left && !in.Right:
		p.VX -= accel
		p.Facing = -1
	case in.Right && !in.Left:
		p.VX += accel
		p.Facing = 1
	default:
		// friction toward zero
		if p.onGround {
			if p.VX > groundFriction {
				p.VX -= groundFriction
			} else if p.VX < -groundFriction {
				p.VX += groundFriction
			} else {
				p.VX = 0
			}
		}
	}
	p.VX = physics.Clamp(p.VX, -maxSpeed, maxSpeed)

	// --- jump (coyote time + input buffering) --------------------------------
	if p.onGround {
		p.coyote = coyoteFrames
	} else if p.coyote > 0 {
		p.coyote--
	}
	if in.JumpPress {
		p.jumpBuf = jumpBufferFrames
	} else if p.jumpBuf > 0 {
		p.jumpBuf--
	}
	if p.jumpBuf > 0 && p.coyote > 0 {
		p.VY = p.stats.jump
		p.onGround = false
		p.coyote = 0
		p.jumpBuf = 0
		ev.Jumped = true
	}
	// Variable jump height: releasing jump early cuts the upward velocity.
	if !in.JumpHeld && p.VY < jumpCutSpeed {
		p.VY = jumpCutSpeed
	}

	// --- gravity -------------------------------------------------------------
	p.VY += playerGravity
	if p.VY > playerMaxFall {
		p.VY = playerMaxFall
	}

	// --- collision resolution ------------------------------------------------
	var hit bool
	p.Body, hit = physics.ResolveX(p.Body, p.VX, ts, solid)
	if hit {
		p.VX = 0
	}
	wasAir := !p.onGround
	p.Body, hit = physics.ResolveY(p.Body, p.VY, ts, solid)
	if hit {
		if p.VY > 0 {
			if wasAir {
				ev.Landed = true
			}
			p.onGround = true
		} else if p.VY < 0 {
			// Head bonk: report the tile above the head centre so the game can
			// pop a "?" block or shatter a bread block.
			tsf := float64(ts)
			ev.Bonked = true
			ev.BonkTX = int(math.Floor(p.Body.CenterX() / tsf))
			ev.BonkTY = int(math.Floor((p.Body.Y - 1) / tsf))
		}
		p.VY = 0
	} else {
		p.onGround = false
	}

	// --- throwing ------------------------------------------------------------
	if in.Throw && p.Power.CanThrow() && p.throwCD == 0 {
		p.throwCD = throwCooldown
		ev.Threw = true
	}

	// --- timers --------------------------------------------------------------
	if p.hitInv > 0 {
		p.hitInv--
	}
	if p.ayran > 0 {
		p.ayran--
	}
	if p.throwCD > 0 {
		p.throwCD--
	}
	if p.flash > 0 {
		p.flash--
	}
	p.AnimTick++
	return ev
}

// applyPower switches to np, resizing the body while keeping the feet planted.
func (p *Player) applyPower(np Power) {
	wasBig, nowBig := p.Power.Big(), np.Big()
	p.Power = np
	if wasBig == nowBig {
		return
	}
	bottom := p.Body.Bottom()
	if nowBig {
		p.Body.W, p.Body.H = bigW, bigH
	} else {
		p.Body.W, p.Body.H = smallW, smallH
	}
	p.Body.Y = bottom - p.Body.H
}

// GivePower upgrades the power tier, returning whether an upgrade actually
// happened (false if already at Master).
func (p *Player) GivePower() bool {
	np := p.Power.Upgrade()
	if np == p.Power {
		return false
	}
	p.applyPower(np)
	p.flash = powerFlashFrames
	return true
}

// GiveAyran starts the Ayran invincibility + speed boost.
func (p *Player) GiveAyran() { p.ayran = AyranFrames }

// AddCoins adds taler, converting every CoinsPerLife into an extra life. It
// returns true if at least one life was earned.
func (p *Player) AddCoins(n int) (oneUp bool) {
	p.Coins += n
	for p.Coins >= CoinsPerLife {
		p.Coins -= CoinsPerLife
		p.Lives++
		oneUp = true
	}
	return oneUp
}

// AddScore adds to the score.
func (p *Player) AddScore(n int) { p.Score += n }

// Bounce gives the little upward hop after stomping an enemy.
func (p *Player) Bounce() {
	p.VY = stompBounce
	p.onGround = false
}

// TakeHit applies enemy contact damage: downgrade one tier, or lose a life if
// already Small. No effect while invincible. Returns true if a life was lost.
func (p *Player) TakeHit() (lostLife bool) {
	if p.dead || p.Invincible() {
		return false
	}
	np, died := p.Power.Downgrade()
	if died {
		p.loseLife()
		return true
	}
	p.applyPower(np)
	p.hitInv = hitInvulnFrames
	return false
}

// Enroll is the Giant Dürüm's grab: it strips a power tier (or a life if Small),
// ignoring the post-hit blink but still stopped by Ayran invincibility.
func (p *Player) Enroll() (lostLife bool) {
	if p.dead || p.ayran > 0 {
		return false
	}
	np, died := p.Power.Downgrade()
	if died {
		p.loseLife()
		return true
	}
	p.applyPower(np)
	p.hitInv = hitInvulnFrames
	return true
}

// Die is an instant kill that bypasses power tiers and invincibility, used for
// the deadly sauce lakes and bottomless gaps.
func (p *Player) Die() {
	if !p.dead {
		p.loseLife()
	}
}

func (p *Player) loseLife() {
	p.Lives--
	p.Power = Small
	bottom := p.Body.Bottom()
	p.Body.W, p.Body.H = smallW, smallH
	p.Body.Y = bottom - smallH
	p.dead = true
	p.VX, p.VY = 0, 0
}
