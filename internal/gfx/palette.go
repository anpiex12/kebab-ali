// Package gfx holds all of the game's home-made graphics: a warm "döner"
// colour palette, a procedural pixel-art sprite builder driven by text masks,
// a text renderer built on the (BSD-licensed) Go font, and small drawing
// helpers. Nothing here is taken from any other game — every pixel is
// generated in code.
//
// Sprite generation produces plain *image.RGBA values so it can be unit-tested
// head-lessly; the game converts those to *ebiten.Image once the graphics
// driver is live.
package gfx

import "image/color"

// Logical (internal) render resolution. The whole game is drawn into a canvas
// of this size and then scaled up with nearest-neighbour filtering, which gives
// the chunky 16-bit pixel look regardless of the window size.
const (
	ScreenWidth  = 480
	ScreenHeight = 270
)

// The warm Döner palette plus a few utility colours. Kept as a small named set
// so the whole game shares one coherent look.
var (
	Transparent = color.RGBA{0, 0, 0, 0}

	// Skin / dough tones.
	Skin      = color.RGBA{0xE8, 0xB0, 0x84, 0xFF}
	SkinDark  = color.RGBA{0xC8, 0x8A, 0x5A, 0xFF}
	Dough     = color.RGBA{0xF0, 0xD8, 0xA8, 0xFF}
	DoughDark = color.RGBA{0xD8, 0xB4, 0x78, 0xFF}

	// Meat / brown tones.
	Meat     = color.RGBA{0x9E, 0x5A, 0x2B, 0xFF}
	MeatDark = color.RGBA{0x6B, 0x3F, 0x1D, 0xFF}
	Brown    = color.RGBA{0x7A, 0x4A, 0x24, 0xFF}

	// Reds (apron, tomato, hot sauce).
	Red     = color.RGBA{0xD3, 0x3A, 0x2C, 0xFF}
	RedDark = color.RGBA{0x9E, 0x20, 0x18, 0xFF}
	Tomato  = color.RGBA{0xE2, 0x3B, 0x2E, 0xFF}

	// Greens (Mehmet's apron, cucumber, peperoni stem).
	Green     = color.RGBA{0x3A, 0xA0, 0x44, 0xFF}
	GreenDark = color.RGBA{0x23, 0x7A, 0x2E, 0xFF}
	Cucumber  = color.RGBA{0x6A, 0xA8, 0x4F, 0xFF}

	// Golds (döner taler, spice tins).
	Gold     = color.RGBA{0xF5, 0xC5, 0x42, 0xFF}
	GoldDark = color.RGBA{0xC9, 0x97, 0x1F, 0xFF}

	// Whites / creams (cap, ayran, garlic, sauce highlight).
	White  = color.RGBA{0xF7, 0xF3, 0xE8, 0xFF}
	Cream  = color.RGBA{0xEF, 0xE9, 0xD8, 0xFF}
	Garlic = color.RGBA{0xEA, 0xE4, 0xD0, 0xFF}

	// Onion.
	Onion     = color.RGBA{0xC9, 0xA8, 0xC8, 0xFF}
	OnionDark = color.RGBA{0x9A, 0x76, 0x9A, 0xFF}

	// Peperoni.
	Chili = color.RGBA{0xC0, 0x39, 0x2B, 0xFF}

	// Neutrals for UI, tiles and shadows.
	Black     = color.RGBA{0x1A, 0x12, 0x14, 0xFF}
	Ink       = color.RGBA{0x2A, 0x1C, 0x1E, 0xFF}
	Shadow    = color.RGBA{0x00, 0x00, 0x00, 0x66}
	Grey      = color.RGBA{0x8A, 0x82, 0x80, 0xFF}
	GreyDark  = color.RGBA{0x4A, 0x44, 0x44, 0xFF}
	Steel     = color.RGBA{0x8C, 0x92, 0x9E, 0xFF}
	SteelDark = color.RGBA{0x55, 0x5B, 0x66, 0xFF}

	// Brick / ground tones for the tiles.
	Brick     = color.RGBA{0xB5, 0x5A, 0x3A, 0xFF}
	BrickDark = color.RGBA{0x8A, 0x3F, 0x28, 0xFF}
	Dirt      = color.RGBA{0x8A, 0x5A, 0x34, 0xFF}
	DirtDark  = color.RGBA{0x66, 0x40, 0x22, 0xFF}
	Grass     = color.RGBA{0x5A, 0xA8, 0x44, 0xFF}

	// Sky / background accents used by the parallax layers.
	SkyTop    = color.RGBA{0x3A, 0x2A, 0x4A, 0xFF}
	SkyBottom = color.RGBA{0xE8, 0x8A, 0x54, 0xFF}
	Neon      = color.RGBA{0xF5, 0x5A, 0x8A, 0xFF}
	NeonBlue  = color.RGBA{0x4A, 0xC8, 0xF5, 0xFF}

	// Sauce lake (deadly). Red base swirled with white.
	SauceRed   = color.RGBA{0xD8, 0x2A, 0x28, 0xFF}
	SauceWhite = color.RGBA{0xF2, 0xEC, 0xE0, 0xFF}
)

// Lighten returns c blended toward white by t in [0,1].
func Lighten(c color.RGBA, t float64) color.RGBA {
	return blend(c, color.RGBA{255, 255, 255, c.A}, t)
}

// Darken returns c blended toward black by t in [0,1].
func Darken(c color.RGBA, t float64) color.RGBA {
	return blend(c, color.RGBA{0, 0, 0, c.A}, t)
}

func blend(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	lerp := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t) }
	return color.RGBA{lerp(a.R, b.R), lerp(a.G, b.G), lerp(a.B, b.B), a.A}
}
