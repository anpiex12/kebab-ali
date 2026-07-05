package physics

import "testing"

func TestIntersects(t *testing.T) {
	a := Rect{X: 0, Y: 0, W: 10, H: 10}
	tests := []struct {
		name string
		b    Rect
		want bool
	}{
		{"overlap", Rect{5, 5, 10, 10}, true},
		{"contained", Rect{2, 2, 2, 2}, true},
		{"touching edge is not overlap", Rect{10, 0, 5, 5}, false},
		{"far away", Rect{100, 100, 5, 5}, false},
	}
	for _, tc := range tests {
		if got := a.Intersects(tc.b); got != tc.want {
			t.Errorf("%s: Intersects=%v want %v", tc.name, got, tc.want)
		}
	}
}

func TestRectEdges(t *testing.T) {
	r := Rect{X: 4, Y: 6, W: 10, H: 20}
	if r.Right() != 14 || r.Bottom() != 26 {
		t.Fatalf("edges: right=%v bottom=%v", r.Right(), r.Bottom())
	}
	if r.CenterX() != 9 || r.CenterY() != 16 {
		t.Fatalf("center: x=%v y=%v", r.CenterX(), r.CenterY())
	}
}

// solidGrid builds a SolidFunc from a set of solid tile coordinates.
func solidGrid(cells ...[2]int) SolidFunc {
	set := map[[2]int]bool{}
	for _, c := range cells {
		set[c] = true
	}
	return func(tx, ty int) bool { return set[[2]int{tx, ty}] }
}

func TestResolveXRightWall(t *testing.T) {
	// A 16px box at x=0 moving right into a wall at tile column 2 (x=32..48).
	solid := solidGrid([2]int{2, 0})
	r := Rect{X: 20, Y: 0, W: 16, H: 16}
	out, hit := ResolveX(r, 20, 16, solid)
	if !hit {
		t.Fatal("expected collision with right wall")
	}
	if out.Right() > 32 {
		t.Fatalf("box penetrated wall: right=%v want <=32", out.Right())
	}
	if out.X != 16 {
		t.Fatalf("box not snapped flush: x=%v want 16", out.X)
	}
}

func TestResolveXLeftWall(t *testing.T) {
	solid := solidGrid([2]int{0, 0})
	r := Rect{X: 20, Y: 0, W: 16, H: 16}
	out, hit := ResolveX(r, -20, 16, solid)
	if !hit {
		t.Fatal("expected collision with left wall")
	}
	if out.X < 16 {
		t.Fatalf("box penetrated wall: x=%v want >=16", out.X)
	}
}

func TestResolveXNoWall(t *testing.T) {
	solid := solidGrid()
	r := Rect{X: 0, Y: 0, W: 16, H: 16}
	out, hit := ResolveX(r, 10, 16, solid)
	if hit {
		t.Fatal("did not expect a collision")
	}
	if out.X != 10 {
		t.Fatalf("free movement wrong: x=%v want 10", out.X)
	}
}

func TestResolveYLanding(t *testing.T) {
	// Ground at row 5 (y=80..96). Box falling from y=40.
	solid := solidGrid([2]int{0, 5})
	r := Rect{X: 0, Y: 40, W: 16, H: 16}
	out, hit := ResolveY(r, 60, 16, solid)
	if !hit {
		t.Fatal("expected landing on ground")
	}
	if out.Bottom() != 80 {
		t.Fatalf("feet not on ground: bottom=%v want 80", out.Bottom())
	}
}

func TestResolveYCeiling(t *testing.T) {
	// Ceiling at row 0 (y=0..16). Box jumping up from y=40.
	solid := solidGrid([2]int{0, 0})
	r := Rect{X: 0, Y: 40, W: 16, H: 16}
	out, hit := ResolveY(r, -40, 16, solid)
	if !hit {
		t.Fatal("expected bonk on ceiling")
	}
	if out.Y != 16 {
		t.Fatalf("head not stopped at ceiling: y=%v want 16", out.Y)
	}
}

func TestStompedFrom(t *testing.T) {
	enemy := Rect{X: 100, Y: 100, W: 20, H: 20}
	// Player descending onto the enemy's head.
	player := Rect{X: 100, Y: 90, W: 16, H: 16}
	if !StompedFrom(player, enemy, 4) {
		t.Error("expected a stomp when landing on head while falling")
	}
	// Same overlap but moving up: not a stomp.
	if StompedFrom(player, enemy, -4) {
		t.Error("moving upward must not count as a stomp")
	}
	// Side bump: feet deep inside the enemy, not a stomp.
	side := Rect{X: 100, Y: 112, W: 16, H: 16}
	if StompedFrom(side, enemy, 4) {
		t.Error("deep side overlap must not count as a stomp")
	}
}

func TestClamp(t *testing.T) {
	if Clamp(5, 0, 10) != 5 || Clamp(-1, 0, 10) != 0 || Clamp(11, 0, 10) != 10 {
		t.Fatal("Clamp out of range")
	}
}
