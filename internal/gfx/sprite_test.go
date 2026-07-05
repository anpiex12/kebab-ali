package gfx

import (
	"image"
	"image/color"
	"testing"
)

// opaqueCount reports how many fully-opaque pixels an image has, a cheap way to
// assert a sprite actually drew something.
func opaqueCount(img interface{ At(x, y int) color.Color }, w, h int) int {
	n := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a == 0xffff {
				n++
			}
		}
	}
	return n
}

func TestMaskBuildDimensions(t *testing.T) {
	m := Mask{"AB", "CDE"}
	img := m.Build(map[rune]color.RGBA{'A': Red, 'B': Red, 'C': Red, 'D': Red, 'E': Red})
	if b := img.Bounds(); b.Dx() != 3 || b.Dy() != 2 {
		t.Fatalf("width should be longest row: got %dx%d want 3x2", b.Dx(), b.Dy())
	}
}

func TestMaskTransparentRunes(t *testing.T) {
	m := Mask{".A."}
	img := m.Build(map[rune]color.RGBA{'A': Red})
	if _, _, _, a := img.At(0, 0).RGBA(); a != 0 {
		t.Error("dot should be transparent")
	}
	if _, _, _, a := img.At(1, 0).RGBA(); a == 0 {
		t.Error("A should be opaque")
	}
}

func TestHeroFrameDimensions(t *testing.T) {
	frames := map[string]Mask{
		"idle": heroIdle,
		"run1": heroRun1,
		"run2": heroRun2,
		"jump": heroJump,
	}
	for name, m := range frames {
		if len(m) != 16 {
			t.Errorf("%s: height=%d want 16", name, len(m))
		}
		for i, row := range m {
			if len(row) != 12 {
				t.Errorf("%s row %d: width=%d want 12 (%q)", name, i, len(row), row)
			}
		}
	}
}

func TestBuildCharacterHasPixels(t *testing.T) {
	cs := BuildCharacter(Red, RedDark)
	if n := opaqueCount(cs.Idle, cs.Idle.Bounds().Dx(), cs.Idle.Bounds().Dy()); n < 40 {
		t.Errorf("idle frame looks empty: %d opaque pixels", n)
	}
	if len(cs.Run) != 2 {
		t.Fatalf("expected 2 run frames, got %d", len(cs.Run))
	}
}

func TestApronColourApplied(t *testing.T) {
	// Ali (red) and Mehmet (green) must differ somewhere on the apron.
	ali := BuildCharacter(Red, RedDark).Idle
	mehmet := BuildCharacter(Green, GreenDark).Idle
	differs := false
	b := ali.Bounds()
	for y := 0; y < b.Dy() && !differs; y++ {
		for x := 0; x < b.Dx(); x++ {
			if ali.At(x, y) != mehmet.At(x, y) {
				differs = true
				break
			}
		}
	}
	if !differs {
		t.Error("Ali and Mehmet sprites should differ by apron colour")
	}
}

func TestItemSpritesBuild(t *testing.T) {
	items := map[string]*image.RGBA{
		"taler": Taler(),
		"spit":  DoenerSpit(),
		"ayran": Ayran(),
		"meat":  MeatSlice(),
	}
	for name, img := range items {
		b := img.Bounds()
		if b.Dx() == 0 || b.Dy() == 0 {
			t.Errorf("%s sprite has zero size", name)
		}
		if n := opaqueCount(img, b.Dx(), b.Dy()); n < 4 {
			t.Errorf("%s sprite looks empty (%d opaque px)", name, n)
		}
	}
}
