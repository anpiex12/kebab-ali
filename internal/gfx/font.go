package gfx

import (
	"bytes"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// Font sources are parsed once at start-up. Parsing is pure CPU work (no
// graphics driver needed) so doing it in init() is safe. The Go font family is
// an original, BSD-licensed typeface shipped as byte data inside the module —
// it is not lifted from any game and needs no external file.
var (
	regularSource *text.GoTextFaceSource
	boldSource    *text.GoTextFaceSource
)

func init() {
	var err error
	if regularSource, err = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF)); err != nil {
		panic("gfx: parse regular font: " + err.Error())
	}
	if boldSource, err = text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF)); err != nil {
		panic("gfx: parse bold font: " + err.Error())
	}
}

// Face returns a regular-weight face at the given pixel size.
func Face(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: regularSource, Size: size}
}

// BoldFace returns a bold-weight face at the given pixel size.
func BoldFace(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: boldSource, Size: size}
}

// Align controls horizontal anchoring of drawn text.
type Align int

const (
	// AlignLeft anchors text at its left edge (x is the left edge).
	AlignLeft Align = iota
	// AlignCenter anchors text at its horizontal centre (x is the centre).
	AlignCenter
	// AlignRight anchors text at its right edge (x is the right edge).
	AlignRight
)

// TextOptions bundles the common text-drawing parameters.
type TextOptions struct {
	Size  float64
	Color color.Color
	Align Align
	Bold  bool
}

func (o TextOptions) face() *text.GoTextFace {
	if o.Bold {
		return BoldFace(o.Size)
	}
	return Face(o.Size)
}

// MeasureText returns the pixel width and height of s in the given face size.
func MeasureText(s string, size float64) (w, h float64) {
	return text.Measure(s, Face(size), size*1.2)
}

// DrawText renders s at (x,y) — the y coordinate is the top of the text — using
// the supplied options. Newlines are honoured.
func DrawText(dst *ebiten.Image, s string, x, y float64, o TextOptions) {
	face := o.face()
	op := &text.DrawOptions{}
	op.LineSpacing = o.Size * 1.2
	switch o.Align {
	case AlignCenter:
		w, _ := text.Measure(s, face, op.LineSpacing)
		x -= w / 2
	case AlignRight:
		w, _ := text.Measure(s, face, op.LineSpacing)
		x -= w
	}
	op.GeoM.Translate(x, y)
	if o.Color != nil {
		op.ColorScale.ScaleWithColor(o.Color)
	}
	text.Draw(dst, s, face, op)
}

// DrawTextShadow draws s with a 1px drop shadow underneath for readability over
// busy backgrounds, then the main text on top.
func DrawTextShadow(dst *ebiten.Image, s string, x, y float64, o TextOptions) {
	shadow := o
	shadow.Color = Black
	DrawText(dst, s, x+1, y+1, shadow)
	DrawText(dst, s, x, y, o)
}
