package entities

import "testing"

func TestGarlicTakesThreeStompsToDefeat(t *testing.T) {
	w := &testWorld{solids: map[[2]int]bool{}, ts: 16, groundTy: 13}
	w.player.X, w.player.Y, w.player.W, w.player.H = 40, 190, 12, 14
	b := NewGarlicBoss(200, 192)
	if b.HP() != 3 || b.MaxHP() != 3 {
		t.Fatalf("garlic should start at 3/3 HP, got %d/%d", b.HP(), b.MaxHP())
	}
	hits := 0
	for i := 0; i < 800 && b.Alive(); i++ {
		if b.Stompable() {
			before := b.HP()
			b.Stomp()
			if b.HP() < before {
				hits++
			}
		}
		b.Update(w)
	}
	if b.Alive() {
		t.Fatal("garlic should be defeated within the frame budget")
	}
	if hits != 3 {
		t.Fatalf("garlic should take exactly 3 landed stomps, took %d", hits)
	}
}

func TestGarlicShootsClouds(t *testing.T) {
	w := &testWorld{solids: map[[2]int]bool{}, ts: 16, groundTy: 13}
	w.player.X, w.player.Y, w.player.W, w.player.H = 40, 190, 12, 14
	b := NewGarlicBoss(200, 192)
	var got []*Projectile
	for i := 0; i < 200; i++ {
		got = append(got, b.Update(w)...)
	}
	found := false
	for _, p := range got {
		if p.Kind == GarlicProj {
			found = true
		}
	}
	if !found {
		t.Fatal("garlic never shot a stink cloud")
	}
}

func TestOnionTwinsBothNeededAndEnrage(t *testing.T) {
	twins := NewOnionTwins(200, 192)
	if len(twins) != 2 {
		t.Fatalf("expected 2 onion twins, got %d", len(twins))
	}
	a := twins[0].(*onionBoss)
	b := twins[1].(*onionBoss)
	// Defeat twin A via meat slices, spacing hits past the invuln window.
	w := &testWorld{solids: map[[2]int]bool{}, ts: 16, groundTy: 13}
	for i := 0; i < 400 && a.Alive(); i++ {
		a.HitByProjectile()
		for j := 0; j < 41; j++ {
			a.Update(w)
		}
	}
	if a.Alive() {
		t.Fatal("twin A should be defeated")
	}
	if !b.Alive() {
		t.Fatal("twin B should still be alive")
	}
	if !b.enraged {
		t.Fatal("surviving twin should be enraged after its sibling falls")
	}
}

func TestDurumPhasesAndVulnerability(t *testing.T) {
	d := NewDurumBoss(200, 192).(*durumBoss)
	if d.Phase() != 1 {
		t.Fatalf("full HP should be phase 1, got %d", d.Phase())
	}
	d.hp = 6
	if d.Phase() != 2 {
		t.Fatalf("hp 6 should be phase 2, got %d", d.Phase())
	}
	d.hp = 2
	if d.Phase() != 3 {
		t.Fatalf("hp 2 should be phase 3, got %d", d.Phase())
	}

	// Only vulnerable (stompable) while idle.
	d.state = durIdle
	if !d.Stompable() {
		t.Error("durum should be stompable while idle")
	}
	d.state = durRoll
	if d.Stompable() {
		t.Error("durum must not be stompable while rolling")
	}
	if d.Contact() != ContactHurt {
		t.Error("rolling durum body should hurt")
	}
	d.state = durLunge
	if d.Contact() != ContactEnroll {
		t.Error("lunging durum should enroll the player")
	}
}

func TestDurumCanBeDefeated(t *testing.T) {
	w := &testWorld{solids: map[[2]int]bool{}, ts: 16, groundTy: 13}
	w.player.X, w.player.Y, w.player.W, w.player.H = 40, 190, 12, 14
	d := NewDurumBoss(200, 192)
	for i := 0; i < 2000 && d.Alive(); i++ {
		if d.Stompable() {
			d.Stomp()
		}
		d.Update(w)
	}
	if d.Alive() {
		t.Fatal("durum should be defeatable within the frame budget")
	}
}

func TestNewBossesFactory(t *testing.T) {
	if got := NewBosses("garlic", 100, 100); len(got) != 1 {
		t.Errorf("garlic should yield 1 boss, got %d", len(got))
	}
	if got := NewBosses("onion", 100, 100); len(got) != 2 {
		t.Errorf("onion should yield 2 bosses, got %d", len(got))
	}
	if got := NewBosses("durum", 100, 100); len(got) != 1 {
		t.Errorf("durum should yield 1 boss, got %d", len(got))
	}
	if got := NewBosses("unknown", 100, 100); got != nil {
		t.Errorf("unknown boss kind should yield nil, got %v", got)
	}
}
