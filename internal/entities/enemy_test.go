package entities

import (
	"testing"

	"github.com/anpiex12/kebab-ali/internal/physics"
)

func TestTomatoReversesAtWall(t *testing.T) {
	w := flatWorld(10)
	// Put a wall at tile column 3 (x = 48..64).
	w.solids[[2]int{3, 9}] = true
	tomato := NewTomato(70, 128) // walking left toward the wall
	if tomato.Facing() != -1 {
		t.Fatalf("tomato should start facing left")
	}
	for i := 0; i < 120; i++ {
		tomato.Update(w)
		if tomato.Facing() == 1 {
			return // reversed as expected
		}
	}
	t.Fatal("tomato never reversed at the wall")
}

func TestTomatoStaysOnLedge(t *testing.T) {
	// Ground only spans a few columns; the tomato must turn back at the edge
	// rather than walking into the void.
	w := &testWorld{solids: map[[2]int]bool{}, ts: 16, groundTy: 1000}
	for tx := 0; tx <= 5; tx++ {
		w.solids[[2]int{tx, 10}] = true
	}
	tomato := NewTomato(64, 144) // on the platform, facing left
	tomato.vx = 0.55
	tomato.facing = 1 // walk right toward the ledge at column 5/6
	turned := false
	for i := 0; i < 240; i++ {
		tomato.Update(w)
		if tomato.Facing() == -1 {
			turned = true
			break
		}
	}
	if !turned {
		t.Fatal("tomato walked off the ledge instead of turning")
	}
	if tomato.Rect().X > 6*16 {
		t.Fatalf("tomato left the platform: x=%v", tomato.Rect().X)
	}
}

func TestPeperoniNotStompableAndShoots(t *testing.T) {
	p := NewPeperoni(100, 144)
	if p.Stompable() {
		t.Error("peperoni must not be stompable")
	}
	w := &testWorld{
		solids:   map[[2]int]bool{},
		ts:       16,
		groundTy: 10,
		player:   physics.Rect{X: 40, Y: 140, W: 12, H: 14}, // in range, to the left
	}
	var shots []*Projectile
	for i := 0; i < 200 && len(shots) == 0; i++ {
		shots = append(shots, p.Update(w)...)
	}
	if len(shots) == 0 {
		t.Fatal("peperoni never fired a chili shot at an in-range player")
	}
	if shots[0].Kind != ChiliProj {
		t.Errorf("expected a chili projectile, got kind %v", shots[0].Kind)
	}
}

func TestStompKillsAfterAnimation(t *testing.T) {
	w := flatWorld(10)
	tomato := NewTomato(100, 128)
	tomato.Stomp()
	if !tomato.Dying() {
		t.Fatal("stomped tomato should enter dying state")
	}
	for i := 0; i < 40 && tomato.Alive(); i++ {
		tomato.Update(w)
	}
	if tomato.Alive() {
		t.Fatal("tomato should be removed after the squish animation")
	}
}

func TestMeatSliceExpires(t *testing.T) {
	w := flatWorld(100) // ground far below
	m := NewMeatSlice(50, 50, 1)
	for i := 0; i < meatSliceLife+5; i++ {
		m.Update(w)
	}
	if m.Alive() {
		t.Fatal("meat slice should expire after its lifetime")
	}
}

func TestMeatSliceDiesOnWall(t *testing.T) {
	w := &testWorld{solids: map[[2]int]bool{{4, 3}: true}, ts: 16, groundTy: 1000}
	m := NewMeatSlice(50, 50, 1) // flying right into the wall at column 4
	for i := 0; i < 60 && m.Alive(); i++ {
		m.Update(w)
	}
	if m.Alive() {
		t.Fatal("meat slice should die when it hits a wall")
	}
}

func TestCucumberFliesAndBobs(t *testing.T) {
	w := flatWorld(1000)
	c := NewCucumber(200, 100)
	startY := c.Rect().Y
	movedX := false
	bobbed := false
	for i := 0; i < 120; i++ {
		c.Update(w)
		if c.Rect().X != 200 {
			movedX = true
		}
		if c.Rect().Y != startY {
			bobbed = true
		}
	}
	if !movedX || !bobbed {
		t.Fatalf("cucumber should glide and bob: movedX=%v bobbed=%v", movedX, bobbed)
	}
}

func TestPowerUpItemFallsToGround(t *testing.T) {
	w := flatWorld(10) // ground at y=160
	spit := NewSpit(60, 100)
	for i := 0; i < 120; i++ {
		spit.Update(w)
	}
	if got := spit.Rect().Bottom(); got != 160 {
		t.Fatalf("spit should rest on the ground at 160, got %v", got)
	}
}

func TestCoinDoesNotMove(t *testing.T) {
	w := flatWorld(10)
	coin := NewTaler(60, 60)
	for i := 0; i < 30; i++ {
		coin.Update(w)
	}
	if coin.Rect().X != 60 || coin.Rect().Y != 60 {
		t.Fatal("coin should stay put")
	}
	if coin.AnimTick == 0 {
		t.Error("coin should still animate")
	}
}
