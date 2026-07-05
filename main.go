// Command doener-ali is the entry point for DÖNER ALI, a kebab-themed
// jump 'n' run built on Ebitengine. All assets (sprites, sounds, music) are
// generated in code or embedded here, so the compiled binary is fully
// self-contained — no external files at runtime.
package main

import (
	"embed"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/game"
)

//go:embed all:assets
var assetsFS embed.FS

// version is stamped in at build time via -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("Döner Ali")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Optional visual-QA mode: DONER_CAPTURE=<dir> auto-drives the game and
	// dumps screenshots, then quits.
	if dir := os.Getenv("DONER_CAPTURE"); dir != "" {
		game.SetCapture(dir)
	}

	g := game.New(assetsFS, version)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
