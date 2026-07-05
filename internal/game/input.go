package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/anpiex12/kebab-ali/internal/entities"
)

// keyDown reports whether any of the keys is currently held.
func keyDown(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if ebiten.IsKeyPressed(k) {
			return true
		}
	}
	return false
}

// keyPressed reports whether any of the keys was pressed this frame.
func keyPressed(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}

// Semantic input groups shared across menus and gameplay.
var (
	keysLeft    = []ebiten.Key{ebiten.KeyArrowLeft, ebiten.KeyA}
	keysRight   = []ebiten.Key{ebiten.KeyArrowRight, ebiten.KeyD}
	keysUp      = []ebiten.Key{ebiten.KeyArrowUp, ebiten.KeyW}
	keysDown    = []ebiten.Key{ebiten.KeyArrowDown, ebiten.KeyS}
	keysJump    = []ebiten.Key{ebiten.KeySpace, ebiten.KeyW, ebiten.KeyArrowUp}
	keysThrow   = []ebiten.Key{ebiten.KeyX, ebiten.KeyJ}
	keysConfirm = []ebiten.Key{ebiten.KeyEnter, ebiten.KeyNumpadEnter, ebiten.KeySpace}
	keysBack    = []ebiten.Key{ebiten.KeyEscape, ebiten.KeyBackspace}
)

// gameplayInput maps the keyboard onto the device-agnostic entities.Input.
func gameplayInput() entities.Input {
	return entities.Input{
		Left:      keyDown(keysLeft...),
		Right:     keyDown(keysRight...),
		Run:       keyDown(ebiten.KeyShiftLeft, ebiten.KeyShiftRight),
		JumpHeld:  keyDown(keysJump...),
		JumpPress: keyPressed(keysJump...),
		Throw:     keyPressed(keysThrow...),
	}
}
