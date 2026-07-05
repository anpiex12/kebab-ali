package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/entities"
	"github.com/anpiex12/kebab-ali/internal/gfx"
	"github.com/anpiex12/kebab-ali/internal/level"
)

const tileF = float64(level.TileSize)

// bakeStatic renders the unchanging tiles (ground and platforms) into a single
// image the size of the whole level, which is then blitted per frame — far
// cheaper than drawing every tile individually.
func bakeStatic(l *level.Level) *ebiten.Image {
	img := ebiten.NewImage(int(l.PixelWidth()), int(l.PixelHeight()))
	for ty := 0; ty < l.H; ty++ {
		for tx := 0; tx < l.W; tx++ {
			x, y := float64(tx)*tileF, float64(ty)*tileF
			switch l.TileAt(tx, ty) {
			case level.Solid:
				gfx.FillRect(img, x, y, tileF, tileF, l.Theme.Solid)
				gfx.FillRect(img, x, y+tileF-3, tileF, 3, l.Theme.SolidShade)
				gfx.FillRect(img, x, y, tileF, 2, gfx.Lighten(l.Theme.Solid, 0.25))
			case level.Platform:
				gfx.FillRect(img, x, y, tileF, tileF*0.6, l.Theme.Platform)
				gfx.FillRect(img, x, y, tileF, 2, gfx.Lighten(l.Theme.Platform, 0.3))
			}
		}
	}
	return img
}

// drawDynamicTiles draws the tiles that can change or animate (blocks and
// sauce) for the currently visible columns only.
func drawDynamicTiles(canvas *ebiten.Image, l *level.Level, camX, camY float64, tick int) {
	x0 := int(camX/tileF) - 1
	x1 := int((camX+screenW)/tileF) + 1
	for tx := x0; tx <= x1; tx++ {
		for ty := 0; ty < l.H; ty++ {
			sx := float64(tx)*tileF - camX
			sy := float64(ty)*tileF - camY
			switch l.TileAt(tx, ty) {
			case level.Bread:
				drawBread(canvas, sx, sy)
			case level.BoxCoin:
				drawBox(canvas, sx, sy, "?", gfx.Gold, tick)
			case level.BoxSpit:
				drawBox(canvas, sx, sy, "!", gfx.Red, tick)
			case level.BoxAyran:
				drawBox(canvas, sx, sy, "?", gfx.NeonBlue, tick)
			case level.BoxUsed:
				gfx.FillRect(canvas, sx, sy, tileF, tileF, gfx.Brown)
				gfx.StrokeRect(canvas, sx, sy, tileF, tileF, 1, gfx.MeatDark)
			case level.Sauce:
				drawSauce(canvas, sx, sy, tick)
			}
		}
	}
}

func drawBread(canvas *ebiten.Image, x, y float64) {
	gfx.FillRect(canvas, x, y, tileF, tileF, gfx.Dough)
	gfx.StrokeRect(canvas, x, y, tileF, tileF, 1, gfx.DoughDark)
	gfx.Line(canvas, x+3, y+3, x+tileF-3, y+tileF-3, 1, gfx.DoughDark)
	gfx.Line(canvas, x+tileF-3, y+3, x+3, y+tileF-3, 1, gfx.DoughDark)
}

func drawBox(canvas *ebiten.Image, x, y float64, glyph string, glyphColor color.RGBA, tick int) {
	bob := math.Sin(float64(tick)*0.12) * 0.5
	gfx.FillRect(canvas, x, y+bob, tileF, tileF, gfx.GoldDark)
	gfx.FillRect(canvas, x+1, y+1+bob, tileF-2, tileF-2, gfx.Gold)
	gfx.DrawText(canvas, glyph, x+tileF/2, y+2+bob, gfx.TextOptions{Size: 12, Color: glyphColor, Align: gfx.AlignCenter, Bold: true})
}

func drawSauce(canvas *ebiten.Image, x, y float64, tick int) {
	gfx.FillRect(canvas, x, y, tileF, tileF, gfx.SauceRed)
	// bubbling white swirls
	for i := 0; i < 2; i++ {
		bx := x + 4 + float64(i)*7
		by := y + 4 + math.Sin(float64(tick)*0.15+float64(i)+x)*2
		gfx.FillCircle(canvas, bx, by, 2, gfx.SauceWhite)
	}
	gfx.FillRect(canvas, x, y, tileF, 2, gfx.Lighten(gfx.SauceRed, 0.2))
}

// --- parallax backgrounds ---------------------------------------------------

// drawParallax draws two silhouette layers scrolling slower than the camera to
// give depth. The shapes vary by theme name.
func drawParallax(canvas *ebiten.Image, theme level.Theme, camX float64, tick int) {
	far := gfx.Darken(theme.Silhouette, 0.15)
	near := theme.Silhouette
	drawSilhouetteLayer(canvas, far, camX*0.2, 150, 55, 26)
	drawSilhouetteLayer(canvas, near, camX*0.45, 175, 40, 40)
	_ = tick
}

