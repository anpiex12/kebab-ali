package game

import (
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// captureDir, when set via SetCapture, activates a scripted screenshot run: the
// game auto-drives through several scenes, dumps PNGs of the low-res canvas and
// then quits. It exists purely for visual QA and is off unless the environment
// asks for it.
var captureDir string

// SetCapture enables the screenshot run, writing PNGs into dir.
func SetCapture(dir string) { captureDir = dir }

type captureDirector struct {
	f    int
	shot string
}

// step drives the scripted capture schedule one frame at a time.
func (d *captureDirector) step(g *Game) {
	d.f++
	switch d.f {
	case 2:
		g.Assets = LoadAssets()
		g.SwitchTo(newMenuScene())
	case 10:
		d.shot = "01-menu"
	case 14:
		ps := newPlayScene(g, 0, nil)
		ps.state = psPlaying
		g.SwitchTo(ps)
	case 60:
		d.shot = "02-play"
	case 64:
		ps := newPlayScene(g, 0, nil)
		ps.state = psPlaying
		ps.player.PlaceAt(ps.lvl.Boss.X-90, ps.lvl.Boss.Y-32)
		g.SwitchTo(ps)
	case 150:
		d.shot = "03-boss"
	case 156:
		g.SwitchTo(newCreditsScene())
	case 170:
		d.shot = "04-credits"
	case 176:
		lb := newLeaderboardScene()
		g.SwitchTo(lb)
	case 190:
		d.shot = "05-leaderboard"
	case 196:
		g.Quit()
	}
}

// saveCanvas writes the current low-res canvas to a PNG.
func saveCanvas(img *ebiten.Image, path string) {
	b := img.Bounds()
	buf := make([]byte, 4*b.Dx()*b.Dy())
	img.ReadPixels(buf)
	rgba := &image.RGBA{Pix: buf, Stride: 4 * b.Dx(), Rect: image.Rect(0, 0, b.Dx(), b.Dy())}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_ = png.Encode(f, rgba)
}

// maybeCapture runs the director each frame (called from Update).
func (g *Game) maybeCapture() {
	if captureDir == "" {
		return
	}
	if g.cap == nil {
		g.cap = &captureDirector{}
	}
	g.cap.step(g)
}

// maybeSaveShot writes a pending screenshot (called from Draw after painting).
func (g *Game) maybeSaveShot() {
	if g.cap != nil && g.cap.shot != "" {
		saveCanvas(g.canvas, filepath.Join(captureDir, g.cap.shot+".png"))
		g.cap.shot = ""
	}
}
