package gfx

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ToEbiten uploads a CPU image to a GPU-backed *ebiten.Image. It must be called
// after the graphics driver is live (i.e. once the game loop has started).
func ToEbiten(src image.Image) *ebiten.Image {
	return ebiten.NewImageFromImage(src)
}

// FillRect draws a solid rectangle.
func FillRect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

// StrokeRect draws a rectangle outline of the given thickness.
func StrokeRect(dst *ebiten.Image, x, y, w, h, thick float64, c color.Color) {
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), float32(thick), c, false)
}

// FillCircle draws a solid (anti-aliased) circle centred at (cx,cy).
func FillCircle(dst *ebiten.Image, cx, cy, r float64, c color.Color) {
	vector.FillCircle(dst, float32(cx), float32(cy), float32(r), c, true)
}

// Line draws a straight line between two points.
func Line(dst *ebiten.Image, x0, y0, x1, y1, thick float64, c color.Color) {
	vector.StrokeLine(dst, float32(x0), float32(y0), float32(x1), float32(y1), float32(thick), c, true)
}

// Panel draws a filled rounded-feel UI box with a border, used by menus and
// speech bubbles.
func Panel(dst *ebiten.Image, x, y, w, h float64, fill, border color.Color) {
	FillRect(dst, x, y, w, h, fill)
	StrokeRect(dst, x, y, w, h, 2, border)
}

// VerticalGradient builds a w×h image that fades from top to bottom, handy for
// skies. It creates a GPU image, so call it after the game loop has started.
func VerticalGradient(w, h int, top, bottom color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	denom := float64(h - 1)
	if denom <= 0 {
		denom = 1
	}
	for y := 0; y < h; y++ {
		c := blend(top, bottom, float64(y)/denom)
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return ToEbiten(img)
}

// DrawSprite blits src onto dst at (x,y) scaled by scale, flipping horizontally
// when faceLeft is true. Nearest-neighbour filtering keeps the pixels crisp.
func DrawSprite(dst, src *ebiten.Image, x, y, scale float64, faceLeft bool) {
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	w := float64(src.Bounds().Dx())
	if faceLeft {
		op.GeoM.Scale(-scale, scale)
		op.GeoM.Translate(w*scale, 0)
	} else {
		op.GeoM.Scale(scale, scale)
	}
	op.GeoM.Translate(x, y)
	dst.DrawImage(src, op)
}

// DrawSpriteRot blits src onto dst centred at (cx,cy), rotated by angle radians
// and scaled — used for the spinning meat slice and rotating spit logo.
func DrawSpriteRot(dst, src *ebiten.Image, cx, cy, scale, angle float64) {
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	w := float64(src.Bounds().Dx())
	h := float64(src.Bounds().Dy())
	op.GeoM.Translate(-w/2, -h/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Rotate(angle)
	op.GeoM.Translate(cx, cy)
	dst.DrawImage(src, op)
}
