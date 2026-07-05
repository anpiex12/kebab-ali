package audio

// Sound-effect identifiers. These keys are what the rest of the game passes to
// Manager.Play.
const (
	SFXJump   = "jump"
	SFXCoin   = "coin"
	SFXPower  = "power"
	SFXHit    = "hit"
	SFXSplash = "splash" // falling into the sauce ("Platsch!")
	SFXStomp  = "stomp"
	SFXThrow  = "throw"
	SFXSelect = "select"
	SFXPause  = "pause"
	SFXOneUp  = "oneup"
	SFXFanfare = "fanfare" // boss appears / boss defeated
	SFXBreak  = "break"    // fladenbread block shatters
)

// Music-track identifiers.
const (
	MusicMenu    = "menu"
	MusicLevel1  = "level1"
	MusicLevel2  = "level2"
	MusicLevel3  = "level3"
	MusicBoss    = "boss"
	MusicVictory = "victory"
)

// buildSFX synthesises every sound effect into 16-bit stereo PCM.
func buildSFX() map[string][]byte {
	return map[string][]byte{
		// Springy upward chirp.
		SFXJump: stereo16(sweep(320, 720, 0.16, 0.35, Square)),
		// Two bright blips — the classic "ping" of grabbing a coin/taler.
		SFXCoin: stereo16(concat(
			note(pitch(noteB4), 0.05, 0.30, Square),
			note(pitch(noteE5), 0.12, 0.30, Square),
		)),
		// Rising arpeggio for the döner-spit power-up.
		SFXPower: stereo16(concat(
			note(pitch(noteC5), 0.06, 0.30, Square),
			note(pitch(noteE5), 0.06, 0.30, Square),
			note(pitch(noteG5), 0.06, 0.30, Square),
			note(pitch(noteA5), 0.14, 0.32, Square),
		)),
		// Descending "ow" on getting hit.
		SFXHit: stereo16(mix(
			sweep(440, 140, 0.28, 0.32, Saw),
			note(0, 0.28, 0.10, Noise),
		)),
		// Wet splat: a short noise burst that decays fast.
		SFXSplash: stereo16(tone(0.30, 0.40, Noise, func(t float64) float64 {
			return 200 * (1 - t) // pointless for noise but keeps signature
		})),
		// Low thud when stomping an enemy.
		SFXStomp: stereo16(sweep(220, 80, 0.12, 0.40, Square)),
		// Quick throw whoosh for the meat slice.
		SFXThrow: stereo16(sweep(600, 900, 0.08, 0.22, Square)),
		// Menu move/confirm blip.
		SFXSelect: stereo16(note(pitch(noteA5), 0.05, 0.25, Square)),
		// Pause toggle two-tone.
		SFXPause: stereo16(concat(
			note(pitch(noteE5), 0.05, 0.25, Square),
			note(pitch(noteC5), 0.07, 0.25, Square),
		)),
		// Extra-life jingle.
		SFXOneUp: stereo16(concat(
			note(pitch(noteE5), 0.07, 0.30, Triangle),
			note(pitch(noteG5), 0.07, 0.30, Triangle),
			note(pitch(noteC5+12), 0.07, 0.30, Triangle),
			note(pitch(noteA5), 0.16, 0.32, Triangle),
		)),
		// Short brassy boss fanfare.
		SFXFanfare: stereo16(concat(
			note(pitch(noteC5), 0.12, 0.34, Saw),
			note(pitch(noteC5), 0.12, 0.34, Saw),
			note(pitch(noteG5), 0.24, 0.34, Saw),
		)),
		// Crumbly break.
		SFXBreak: stereo16(note(0, 0.14, 0.35, Noise)),
	}
}

// step is one note (or rest) in a melody: a semitone offset from A4 and a
// length in beats. A rest is marked by the rest sentinel.
type step struct {
	n     int
	beats float64
}

const rest = -1000 // sentinel semitone meaning "silence"

// sequence renders a monophonic line at the given tempo, wave and volume.
func sequence(bpm float64, w Wave, vol float64, steps []step) []float64 {
	beat := 60.0 / bpm
	runs := make([][]float64, 0, len(steps))
	for _, s := range steps {
		d := s.beats * beat
		if s.n == rest {
			runs = append(runs, silence(d))
		} else {
			runs = append(runs, note(pitch(s.n), d, vol, w))
		}
	}
	return concat(runs...)
}

// buildMusic synthesises every looping music track. All melodies are original
// note sequences written for this game.
func buildMusic() map[string][]byte {
	return map[string][]byte{
		MusicMenu:    stereo16(menuTheme()),
		MusicLevel1:  stereo16(level1Theme()),
		MusicLevel2:  stereo16(level2Theme()),
		MusicLevel3:  stereo16(level3Theme()),
		MusicBoss:    stereo16(bossTheme()),
		MusicVictory: stereo16(victoryTheme()),
	}
}

