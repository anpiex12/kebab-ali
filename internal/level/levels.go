package level

import "image/color"

// grid is an authoring helper: a mutable rune buffer that the level functions
// paint into and then emit as the []string rows Build consumes. Authoring in
// code (rather than by hand-counting ASCII columns) keeps the maps correct.
type grid struct {
	w, h  int
	cells [][]rune
}

func newGrid(w, h int) *grid {
	cells := make([][]rune, h)
	for y := range cells {
		row := make([]rune, w)
		for x := range row {
			row[x] = ' '
		}
		cells[y] = row
	}
	return &grid{w: w, h: h, cells: cells}
}

func (g *grid) set(x, y int, r rune) {
	if x >= 0 && x < g.w && y >= 0 && y < g.h {
		g.cells[y][x] = r
	}
}

func (g *grid) hline(x0, x1, y int, r rune) {
	for x := x0; x <= x1; x++ {
		g.set(x, y, r)
	}
}

func (g *grid) coins(y int, xs ...int) {
	for _, x := range xs {
		g.set(x, y, 'c')
	}
}

func (g *grid) rows() []string {
	out := make([]string, g.h)
	for y, row := range g.cells {
		out[y] = string(row)
	}
	return out
}

// base returns a grid of the given width with a two-tile-thick solid floor and
// the player start near the left. Height is a fixed 15 tiles.
func base(w int) *grid {
	g := newGrid(w, 15)
	g.hline(0, w-1, 13, '#')
	g.hline(0, w-1, 14, '#')
	g.set(3, 12, 'P')
	return g
}

// pit carves a deadly sauce lake into the floor between columns x0..x1.
func (g *grid) pit(x0, x1 int) {
	for x := x0; x <= x1; x++ {
		g.set(x, 13, '~')
		g.set(x, 14, '~')
	}
}

// Themes ---------------------------------------------------------------------

func rgb(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 0xFF} }

var themeImbiss = Theme{
	Name:       "imbiss",
	SkyTop:     rgb(0x3A, 0x2A, 0x4A),
	SkyBottom:  rgb(0xE8, 0x8A, 0x54),
	Solid:      rgb(0xB5, 0x5A, 0x3A),
	SolidShade: rgb(0x8A, 0x3F, 0x28),
	Platform:   rgb(0x9A, 0x6A, 0x3A),
	Accent:     rgb(0xF5, 0x5A, 0x8A),
	Silhouette: rgb(0x2A, 0x1E, 0x3A),
}

var themeBazaar = Theme{
	Name:       "bazaar",
	SkyTop:     rgb(0x2E, 0x8A, 0x9A),
	SkyBottom:  rgb(0xF0, 0xD0, 0x90),
	Solid:      rgb(0xC8, 0xA0, 0x60),
	SolidShade: rgb(0x9A, 0x74, 0x40),
	Platform:   rgb(0xC0, 0x50, 0x50),
	Accent:     rgb(0xF5, 0xC5, 0x42),
	Silhouette: rgb(0x8A, 0x5A, 0x30),
}

var themeFactory = Theme{
	Name:       "factory",
	SkyTop:     rgb(0x22, 0x28, 0x33),
	SkyBottom:  rgb(0x50, 0x58, 0x64),
	Solid:      rgb(0x8C, 0x92, 0x9E),
	SolidShade: rgb(0x55, 0x5B, 0x66),
	Platform:   rgb(0x6A, 0x72, 0x80),
	Accent:     rgb(0x5A, 0xC8, 0x6A),
	Silhouette: rgb(0x2A, 0x30, 0x3A),
}

// Level definitions ----------------------------------------------------------

// Level1 builds "Closing-Time Chaos at the Imbiss" — a gentle tutorial that
// ends at Captain Garlic.
func Level1() *Level {
	g := base(100)
	g.set(6, 12, '1')           // welcome sign
	g.coins(11, 8, 9, 10)       // starter coins
	g.set(12, 9, '?')           // coin box
	g.set(14, 9, 'S')           // spit power-up box
	g.set(18, 12, 'o')          // tomato
	g.coins(11, 22, 23)
	g.pit(26, 28)               // small jumpable sauce pit
	g.hline(26, 28, 10, '=')    // platform over the pit
	g.coins(9, 27)
	g.set(34, 12, 'n')          // onion
	g.hline(38, 39, 10, 'B')    // bread blocks
	g.coins(9, 38, 39)
	g.set(44, 12, 'F')          // checkpoint
	g.set(50, 12, 'x')          // peperoni
	g.hline(54, 57, 9, '=')
	g.coins(8, 55, 56)
	g.pit(60, 62)
	g.set(66, 7, 'u')           // cucumber glider
	g.coins(11, 70, 71, 72)
	g.set(76, 12, '2')          // "mind the sauce" sign
	g.set(80, 9, '?')
	g.set(92, 12, 'D')          // Captain Garlic boss spawn
	return Build(Def{
		NameKey:   "level.1.name",
		Music:     "level1",
		Theme:     themeImbiss,
		BossKind:  "garlic",
		TimeLimit: 300,
		Signs:     map[rune]string{'1': "sign.welcome", '2': "sign.sauce"},
		Rows:      g.rows(),
	})
}

