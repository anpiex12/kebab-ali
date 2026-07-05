package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/gfx"
)

const (
	screenW = gfx.ScreenWidth
	screenH = gfx.ScreenHeight
)

// fillGradient paints a top-to-bottom colour fade over the whole canvas.
func fillGradient(canvas *ebiten.Image, top, bottom color.RGBA) {
	for y := 0; y < screenH; y++ {
		t := float64(y) / float64(screenH-1)
		gfx.FillRect(canvas, 0, float64(y), screenW, 1, gfx.Blend(top, bottom, t))
	}
}

// drawLogo draws the rotating döner-spit emblem and the "DÖNER ALI" wordmark
// centred horizontally at cx with the spit at height y.
func drawLogo(canvas *ebiten.Image, a *Assets, g *Game, cx, y float64, tick int) {
	if a != nil {
		angle := float64(tick) * 0.06
		gfx.DrawSpriteRot(canvas, a.Spit, cx, y, 3.5, angle)
	}
	gfx.DrawTextShadow(canvas, g.T("app.title"), cx, y+26, gfx.TextOptions{
		Size: 34, Color: gfx.Gold, Align: gfx.AlignCenter, Bold: true,
	})
	gfx.DrawTextShadow(canvas, g.T("app.subtitle"), cx, y+62, gfx.TextOptions{
		Size: 12, Color: gfx.Cream, Align: gfx.AlignCenter,
	})
}

// drawMenuList renders a vertical list of options centred at cx, highlighting
// the selected one with a bobbing spit marker.
func drawMenuList(canvas *ebiten.Image, cx, y float64, options []string, selected, tick int) {
	for i, opt := range options {
		oy := y + float64(i)*20
		clr := gfx.Cream
		size := 16.0
		if i == selected {
			clr = gfx.Gold
			size = 18.0
			// bobbing dot markers either side (drawn, so no font glyph needed)
			off := math.Sin(float64(tick)*0.15) * 2
			w, _ := gfx.MeasureText(opt, size)
			gfx.FillCircle(canvas, cx-w/2-12+off, oy+size*0.5, 2.5, gfx.Red)
			gfx.FillCircle(canvas, cx+w/2+12-off, oy+size*0.5, 2.5, gfx.Red)
		}
		gfx.DrawTextShadow(canvas, opt, cx, oy, gfx.TextOptions{Size: size, Color: clr, Align: gfx.AlignCenter})
	}
}

// drawHint draws a dim helper line centred near the bottom of the screen.
func drawHint(canvas *ebiten.Image, text string) {
	gfx.DrawText(canvas, text, screenW/2, screenH-18, gfx.TextOptions{
		Size: 10, Color: gfx.Grey, Align: gfx.AlignCenter,
	})
}

// drawPanel draws a bordered box for dialogs and speech bubbles.
func drawPanel(canvas *ebiten.Image, x, y, w, h float64) {
	gfx.Panel(canvas, x, y, w, h, color.RGBA{0x1A, 0x12, 0x14, 0xE0}, gfx.Gold)
}

// drawSpeechBubble draws a small bordered bubble with a name and a taunt above a
// world position (already converted to canvas space).
func drawSpeechBubble(canvas *ebiten.Image, cx, bottomY float64, name, text string) {
	w := math.Max(measure(name, 10), measure(text, 11)) + 16
	h := 34.0
	x := cx - w/2
	y := bottomY - h - 6
	if x < 2 {
		x = 2
	}
	if x+w > screenW-2 {
		x = screenW - 2 - w
	}
	if y < 2 {
		y = 2
	}
	drawPanel(canvas, x, y, w, h)
	gfx.DrawText(canvas, name, x+w/2, y+5, gfx.TextOptions{Size: 10, Color: gfx.Gold, Align: gfx.AlignCenter, Bold: true})
	gfx.DrawText(canvas, text, x+w/2, y+18, gfx.TextOptions{Size: 11, Color: gfx.Cream, Align: gfx.AlignCenter})
}

func measure(s string, size float64) float64 {
	w, _ := gfx.MeasureText(s, size)
	return w
}
