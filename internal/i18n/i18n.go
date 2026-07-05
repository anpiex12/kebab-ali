// Package i18n is a tiny key/value localization system. Every UI string is
// looked up by key in the currently selected language, falling back to English
// and finally to the key itself, so a missing translation degrades gracefully
// instead of showing a blank label.
//
// Catalogs are loaded from any fs.FS (the embedded assets in the real game, an
// in-memory fstest.MapFS in tests), which keeps the package head-less.
package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// hasVerb reports whether s contains a printf-style verb, so T only formats
// strings that actually expect arguments (mixing plain and templated messages
// then never yields spurious %!(EXTRA ...) output).
func hasVerb(s string) bool { return strings.Contains(s, "%") }

// DefaultLang is the base language every lookup falls back to.
const DefaultLang = "en"

// Catalog holds the loaded translations and the active language.
type Catalog struct {
	// messages maps language code -> key -> text.
	messages map[string]map[string]string
	lang     string
}

// New returns an empty catalog set to the default language. Useful as a safe
// zero value when asset loading fails.
func New() *Catalog {
	return &Catalog{
		messages: map[string]map[string]string{DefaultLang: {}},
		lang:     DefaultLang,
	}
}

// Load reads every "*.json" file in dir of fsys as a flat map of key->text. The
// file name without extension is the language code (en.json -> "en"). The
// active language defaults to English when present, otherwise the first
// language found (alphabetically, for determinism).
func Load(fsys fs.FS, dir string) (*Catalog, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	c := &Catalog{messages: map[string]map[string]string{}}
	var langs []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := fs.ReadFile(fsys, path.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("i18n: %s: %w", e.Name(), err)
		}
		lang := strings.TrimSuffix(e.Name(), ".json")
		c.messages[lang] = m
		langs = append(langs, lang)
	}
	if len(langs) == 0 {
		return nil, fmt.Errorf("i18n: no language files in %q", dir)
	}
	sort.Strings(langs)
	if _, ok := c.messages[DefaultLang]; ok {
		c.lang = DefaultLang
	} else {
		c.lang = langs[0]
	}
	return c, nil
}

// SetLanguage switches the active language. Unknown codes are ignored so the
// catalog always stays in a valid state.
func (c *Catalog) SetLanguage(lang string) {
	if _, ok := c.messages[lang]; ok {
		c.lang = lang
	}
}

// Language returns the active language code.
func (c *Catalog) Language() string { return c.lang }

// Languages returns all available language codes, sorted.
func (c *Catalog) Languages() []string {
	out := make([]string, 0, len(c.messages))
	for k := range c.messages {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Has reports whether key exists in the active language or the fallback.
func (c *Catalog) Has(key string) bool {
	if _, ok := c.messages[c.lang][key]; ok {
		return true
	}
	_, ok := c.messages[DefaultLang][key]
	return ok
}

// T translates key in the active language. Missing keys fall back to English,
// then to the raw key. If args are supplied the result is run through
// fmt.Sprintf, so templates may contain verbs like %d and %s.
func (c *Catalog) T(key string, args ...any) string {
	s, ok := c.messages[c.lang][key]
	if !ok {
		if s, ok = c.messages[DefaultLang][key]; !ok {
			s = key
		}
	}
	if len(args) > 0 && hasVerb(s) {
		return fmt.Sprintf(s, args...)
	}
	return s
}
