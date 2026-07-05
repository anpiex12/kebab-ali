package entities

import (
	"testing"

	"github.com/anpiex12/kebab-ali/internal/physics"
)

// TestGarlicApproachNeverHurtsButCanBeStomped reproduces the reported bug — the
// player kept losing power/lives while trying to jump on Captain Garlic — and
// locks in the fix. It drives the player through the exact collision logic the
// play scene uses (stomp / ayran / contact), chasing and hopping onto the boss,
// and asserts the player takes zero body damage while the stomps still land.
func TestGarlicApproachNeverHurtsButCanBeStomped(t *testing.T) {
	w := &testWorld{solids: map[[2]int]bool{}, ts: 16, groundTy: 13}
	boss := NewGarlicBoss(200, 192)
	p := NewPlayer(Ali, 120, 192)
	p.hitInv = 0
	w.player = p.Body
	for i := 0; i < 20; i++ { // land on the ground first
		p.Update(Input{}, w)
	}

	hits, stomps := 0, 0
	for i := 0; i < 400 && boss.Alive(); i++ {
		in := Input{}
		switch {
		case boss.Rect().CenterX() > p.Body.CenterX()+2:
			in.Right = true
		case boss.Rect().CenterX() < p.Body.CenterX()-2:
			in.Left = true
		}
		if p.OnGround() {
			in.JumpPress, in.JumpHeld = true, true
		}
		p.Update(in, w)
		w.player = p.Body
		boss.Update(w)

		if boss.Alive() && p.Body.Intersects(boss.Rect()) {
			switch {
			case boss.Stompable() && physics.StompedFrom(p.Body, boss.Rect(), p.VY):
				boss.Stomp()
				p.Bounce()
				stomps++
			case p.AyranActive():
				boss.Stomp()
				p.Bounce()
			default:
				if boss.Contact() == ContactHurt && !p.Invincible() {
					p.TakeHit()
					hits++
				}
			}
		}
	}

	if hits != 0 {
		t.Fatalf("player took %d body hits from Captain Garlic while stomping (want 0)", hits)
	}
	if stomps == 0 {
		t.Fatal("player never managed to stomp the boss")
	}
	if boss.Alive() {
		t.Fatalf("boss should have been defeated by stomps (HP left %d)", boss.HP())
	}
}