func drawSilhouetteLayer(canvas *ebiten.Image, c color.RGBA, offset, baseY, spacing, height float64) {
	start := -math.Mod(offset, spacing)
	for x := start; x < screenW+spacing; x += spacing {
		h := height * (0.6 + 0.4*math.Abs(math.Sin(x*0.3+offset*0.01)))
		w := spacing * 0.7
		gfx.FillRect(canvas, x, baseY-h, w, h+40, c)
		// a little minaret/tent cap
		gfx.FillCircle(canvas, x+w/2, baseY-h, w*0.28, c)
	}
}

// drawSparkle draws a small four-point star/sparkle, used where an emoji star
// would be (the Go font has no emoji glyphs).
func drawSparkle(canvas *ebiten.Image, x, y, r float64, c color.RGBA) {
	gfx.FillCircle(canvas, x, y, r*0.45, c)
	gfx.Line(canvas, x-r, y, x+r, y, 1.5, c)
	gfx.Line(canvas, x, y-r, x, y+r, 1.5, c)
	gfx.Line(canvas, x-r*0.6, y-r*0.6, x+r*0.6, y+r*0.6, 1, c)
	gfx.Line(canvas, x-r*0.6, y+r*0.6, x+r*0.6, y-r*0.6, 1, c)
}

// drawCitySilhouette is a simpler skyline used behind the menus.
func drawCitySilhouette(canvas *ebiten.Image, c color.RGBA, tick int) {
	for x := 0.0; x < screenW; x += 34 {
		h := 30 + 20*math.Abs(math.Sin(x*0.21))
		gfx.FillRect(canvas, x, screenH-h-20, 28, h+20, c)
		gfx.FillCircle(canvas, x+14, screenH-h-20, 6, c)
	}
	_ = tick
}

// --- entities ---------------------------------------------------------------

func drawPlayer(canvas *ebiten.Image, p *entities.Player, f CharFrames, camX, camY float64) {
	// Blink out on alternate frames while invulnerable.
	if p.Invincible() && (p.AnimTick/4)%2 == 0 {
		return
	}
	frame := f.Idle
	switch {
	case !p.OnGround():
		frame = f.Jump
	case math.Abs(p.VX) > 0.3 && len(f.Run) > 0:
		frame = f.Run[(p.AnimTick/6)%len(f.Run)]
	}
	scale := 1.0
	if p.Power.Big() {
		scale = 1.5
	}
	sw := float64(frame.Bounds().Dx()) * scale
	sh := float64(frame.Bounds().Dy()) * scale
	x := p.Body.CenterX() - sw/2 - camX
	y := p.Body.Bottom() - sh - camY
	gfx.DrawSprite(canvas, frame, x, y, scale, p.Facing < 0)
	// Master Ali gets a golden sizzle mark.
	if p.Power == entities.Master {
		gfx.FillCircle(canvas, p.Body.CenterX()-camX, y-2, 2, gfx.Gold)
	}
}

func drawEnemy(canvas *ebiten.Image, e entities.Enemy, camX, camY float64) {
	r := e.Rect()
	cx := r.CenterX() - camX
	cy := r.CenterY() - camY
	squish := 1.0
	if e.Dying() {
		squish = 0.4
		cy = r.Bottom() - camY - r.H*0.2
	}
	switch e.Kind() {
	case entities.KindTomato:
		gfx.FillCircle(canvas, cx, cy, r.W/2*squish, gfx.Tomato)
		gfx.FillRect(canvas, cx-1, r.Y-camY-2, 2, 4, gfx.Green) // stem
		eyeDots(canvas, cx, cy-1, e.Facing())
	case entities.KindOnion:
		gfx.FillCircle(canvas, cx, cy, r.W/2, gfx.Onion)
		gfx.FillCircle(canvas, cx, cy, r.W/2*0.62, gfx.Cream)
		gfx.FillCircle(canvas, cx, cy, r.W/2*0.3, gfx.Onion)
		eyeDots(canvas, cx, cy, e.Facing())
	case entities.KindPeperoni:
		gfx.FillRect(canvas, cx-r.W/2, cy-r.H/2, r.W, r.H, gfx.Chili)
		gfx.FillCircle(canvas, cx, cy+r.H/2, r.W/2, gfx.Chili)
		gfx.FillRect(canvas, cx-1, r.Y-camY-3, 2, 4, gfx.Green)
		eyeDots(canvas, cx, cy-2, e.Facing())
	case entities.KindCucumber:
		gfx.FillCircle(canvas, cx, cy, r.W/2*0.9, gfx.Cucumber)
		gfx.FillRect(canvas, cx-r.W/2, cy-2, r.W, 4, gfx.Cucumber)
		// flapping wings
		flap := math.Sin(float64(e.AnimTick())*0.4) * 3
		gfx.FillCircle(canvas, cx-r.W/2, cy-flap, 3, gfx.Cream)
		gfx.FillCircle(canvas, cx+r.W/2, cy-flap, 3, gfx.Cream)
		eyeDots(canvas, cx, cy, e.Facing())
	}
}

// eyeDots draws two little eyes biased toward the facing direction.
func eyeDots(canvas *ebiten.Image, cx, cy float64, facing int) {
	off := float64(facing) * 1.5
	gfx.FillCircle(canvas, cx-2+off, cy-1, 1.3, gfx.Ink)
	gfx.FillCircle(canvas, cx+2+off, cy-1, 1.3, gfx.Ink)
}

