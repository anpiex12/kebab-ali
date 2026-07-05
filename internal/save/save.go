// Package save persists player settings and the local high-score leaderboard
// as JSON files under the per-user config directory (os.UserConfigDir), e.g.
// %AppData%\DoenerAli on Windows or ~/.config/DoenerAli on Linux.
//
// The leaderboard ranking logic is pure and lives here so it can be unit-tested
// without any file system or rendering dependency; a Store binds it to disk.
package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// appDirName is the sub-directory created inside the user config dir.
const appDirName = "DoenerAli"

// MaxEntries is the number of high scores kept on the leaderboard (Top 10).
const MaxEntries = 10

// Language codes understood by the game.
const (
	LangEN = "en"
	LangDE = "de"
)

// Character identifiers.
const (
	CharAli    = "ali"
	CharMehmet = "mehmet"
)

// Settings holds all persisted user preferences.
type Settings struct {
	Language   string `json:"language"`
	Muted      bool   `json:"muted"`
	Fullscreen bool   `json:"fullscreen"`
	Character  string `json:"character"`
}

// DefaultSettings returns the initial settings: English, sound on, windowed,
// playing as Ali.
func DefaultSettings() Settings {
	return Settings{
		Language:   LangEN,
		Muted:      false,
		Fullscreen: false,
		Character:  CharAli,
	}
}

// normalize repairs any unknown/empty fields back to sane defaults so a
// hand-edited or partial file can never put the game into an invalid state.
func (s Settings) normalize() Settings {
	if s.Language != LangEN && s.Language != LangDE {
		s.Language = LangEN
	}
	if s.Character != CharAli && s.Character != CharMehmet {
		s.Character = CharAli
	}
	return s
}

// Entry is a single leaderboard record: who, how many points, and how long the
// run took (in seconds).
type Entry struct {
	Name    string  `json:"name"`
	Score   int     `json:"score"`
	Seconds float64 `json:"seconds"`
}

// Leaderboard is an ordered list of the best runs, highest score first.
type Leaderboard struct {
	Entries []Entry `json:"entries"`
}

// less reports whether entry a should rank ahead of b: higher score wins; on a
// tie the faster (smaller) time wins.
func less(a, b Entry) bool {
	if a.Score != b.Score {
		return a.Score > b.Score
	}
	return a.Seconds < b.Seconds
}

// Qualifies reports whether a run with the given score would earn a place on
// the board (either there is a free slot, or it beats the current last place).
func (l Leaderboard) Qualifies(score int) bool {
	if len(l.Entries) < MaxEntries {
		return true
	}
	return score > l.Entries[len(l.Entries)-1].Score
}

// Add inserts e in ranked order, trims the board back to MaxEntries, and
// returns the 1-based rank e landed at, or 0 if it did not make the board.
func (l *Leaderboard) Add(e Entry) int {
	l.Entries = append(l.Entries, e)
	sort.SliceStable(l.Entries, func(i, j int) bool {
		return less(l.Entries[i], l.Entries[j])
	})
	if len(l.Entries) > MaxEntries {
		l.Entries = l.Entries[:MaxEntries]
	}
	for i := range l.Entries {
		if l.Entries[i] == e {
			return i + 1
		}
	}
	return 0 // e was trimmed off the bottom
}

// Store binds the pure logic above to a directory on disk.
type Store struct {
	Dir string
}

// DefaultStore returns a Store rooted at <UserConfigDir>/DoenerAli.
func DefaultStore() (*Store, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &Store{Dir: filepath.Join(base, appDirName)}, nil
}

func (s *Store) settingsPath() string    { return filepath.Join(s.Dir, "settings.json") }
func (s *Store) leaderboardPath() string { return filepath.Join(s.Dir, "leaderboard.json") }

// writeJSON marshals v (pretty-printed) and writes it atomically-ish to path,
// creating the directory if necessary.
func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadSettings reads settings.json. A missing or corrupt file yields the
// defaults rather than an error, so first launch and hand-edits are safe.
func (s *Store) LoadSettings() Settings {
	data, err := os.ReadFile(s.settingsPath())
	if err != nil {
		return DefaultSettings()
	}
	var out Settings
	if err := json.Unmarshal(data, &out); err != nil {
		return DefaultSettings()
	}
	return out.normalize()
}

// SaveSettings writes settings.json.
func (s *Store) SaveSettings(v Settings) error {
	return writeJSON(s.settingsPath(), v.normalize())
}

// LoadLeaderboard reads leaderboard.json, returning an empty (but valid) board
// if the file is missing or corrupt. The result is always sorted and trimmed.
func (s *Store) LoadLeaderboard() Leaderboard {
	var lb Leaderboard
	data, err := os.ReadFile(s.leaderboardPath())
	if err != nil {
		return lb
	}
	if err := json.Unmarshal(data, &lb); err != nil {
		return Leaderboard{}
	}
	sort.SliceStable(lb.Entries, func(i, j int) bool {
		return less(lb.Entries[i], lb.Entries[j])
	})
	if len(lb.Entries) > MaxEntries {
		lb.Entries = lb.Entries[:MaxEntries]
	}
	return lb
}

// SaveLeaderboard writes leaderboard.json.
func (s *Store) SaveLeaderboard(lb Leaderboard) error {
	return writeJSON(s.leaderboardPath(), lb)
}