func level1Theme() []float64 {
	bpm := 138.0
	melody := sequence(bpm, Square, 0.22, []step{
		{noteC5, 0.5}, {noteE5, 0.5}, {noteG5, 0.5}, {noteE5, 0.5},
		{noteA4, 0.5}, {noteC5, 0.5}, {noteE5, 0.5}, {noteC5, 0.5},
		{noteD5, 0.5}, {noteF5, 0.5}, {noteA5, 0.5}, {noteF5, 0.5},
		{noteG5, 1.0}, {noteE5, 0.5}, {noteC5, 0.5},
	})
	bass := sequence(bpm, Triangle, 0.28, []step{
		{noteC3, 1}, {noteG3, 1}, {noteA3, 1}, {noteE3, 1},
		{noteF3, 1}, {noteC3, 1}, {noteG3, 1}, {noteG3, 1},
	})
	return mix(melody, bass)
}

func level2Theme() []float64 {
	// A more exotic, bazaar-flavoured line (raised notes give it spice).
	bpm := 126.0
	melody := sequence(bpm, Square, 0.20, []step{
		{noteA4, 0.5}, {noteB4, 0.5}, {noteC5, 0.5}, {noteE5, 0.5},
		{noteF5, 0.5}, {noteE5, 0.5}, {noteC5, 0.5}, {noteB4, 0.5},
		{noteA4, 0.5}, {noteC5, 0.5}, {noteE5, 0.5}, {noteA5, 0.5},
		{noteG5, 1.0}, {noteE5, 0.5}, {noteA4, 0.5},
	})
	bass := sequence(bpm, Triangle, 0.28, []step{
		{noteA3, 1}, {noteA3, 1}, {noteF3, 1}, {noteF3, 1},
		{noteC3, 1}, {noteC3, 1}, {noteE3, 1}, {noteE3, 1},
	})
	return mix(melody, bass)
}

func level3Theme() []float64 {
	// Driving factory groove in a darker minor.
	bpm := 150.0
	melody := sequence(bpm, Saw, 0.16, []step{
		{noteA4, 0.25}, {noteA4, 0.25}, {noteC5, 0.5}, {noteA4, 0.25}, {noteA4, 0.25}, {noteG4, 0.5},
		{noteA4, 0.25}, {noteA4, 0.25}, {noteE5, 0.5}, {noteD5, 0.5}, {noteC5, 0.5},
		{rest, 0.5}, {noteE5, 0.5}, {noteA5, 0.5}, {noteG5, 0.5},
	})
	bass := sequence(bpm, Square, 0.22, []step{
		{noteA3, 0.5}, {noteA3, 0.5}, {noteA3, 0.5}, {noteA3, 0.5},
		{noteF3, 0.5}, {noteF3, 0.5}, {noteE3, 0.5}, {noteE3, 0.5},
		{noteA3, 0.5}, {noteA3, 0.5}, {noteG3, 0.5}, {noteG3, 0.5},
	})
	return mix(melody, bass)
}

func menuTheme() []float64 {
	bpm := 108.0
	melody := sequence(bpm, Triangle, 0.24, []step{
		{noteE5, 1}, {noteC5, 1}, {noteD5, 1}, {noteG4, 1},
		{noteA4, 1}, {noteC5, 1}, {noteB4, 1}, {rest, 1},
	})
	bass := sequence(bpm, Triangle, 0.24, []step{
		{noteC3, 2}, {noteG3, 2}, {noteA3, 2}, {noteE3, 2},
	})
	return mix(melody, bass)
}

func bossTheme() []float64 {
	bpm := 160.0
	melody := sequence(bpm, Saw, 0.18, []step{
		{noteE5, 0.5}, {noteF5, 0.5}, {noteE5, 0.5}, {noteD5, 0.5},
		{noteC5, 0.5}, {noteD5, 0.5}, {noteE5, 1.0},
		{noteA5, 0.5}, {noteG5, 0.5}, {noteF5, 0.5}, {noteE5, 0.5},
	})
	bass := sequence(bpm, Square, 0.24, []step{
		{noteA3, 0.5}, {noteA3, 0.5}, {noteA3, 0.5}, {noteA3, 0.5},
		{noteC3, 0.5}, {noteC3, 0.5}, {noteE3, 0.5}, {noteE3, 0.5},
	})
	return mix(melody, bass)
}

func victoryTheme() []float64 {
	bpm := 120.0
	return sequence(bpm, Triangle, 0.30, []step{
		{noteC5, 0.5}, {noteE5, 0.5}, {noteG5, 0.5}, {noteC5 + 12, 1.0},
		{noteA5, 0.5}, {noteG5, 0.5}, {noteC5 + 12, 1.5},
	})
}
