package i18n

import (
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"lang/en.json": {Data: []byte(`{"hello":"Hello","only_en":"English only","score":"Score: %d"}`)},
		"lang/de.json": {Data: []byte(`{"hello":"Hallo","score":"Punkte: %d"}`)},
	}
}

func TestLoadDefaultsToEnglish(t *testing.T) {
	c, err := Load(testFS(), "lang")
	if err != nil {
		t.Fatal(err)
	}
	if c.Language() != "en" {
		t.Fatalf("default language should be en, got %s", c.Language())
	}
	if got := c.T("hello"); got != "Hello" {
		t.Fatalf("T(hello)=%q want Hello", got)
	}
}

func TestSwitchLanguage(t *testing.T) {
	c, _ := Load(testFS(), "lang")
	c.SetLanguage("de")
	if got := c.T("hello"); got != "Hallo" {
		t.Fatalf("T(hello) in de = %q want Hallo", got)
	}
}

func TestFallbackToEnglish(t *testing.T) {
	c, _ := Load(testFS(), "lang")
	c.SetLanguage("de")
	// "only_en" is absent from de.json; must fall back to the English value.
	if got := c.T("only_en"); got != "English only" {
		t.Fatalf("fallback failed: got %q", got)
	}
}

func TestFallbackToKey(t *testing.T) {
	c, _ := Load(testFS(), "lang")
	if got := c.T("does_not_exist"); got != "does_not_exist" {
		t.Fatalf("missing key should return the key itself, got %q", got)
	}
}

func TestTWithArgs(t *testing.T) {
	c, _ := Load(testFS(), "lang")
	c.SetLanguage("de")
	if got := c.T("score", 42); got != "Punkte: 42" {
		t.Fatalf("templated T failed: got %q", got)
	}
}

func TestSetUnknownLanguageIsIgnored(t *testing.T) {
	c, _ := Load(testFS(), "lang")
	c.SetLanguage("xx")
	if c.Language() != "en" {
		t.Fatalf("unknown language should be ignored, got %s", c.Language())
	}
}

func TestLanguagesSorted(t *testing.T) {
	c, _ := Load(testFS(), "lang")
	langs := c.Languages()
	if len(langs) != 2 || langs[0] != "de" || langs[1] != "en" {
		t.Fatalf("languages not sorted as expected: %v", langs)
	}
}

func TestHas(t *testing.T) {
	c, _ := Load(testFS(), "lang")
	c.SetLanguage("de")
	if !c.Has("only_en") {
		t.Error("Has should be true via English fallback")
	}
	if c.Has("nope") {
		t.Error("Has should be false for a truly missing key")
	}
}

func TestLoadErrorsWithoutFiles(t *testing.T) {
	if _, err := Load(fstest.MapFS{}, "lang"); err == nil {
		t.Error("expected an error when no language files are present")
	}
}

func TestNewCatalogIsUsable(t *testing.T) {
	c := New()
	if c.Language() != DefaultLang {
		t.Fatalf("New catalog language = %s", c.Language())
	}
	if got := c.T("anything"); got != "anything" {
		t.Fatalf("empty catalog should echo the key, got %q", got)
	}
}
