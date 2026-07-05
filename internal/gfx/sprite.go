package gfx

import (
	"image"
	"image/color"
)

// Mask is a small text grid describing a sprite. Each rune is one pixel and is
// looked up in a palette; '.', ' ' and any rune absent from the palette are
// left transparent. Masks are ASCII so byte index == pixel column.
type Mask []string

// Build rasterises a mask into an *image.RGBA using pal. The image width is the
// longest row; shorter rows are padded with transparency on the right.
func (m Mask) Build(pal map[rune]color.RGBA) *image.RGBA {
	h := len(m)
	w := 0
	for _, row := range m {
		if len(row) > w {
			w = len(row)
		}
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y, row := range m {
		for x := 0; x < len(row); x++ {
			r := rune(row[x])
			if r == '.' || r == ' ' {
				continue
			}
			if c, ok := pal[r]; ok {
				img.SetRGBA(x, y, c)
			}
		}
	}
	return img
}

// CharacterSprites is the full frame set for a playable character.
type CharacterSprites struct {
	Idle *image.RGBA
	Run  []*image.RGBA // walk cycle
	Jump *image.RGBA
}

// Character masks. Every row is exactly 12 columns wide and there are 16 rows;
// TestHeroFrameDimensions guards this. Runes:
//
//	K cap   b cap-band   S skin   e eye   m moustache
//	A apron a apron-shade W sleeve H hand  P leg  o shoe
var (
	heroIdle = Mask{
		"...KKKKKK...",
		"..KKKKKKKK..",
		"..bbbbbbbb..",
		"..SSSSSSSS..",
		"..SeSSSSeS..",
		"..SSSSSSSS..",
		"..SmmmmmmS..",
		"..SSSSSSSS..",
		".WWAAAAAAWW.",
		".WHAAAAAAHW.",
		"..AAAAAAAA..",
		"..AaAAAAaA..",
		"..AAAAAAAA..",
		"..PPP..PPP..",
		"..PPP..PPP..",
		"..ooo..ooo..",
	}
	heroRun1 = Mask{
		"...KKKKKK...",
		"..KKKKKKKK..",
		"..bbbbbbbb..",
		"..SSSSSSSS..",
		"..SeSSSSeS..",
		"..SSSSSSSS..",
		"..SmmmmmmS..",
		"..SSSSSSSS..",
		"HWWAAAAAAW..",
		".WHAAAAAAHW.",
		"..AAAAAAAA..",
		"..AaAAAAaA..",
		"..AAAAAAAA..",
		"...PPPP.....",
		"..PP..PPP...",
		"..oo...ooo..",
	}
	heroRun2 = Mask{
		"...KKKKKK...",
		"..KKKKKKKK..",
		"..bbbbbbbb..",
		"..SSSSSSSS..",
		"..SeSSSSeS..",
		"..SSSSSSSS..",
		"..SmmmmmmS..",
		"..SSSSSSSS..",
		"..WAAAAAAWWH",
		".WHAAAAAAHW.",
		"..AAAAAAAA..",
		"..AaAAAAaA..",
		"..AAAAAAAA..",
		".....PPPP...",
		"...PPP..PP..",
		"..ooo...oo..",
	}
	heroJump = Mask{
		"...KKKKKK...",
		"..KKKKKKKK..",
		"..bbbbbbbb..",
		"..SSSSSSSS..",
		"..SeSSSSeS..",
		"..SSSSSSSS..",
		"..SmmmmmmS..",
		"HWSSSSSSSS..",
		".WWAAAAAAWH.",
		"..HAAAAAAH..",
		"..AAAAAAAA..",
		"..AaAAAAaA..",
		"..AAAAAAAA..",
		"..PPP.PPP...",
		".PPP...PPP..",
		".ooo....oo..",
	}
)

// BuildCharacter rasterises the character frames with the given apron colours,
// which is how Ali (red apron) and Mehmet (green apron) are distinguished.
func BuildCharacter(apron, apronDark color.RGBA) CharacterSprites {
	pal := map[rune]color.RGBA{
		'K': White,
		'b': Red,
		'S': Skin,
		'e': Ink,
		'm': MeatDark,
		'A': apron,
		'a': apronDark,
		'W': Cream,
		'H': SkinDark,
		'P': Brown,
		'o': Ink,
	}
	return CharacterSprites{
		Idle: heroIdle.Build(pal),
		Run:  []*image.RGBA{heroRun1.Build(pal), heroRun2.Build(pal)},
		Jump: heroJump.Build(pal),
	}
}

// --- Item sprites -----------------------------------------------------------

var talerMask = Mask{
	"..gggg..",
	".gGGGGg.",
	"gGGwGGGg",
	"gGwGGGGg",
	"gGGGGGGg",
	"gGGGGwGg",
	".gGGGGg.",
	"..gggg..",
}

// Taler builds the golden collectible coin sprite.
func Taler() *image.RGBA {
	return talerMask.Build(map[rune]color.RGBA{
		'g': GoldDark,
		'G': Gold,
		'w': White,
	})
}

var spitMask = Mask{
	"..ss..",
	".mmmm.",
	"mMMMMm",
	"MMMMMM",
	"mMMMMm",
	"MMMMMM",
	"mMMMMm",
	"MMMMMM",
	"mMMMMm",
	".mmmm.",
	"..ss..",
	"..ss..",
}

// DoenerSpit builds the rotating döner-spit power-up sprite (replaces the
// classic mushroom).
func DoenerSpit() *image.RGBA {
	return spitMask.Build(map[rune]color.RGBA{
		's': Steel,
		'm': MeatDark,
		'M': Meat,
	})
}

var ayranMask = Mask{
	"..LL..",
	".LWWL.",
	".LWWL.",
	"LWWWWL",
	"LWbWWL",
	"LWWWWL",
	"LWWWWL",
	"LWbWWL",
	"LWWWWL",
	"LWWWWL",
	".LWWL.",
	"..LL..",
}

// Ayran builds the rare invincibility bottle sprite.
func Ayran() *image.RGBA {
	return ayranMask.Build(map[rune]color.RGBA{
		'L': NeonBlue,
		'W': White,
		'b': Red,
	})
}

var meatSliceMask = Mask{
	".mMm.",
	"mMMMm",
	"MMMMM",
	"mMMMm",
	".mMm.",
}

// MeatSlice builds the spinning meat-slice projectile thrown by Master Ali.
func MeatSlice() *image.RGBA {
	return meatSliceMask.Build(map[rune]color.RGBA{
		'm': MeatDark,
		'M': Meat,
	})
}