// Level2 builds "The Grand Bazaar" — floating carpets and spice pits, ending at
// the Onion Twins.
func Level2() *Level {
	g := base(112)
	g.set(6, 12, '1')            // bazaar sign
	g.set(10, 9, 'S')            // spit box (in case player arrives small)
	g.coins(11, 14, 15, 16)
	g.set(20, 12, 'o')
	g.hline(24, 27, 10, '=')     // flying carpet
	g.coins(9, 25, 26)
	g.pit(30, 33)
	g.hline(30, 33, 9, '=')      // carpet over pit
	g.set(38, 12, 'n')
	g.set(39, 12, 'n')
	g.hline(43, 46, 8, '=')
	g.coins(7, 44, 45)
	g.set(48, 7, 'u')
	g.set(52, 12, 'F')           // checkpoint
	g.pit(56, 59)
	g.hline(55, 60, 10, '=')
	g.set(64, 12, 'x')           // peperoni
	g.hline(68, 71, 9, '=')
	g.set(70, 8, 'c')
	g.set(74, 9, 'Y')            // Ayran box (rare)
	g.set(78, 12, 'o')
	g.pit(82, 85)
	g.hline(82, 85, 10, '=')
	g.set(90, 12, 'x')
	g.set(92, 12, 'n')
	g.set(96, 12, '2')           // carpet sign
	g.coins(11, 100, 101, 102)
	g.set(104, 12, 'D')          // Onion Twins boss spawn
	return Build(Def{
		NameKey:   "level.2.name",
		Music:     "level2",
		Theme:     themeBazaar,
		BossKind:  "onion",
		TimeLimit: 340,
		Signs:     map[rune]string{'1': "sign.bazaar", '2': "sign.bazaar"},
		Rows:      g.rows(),
	})
}

// Level3 builds "The Sauce Factory" — the industrial finale ending at the Giant
// Dürüm.
func Level3() *Level {
	g := base(114)
	g.set(6, 12, '1')            // danger sign
	g.set(10, 9, 'S')
	g.set(12, 9, '?')
	g.set(16, 12, 'x')           // peperoni gauntlet
	g.coins(11, 20, 21)
	g.pit(24, 27)
	g.hline(24, 27, 10, '=')
	g.set(31, 12, 'o')
	g.set(33, 12, 'o')
	g.hline(37, 40, 9, '=')      // gear platform
	g.coins(8, 38, 39)
	g.set(43, 12, 'x')
	g.pit(46, 49)
	g.hline(45, 50, 8, '=')
	g.set(48, 7, 'u')
	g.set(54, 12, 'F')           // checkpoint
	g.set(58, 12, 'n')
	g.hline(62, 64, 10, 'B')     // breakable pipe cover
	g.coins(9, 62, 63, 64)
	g.pit(68, 71)
	g.hline(67, 72, 9, '=')
	g.set(70, 8, 'Y')            // Ayran for the boss run
	g.set(76, 12, 'x')
	g.set(78, 12, 'x')
	g.hline(82, 85, 9, '=')
	g.set(88, 12, 'o')
	g.pit(90, 93)
	g.hline(90, 93, 10, '=')
	g.set(98, 12, 'u')
	g.set(101, 12, '2')          // final sign
	g.coins(11, 104, 105)
	g.set(106, 12, 'D')          // Giant Dürüm boss spawn
	return Build(Def{
		NameKey:   "level.3.name",
		Music:     "level3",
		Theme:     themeFactory,
		BossKind:  "durum",
		TimeLimit: 380,
		Signs:     map[rune]string{'1': "sign.factory", '2': "sign.factory"},
		Rows:      g.rows(),
	})
}

// All returns the three levels in play order.
func All() []*Level {
	return []*Level{Level1(), Level2(), Level3()}
}