func drawItem(canvas *ebiten.Image, it *entities.Item, a *Assets, camX, camY float64) {
	bob := math.Sin(float64(it.AnimTick)*0.12) * 1.5
	x := it.Body.X - camX
	y := it.Body.Y - camY + bob
	switch it.Kind {
	case entities.ItemTaler:
		if a != nil {
			gfx.DrawSprite(canvas, a.Taler, x, y, 1, false)
		}
	case entities.ItemSpit:
		if a != nil {
			gfx.DrawSprite(canvas, a.Spit, x, y, 1, false)
		}
	case entities.ItemAyran:
		if a != nil {
			gfx.DrawSprite(canvas, a.Ayran, x, y, 1, false)
		}
	}
}

func drawProjectile(canvas *ebiten.Image, pr *entities.Projectile, a *Assets, camX, camY float64) {
	r := pr.Rect()
	cx := r.CenterX() - camX
	cy := r.CenterY() - camY
	switch pr.Kind {
	case entities.MeatSliceProj:
		if a != nil {
			gfx.DrawSpriteRot(canvas, a.MeatSlice, cx, cy, 1, float64(pr.AnimTick)*0.4)
		}
	case entities.ChiliProj:
		gfx.FillCircle(canvas, cx, cy, 3, gfx.Chili)
		gfx.FillCircle(canvas, cx-float64(entitiesSign(pr))*2, cy, 2, gfx.Red)
	case entities.GarlicProj:
		gfx.FillCircle(canvas, cx, cy, 5, gfx.Cream)
		gfx.FillCircle(canvas, cx+2, cy-1, 3, gfx.White)
	case entities.RingProj:
		gfx.FillCircle(canvas, cx, cy, 5, gfx.Onion)
		gfx.FillCircle(canvas, cx, cy, 2.5, gfx.SkyTop)
	}
}

func entitiesSign(pr *entities.Projectile) float64 {
	if pr.VX < 0 {
		return -1
	}
	return 1
}

func drawBoss(canvas *ebiten.Image, b entities.Boss, camX, camY float64) {
	r := b.Rect()
	cx := r.CenterX() - camX
	cy := r.CenterY() - camY
	flash := b.HP() >= 0 && (b.AnimTick()/3)%2 == 0
	switch b.Kind() {
	case "garlic":
		body := gfx.Garlic
		if flash && bossJustHit(b) {
			body = gfx.White
		}
		gfx.FillCircle(canvas, cx, cy+2, r.W/2, body)
		gfx.FillCircle(canvas, cx-4, cy, r.W/3, body)
		gfx.FillCircle(canvas, cx+4, cy, r.W/3, body)
		gfx.FillRect(canvas, cx-1, r.Y-camY-3, 2, 5, gfx.Green)
		eyeDots(canvas, cx, cy, b.Facing())
	case "onion":
		gfx.FillCircle(canvas, cx, cy, r.W/2, gfx.Onion)
		gfx.FillCircle(canvas, cx, cy, r.W/2*0.6, gfx.Cream)
		gfx.FillRect(canvas, cx-1, r.Y-camY-3, 2, 4, gfx.Green)
		eyeDots(canvas, cx, cy, b.Facing())
	case "durum":
		// A big wrapped dürüm; the wrap seams spin while it rolls.
		gfx.FillCircle(canvas, cx, cy, r.W/2, gfx.Dough)
		gfx.FillRect(canvas, cx-r.W/2, cy-r.H/2, r.W, r.H, gfx.Dough)
		gfx.StrokeRect(canvas, cx-r.W/2, cy-r.H/2, r.W, r.H, 2, gfx.DoughDark)
		seam := math.Mod(float64(b.AnimTick())*3, r.W)
		for s := seam - r.W; s < r.W; s += 10 {
			gfx.Line(canvas, cx-r.W/2+s, cy-r.H/2, cx-r.W/2+s-6, cy+r.H/2, 1, gfx.DoughDark)
		}
		// a peeking Achmet moustache face
		gfx.FillRect(canvas, cx-6, cy-2, 12, 3, gfx.MeatDark)
		eyeDots(canvas, cx, cy-6, b.Facing())
	}
}

// bossJustHit reports whether the boss is in its post-hit flash window (approx).
func bossJustHit(b entities.Boss) bool { return b.HP() < b.MaxHP() }

func drawSign(canvas *ebiten.Image, g *Game, s level.Sign, camX, camY float64) {
	x := s.X - camX
	y := s.Y - camY
	gfx.FillRect(canvas, x+6, y+4, 3, tileF-2, gfx.Brown) // post
	text := g.T(s.TextKey)
	w := measure(text, 8) + 8
	gfx.Panel(canvas, x+7-w/2, y-8, w, 12, gfx.Dough, gfx.MeatDark)
	gfx.DrawText(canvas, text, x+7, y-6, gfx.TextOptions{Size: 8, Color: gfx.Ink, Align: gfx.AlignCenter})
}
