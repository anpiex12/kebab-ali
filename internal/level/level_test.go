package level

import "testing"

func TestBuildParsesTilesAndSpawns(t *testing.T) {
	def := Def{
		NameKey: "test",
		Rows: []string{
			"      ",
			" P  o ",
			"?  ~  ",
			"######",
		},
		Signs: map[rune]string{},
	}
	l := Build(def)
	if l.W != 6 || l.H != 4 {
		t.Fatalf("dimensions W=%d H=%d want 6x4", l.W, l.H)
	}
	// Player start at column 1, row 1.
	if l.PlayerStart.X != 1*TileSize || l.PlayerStart.Y != 1*TileSize {
		t.Fatalf("player start = %+v", l.PlayerStart)
	}
	// The '?' becomes a coin box.
	if l.TileAt(0, 2) != BoxCoin {
		t.Errorf("expected BoxCoin at (0,2), got %v", l.TileAt(0, 2))
	}
	// The '~' becomes sauce.
	if l.TileAt(3, 2) != Sauce {
		t.Errorf("expected Sauce at (3,2), got %v", l.TileAt(3, 2))
	}
	// The tomato became an enemy spawn.
	if len(l.Enemies) != 1 || l.Enemies[0].Kind != "tomato" {
		t.Fatalf("expected one tomato spawn, got %+v", l.Enemies)
	}
	// The bottom row is solid.
	if !l.Solid(0, 3) {
		t.Error("bottom row should be solid")
	}
}

func TestSolidBoundaries(t *testing.T) {
	l := Build(Def{Rows: []string{"  ", "  "}})
	if !l.Solid(-1, 0) || !l.Solid(2, 0) {
		t.Error("left/right of map should be solid walls")
	}
	if l.Solid(0, -1) {
		t.Error("above the map should be open sky")
	}
	if l.Solid(0, 5) {
		t.Error("below the map should be an open pit (non-solid)")
	}
}

func TestSauceOverlap(t *testing.T) {
	l := Build(Def{Rows: []string{
		"  ~ ",
		"####",
	}})
	// Sauce tile is at column 2, row 0 => pixel x 32..48, y 0..16.
	if !l.SauceOverlap(34, 4, 8, 8) {
		t.Error("rectangle inside the sauce tile should overlap")
	}
	if l.SauceOverlap(0, 0, 8, 8) {
		t.Error("rectangle away from sauce should not overlap")
	}
}

func TestSetTileBreaksBlock(t *testing.T) {
	l := Build(Def{Rows: []string{"B"}})
	if l.TileAt(0, 0) != Bread {
		t.Fatal("expected a bread block")
	}
	l.SetTile(0, 0, Empty)
	if l.TileAt(0, 0) != Empty || l.Solid(0, 0) {
		t.Error("broken block should be empty and non-solid")
	}
}

func TestAllThreeLevelsAreWellFormed(t *testing.T) {
	levels := All()
	if len(levels) != 3 {
		t.Fatalf("expected 3 levels, got %d", len(levels))
	}
	wantBoss := []string{"garlic", "onion", "durum"}
	for i, l := range levels {
		if l.BossKind != wantBoss[i] {
			t.Errorf("level %d boss = %q want %q", i+1, l.BossKind, wantBoss[i])
		}
		// Every level must define a player start and a boss spawn.
		if l.PlayerStart == (Point{}) {
			t.Errorf("level %d has no player start", i+1)
		}
		if l.Boss == (Point{}) {
			t.Errorf("level %d has no boss spawn", i+1)
		}
		// The player must start standing on solid ground somewhere below.
		psx := int(l.PlayerStart.X / TileSize)
		grounded := false
		for ty := int(l.PlayerStart.Y/TileSize) + 1; ty < l.H; ty++ {
			if l.Solid(psx, ty) {
				grounded = true
				break
			}
		}
		if !grounded {
			t.Errorf("level %d player start is over a bottomless void", i+1)
		}
		// Sanity: each level should have at least a few enemies and coins.
		if len(l.Enemies) < 3 {
			t.Errorf("level %d has too few enemies: %d", i+1, len(l.Enemies))
		}
		if len(l.Items) < 2 {
			t.Errorf("level %d has too few coins/items: %d", i+1, len(l.Items))
		}
	}
}

func TestBossSpawnReachableOnGround(t *testing.T) {
	// Regression guard: the boss spawn column must have solid ground under it so
	// the boss and player can actually stand in the arena.
	for i, l := range All() {
		bx := int(l.Boss.X / TileSize)
		if !l.Solid(bx, l.H-1) {
			t.Errorf("level %d boss arena has no floor at column %d", i+1, bx)
		}
	}
}
