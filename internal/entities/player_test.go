package entities

import (
	"testing"

	"github.com/anpiex12/kebab-ali/internal/physics"
)

// testWorld is a flat world with solid ground at the given tile row.
type testWorld struct {
	solids   map[[2]int]bool
	ts       int
	groundTy int
	player   physics.Rect
}

func flatWorld(groundRow int) *testWorld {
	return &testWorld{solids: map[[2]int]bool{}, ts: 16, groundTy: groundRow}
}

func (w *testWorld) Solid(tx, ty int) bool {
	if ty >= w.groundTy {
		return true // everything at/below the ground row is solid
	}
	return w.solids[[2]int{tx, ty}]
}
func (w *testWorld) TileSize() int             { return w.ts }
func (w *testWorld) PlayerRect() physics.Rect  { return w.player }

// settle runs empty-input frames until the player is grounded (or a cap).
func settle(p *Player, w World) {
	for i := 0; i < 240 && !p.OnGround(); i++ {
		p.Update(Input{}, w)
	}
}

func TestPlayerFallsAndLands(t *testing.T) {
	w := flatWorld(10) // ground at y = 160
	p := NewPlayer(Ali, 40, 40)
	settle(p, w)
	if !p.OnGround() {
		t.Fatal("player never landed")
	}
	if p.VY != 0 {
		t.Fatalf("landed player should have VY 0, got %v", p.VY)
	}
	// Feet should rest exactly on the ground surface (y = 160).
	if got := p.Body.Bottom(); got != 160 {
		t.Fatalf("feet at %v want 160", got)
	}
}

func TestPlayerJumpLeavesGround(t *testing.T) {
	w := flatWorld(10)
	p := NewPlayer(Ali, 40, 100)
	settle(p, w)
	p.Update(Input{JumpPress: true, JumpHeld: true}, w)
	if p.VY >= 0 {
		t.Fatalf("after jump VY should be negative, got %v", p.VY)
	}
	if p.OnGround() {
		t.Fatal("player should be airborne right after jumping")
	}
}

func TestVariableJumpHeightCut(t *testing.T) {
	w := flatWorld(10)
	p := NewPlayer(Ali, 40, 100)
	settle(p, w)
	p.Update(Input{JumpPress: true, JumpHeld: true}, w)
	// Release jump immediately: upward velocity should be clamped (cut).
	p.Update(Input{JumpHeld: false}, w)
	if p.VY < jumpCutSpeed-0.5 {
		t.Fatalf("early release should cut jump; VY=%v want >= %v", p.VY, jumpCutSpeed)
	}
}

func TestMehmetJumpsHigher(t *testing.T) {
	if statsFor(Mehmet).jump >= statsFor(Ali).jump {
		t.Error("Mehmet should have a stronger (more negative) jump than Ali")
	}
	if statsFor(Mehmet).runMax >= statsFor(Ali).runMax {
		t.Error("Mehmet should run slower than Ali")
	}
}

func TestPowerUpGrowsBodyKeepingFeet(t *testing.T) {
	w := flatWorld(10)
	p := NewPlayer(Ali, 40, 100)
	settle(p, w)
	feet := p.Body.Bottom()
	if !p.GivePower() {
		t.Fatal("first GivePower should upgrade")
	}
	if !p.Power.Big() {
		t.Fatal("player should be big after power-up")
	}
	if p.Body.Bottom() != feet {
		t.Fatalf("feet moved on grow: %v want %v", p.Body.Bottom(), feet)
	}
	if p.Body.H != bigH {
		t.Fatalf("body height not enlarged: %v", p.Body.H)
	}
}

func TestTakeHitDowngradesThenKills(t *testing.T) {
	w := flatWorld(10)
	p := NewPlayer(Ali, 40, 100)
	settle(p, w)
	p.GivePower() // -> Chef
	if lost := p.TakeHit(); lost {
		t.Fatal("hit as Chef should not cost a life")
	}
	if p.Power != Small {
		t.Fatalf("after hit power=%v want Small", p.Power)
	}
	// Immediately hitting again does nothing (invulnerable blink).
	if p.TakeHit() {
		t.Fatal("hit during invuln should be ignored")
	}
	// Clear invulnerability and hit while Small -> lose a life.
	p.hitInv = 0
	before := p.Lives
	if !p.TakeHit() {
		t.Fatal("hit as Small should cost a life")
	}
	if p.Lives != before-1 {
		t.Fatalf("lives=%d want %d", p.Lives, before-1)
	}
	if !p.Dead() {
		t.Fatal("player should be marked dead after losing a life")
	}
}

func TestEnrollFromMaster(t *testing.T) {
	p := NewPlayer(Ali, 0, 0)
	p.GivePower()
	p.GivePower() // Master
	if p.Power != Master {
		t.Fatalf("setup: power=%v want Master", p.Power)
	}
	p.Enroll()
	if p.Power != Chef {
		t.Fatalf("enroll from Master should drop to Chef, got %v", p.Power)
	}
}

func TestAyranMakesInvincible(t *testing.T) {
	p := NewPlayer(Ali, 0, 0)
	p.GiveAyran()
	if !p.Invincible() || !p.AyranActive() {
		t.Fatal("Ayran should grant invincibility")
	}
	if p.TakeHit() {
		t.Fatal("no damage should apply during Ayran")
	}
}

func TestCoinsGrantExtraLife(t *testing.T) {
	p := NewPlayer(Ali, 0, 0)
	before := p.Lives
	if oneUp := p.AddCoins(99); oneUp {
		t.Fatal("99 coins should not grant a life yet")
	}
	if oneUp := p.AddCoins(1); !oneUp {
		t.Fatal("100th coin should grant a life")
	}
	if p.Lives != before+1 {
		t.Fatalf("lives=%d want %d", p.Lives, before+1)
	}
	if p.Coins != 0 {
		t.Fatalf("coins should wrap to 0, got %d", p.Coins)
	}
}

func TestDieBypassesInvincibility(t *testing.T) {
	p := NewPlayer(Ali, 0, 0)
	p.GiveAyran() // invincible…
	before := p.Lives
	p.Die() // …but the sauce still kills
	if p.Lives != before-1 || !p.Dead() {
		t.Fatalf("Die should cost a life even during Ayran: lives=%d dead=%v", p.Lives, p.Dead())
	}
}
