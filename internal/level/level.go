// Package level parses the game's ASCII tile maps into a Level: a grid of tiles
// plus neutral spawn descriptors (player start, checkpoints, boss, enemies,
// items, signs). It has no dependency on the entities or rendering packages —
// spawns are plain strings/coords that the game turns into live objects — which
// keeps the parser pure and unit-testable.
package level

import "image/color"

// TileSize is the width/height of one tile in pixels.
const TileSize = 16

// Tile is a static map cell.
type Tile uint8

const (
	// Empty is open air.
	Empty Tile = iota
	// Solid is impassable ground or brick.
	Solid
	// Platform is a thin standable block (rendered differently, same collision).
	Platform
	// Bread is a breakable fladenbread block.
	Bread
	// BoxCoin is a spice-tin "?" block that yields taler when bonked.
	BoxCoin
	// BoxSpit is a "?" block that yields a döner-spit power-up.
	BoxSpit
	// BoxAyran is a "?" block that yields an Ayran bottle.
	BoxAyran
	// BoxUsed is a spent block (still solid).
	BoxUsed
	// Sauce is a deadly döner-sauce lake (non-solid, instant death on contact).
	Sauce
)

// Solid reports whether the tile blocks movement.
func (t Tile) IsSolid() bool {
	switch t {
	case Solid, Platform, Bread, BoxCoin, BoxSpit, BoxAyran, BoxUsed:
		return true
	default:
		return false
	}
}

// Theme is the per-level colour set used by the renderer.
type Theme struct {
	Name       string
	SkyTop     color.RGBA
	SkyBottom  color.RGBA
	Solid      color.RGBA
	SolidShade color.RGBA
	Platform   color.RGBA
	Accent     color.RGBA
	Silhouette color.RGBA
}

// Point is a pixel-space coordinate.
type Point struct{ X, Y float64 }

// Spawn describes something to instantiate at a pixel position.
type Spawn struct {
	Kind string
	X, Y float64
}

// Sign is a decorative in-level signpost showing a localized message.
type Sign struct {
	X, Y    float64
	TextKey string
}

// Def is the authored definition of a level.
type Def struct {
	NameKey   string
	Music     string
	Theme     Theme
	BossKind  string
	TimeLimit int
	Signs     map[rune]string // digit rune in the map -> i18n key
	Rows      []string
}

// Level is a parsed, playable map.
type Level struct {
	Name      string
	Music     string
	Theme     Theme
	BossKind  string
	TimeLimit int

	W, H  int
	tiles []Tile

	PlayerStart Point
	Checkpoints []Point
	Boss        Point
	Enemies     []Spawn
	Items       []Spawn
	Signs       []Sign
}

// enemyGlyphs maps map runes to enemy kind strings.
var enemyGlyphs = map[rune]string{
	'o': "tomato",
	'n': "onion",
	'x': "peperoni",
	'u': "cucumber",
}

// Build parses a Def into a Level.
func Build(def Def) *Level {
	h := len(def.Rows)
	w := 0
	for _, r := range def.Rows {
		if len(r) > w {
			w = len(r)
		}
	}
	l := &Level{
		Name:      def.NameKey,
		Music:     def.Music,
		Theme:     def.Theme,
		BossKind:  def.BossKind,
		TimeLimit: def.TimeLimit,
		W:         w,
		H:         h,
		tiles:     make([]Tile, w*h),
	}
	ts := float64(TileSize)
	for y, row := range def.Rows {
		for x := 0; x < len(row); x++ {
			r := rune(row[x])
			px, py := float64(x)*ts, float64(y)*ts
			switch r {
			case '#':
				l.set(x, y, Solid)
			case '=':
				l.set(x, y, Platform)
			case 'B':
				l.set(x, y, Bread)
			case '?':
				l.set(x, y, BoxCoin)
			case 'S':
				l.set(x, y, BoxSpit)
			case 'Y':
				l.set(x, y, BoxAyran)
			case '~':
				l.set(x, y, Sauce)
			case 'P':
				l.PlayerStart = Point{px, py}
			case 'F':
				l.Checkpoints = append(l.Checkpoints, Point{px, py})
			case 'D':
				l.Boss = Point{px, py}
			case 'c':
				l.Items = append(l.Items, Spawn{Kind: "coin", X: px, Y: py})
			default:
				if kind, ok := enemyGlyphs[r]; ok {
					l.Enemies = append(l.Enemies, Spawn{Kind: kind, X: px, Y: py})
				} else if key, ok := def.Signs[r]; ok {
					l.Signs = append(l.Signs, Sign{X: px, Y: py, TextKey: key})
				}
				// spaces, dots and unknown runes are left as Empty
			}
		}
	}
	return l
}

func (l *Level) idx(tx, ty int) int { return ty*l.W + tx }

func (l *Level) set(tx, ty int, t Tile) {
	if tx >= 0 && tx < l.W && ty >= 0 && ty < l.H {
		l.tiles[l.idx(tx, ty)] = t
	}
}

// TileAt returns the tile at grid coords, treating out-of-bounds as Empty.
func (l *Level) TileAt(tx, ty int) Tile {
	if tx < 0 || tx >= l.W || ty < 0 || ty >= l.H {
		return Empty
	}
	return l.tiles[l.idx(tx, ty)]
}

// SetTile mutates a tile (used to break bread blocks and spend "?" boxes).
func (l *Level) SetTile(tx, ty int, t Tile) { l.set(tx, ty, t) }

// Solid implements the collision query. The left/right edges of the map are
// walls; the top is open sky and the bottom is an open pit (so falling off the
// bottom is deadly rather than solid).
func (l *Level) Solid(tx, ty int) bool {
	if tx < 0 || tx >= l.W {
		return true
	}
	if ty < 0 || ty >= l.H {
		return false
	}
	return l.tiles[l.idx(tx, ty)].IsSolid()
}

// TileSize returns the pixel size of a tile (satisfies the entities.World shape).
func (l *Level) TileSize() int { return TileSize }

// PixelWidth returns the level width in pixels.
func (l *Level) PixelWidth() float64 { return float64(l.W * TileSize) }

// PixelHeight returns the level height in pixels.
func (l *Level) PixelHeight() float64 { return float64(l.H * TileSize) }

// SauceOverlap reports whether any tile overlapping the pixel rectangle
// (x,y,w,h) is a deadly sauce tile.
func (l *Level) SauceOverlap(x, y, w, h float64) bool {
	ts := float64(TileSize)
	x0, x1 := int(x/ts), int((x+w-1)/ts)
	y0, y1 := int(y/ts), int((y+h-1)/ts)
	for ty := y0; ty <= y1; ty++ {
		for tx := x0; tx <= x1; tx++ {
			if l.TileAt(tx, ty) == Sauce {
				return true
			}
		}
	}
	return false
}
