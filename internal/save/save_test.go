package save

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLeaderboardAddRanksByScore(t *testing.T) {
	var lb Leaderboard
	lb.Add(Entry{Name: "A", Score: 100, Seconds: 60})
	lb.Add(Entry{Name: "B", Score: 300, Seconds: 60})
	rank := lb.Add(Entry{Name: "C", Score: 200, Seconds: 60})
	if rank != 2 {
		t.Fatalf("C should rank 2nd, got %d", rank)
	}
	want := []string{"B", "C", "A"}
	for i, e := range lb.Entries {
		if e.Name != want[i] {
			t.Fatalf("position %d: got %s want %s", i, e.Name, want[i])
		}
	}
}

func TestLeaderboardTieBreakByTime(t *testing.T) {
	var lb Leaderboard
	lb.Add(Entry{Name: "Slow", Score: 500, Seconds: 120})
	rank := lb.Add(Entry{Name: "Fast", Score: 500, Seconds: 90})
	if rank != 1 {
		t.Fatalf("faster run with equal score should rank 1st, got %d", rank)
	}
	if lb.Entries[0].Name != "Fast" {
		t.Fatalf("Fast should be first, got %s", lb.Entries[0].Name)
	}
}

func TestLeaderboardTrimsToMax(t *testing.T) {
	var lb Leaderboard
	for i := 0; i < MaxEntries+5; i++ {
		lb.Add(Entry{Name: "P", Score: i * 10, Seconds: 60})
	}
	if len(lb.Entries) != MaxEntries {
		t.Fatalf("board should hold %d, got %d", MaxEntries, len(lb.Entries))
	}
	// The lowest scores must have been dropped: min kept score is 50.
	lowest := lb.Entries[len(lb.Entries)-1].Score
	if lowest != 50 {
		t.Fatalf("lowest kept score should be 50, got %d", lowest)
	}
}

func TestLeaderboardAddReturnsZeroWhenNotQualified(t *testing.T) {
	var lb Leaderboard
	for i := 0; i < MaxEntries; i++ {
		lb.Add(Entry{Name: "P", Score: 1000, Seconds: 60})
	}
	rank := lb.Add(Entry{Name: "Loser", Score: 1, Seconds: 60})
	if rank != 0 {
		t.Fatalf("non-qualifying entry should return rank 0, got %d", rank)
	}
	if len(lb.Entries) != MaxEntries {
		t.Fatalf("board grew past max: %d", len(lb.Entries))
	}
}

func TestQualifies(t *testing.T) {
	var lb Leaderboard
	if !lb.Qualifies(0) {
		t.Error("empty board should qualify any score")
	}
	for i := 0; i < MaxEntries; i++ {
		lb.Add(Entry{Name: "P", Score: 100, Seconds: 60})
	}
	if lb.Qualifies(100) {
		t.Error("equal-to-lowest score should not qualify a full board")
	}
	if !lb.Qualifies(101) {
		t.Error("higher-than-lowest score should qualify")
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	store := &Store{Dir: t.TempDir()}
	in := Settings{Language: LangDE, Muted: true, Fullscreen: true, Character: CharMehmet}
	if err := store.SaveSettings(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	out := store.LoadSettings()
	if out != in {
		t.Fatalf("round-trip mismatch: got %+v want %+v", out, in)
	}
}

func TestLoadSettingsDefaultsWhenMissing(t *testing.T) {
	store := &Store{Dir: t.TempDir()}
	if got := store.LoadSettings(); got != DefaultSettings() {
		t.Fatalf("missing file should yield defaults, got %+v", got)
	}
}

func TestLoadSettingsRepairsCorruptFile(t *testing.T) {
	dir := t.TempDir()
	store := &Store{Dir: dir}
	if err := os.WriteFile(store.settingsPath(), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := store.LoadSettings(); got != DefaultSettings() {
		t.Fatalf("corrupt file should yield defaults, got %+v", got)
	}
}

func TestLoadSettingsNormalizesUnknownValues(t *testing.T) {
	dir := t.TempDir()
	store := &Store{Dir: dir}
	if err := os.WriteFile(store.settingsPath(),
		[]byte(`{"language":"fr","character":"ghost"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got := store.LoadSettings()
	if got.Language != LangEN || got.Character != CharAli {
		t.Fatalf("unknown values not normalized: %+v", got)
	}
}

func TestLeaderboardRoundTrip(t *testing.T) {
	store := &Store{Dir: t.TempDir()}
	var lb Leaderboard
	lb.Add(Entry{Name: "Ali", Score: 999, Seconds: 42.5})
	lb.Add(Entry{Name: "Mehmet", Score: 500, Seconds: 61})
	if err := store.SaveLeaderboard(lb); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := store.LoadLeaderboard()
	if len(got.Entries) != 2 || got.Entries[0].Name != "Ali" || got.Entries[0].Score != 999 {
		t.Fatalf("leaderboard round-trip failed: %+v", got.Entries)
	}
}

func TestLeaderboardLoadSortsUnsortedFile(t *testing.T) {
	dir := t.TempDir()
	store := &Store{Dir: dir}
	// Deliberately out-of-order on disk.
	raw := `{"entries":[{"name":"low","score":10,"seconds":5},{"name":"high","score":90,"seconds":5}]}`
	if err := os.WriteFile(store.leaderboardPath(), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	got := store.LoadLeaderboard()
	if got.Entries[0].Name != "high" {
		t.Fatalf("loaded board not sorted: %+v", got.Entries)
	}
}

func TestDefaultStorePath(t *testing.T) {
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skip("no user config dir on this platform")
	}
	store, err := DefaultStore()
	if err != nil {
		t.Fatal(err)
	}
	if store.Dir != filepath.Join(base, appDirName) {
		t.Fatalf("unexpected store dir: %s", store.Dir)
	}
}
