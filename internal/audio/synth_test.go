package audio

import "testing"

func TestNoteLengthMatchesDuration(t *testing.T) {
	s := note(440, 0.5, 0.3, Square)
	want := int(0.5 * SampleRate)
	if len(s) != want {
		t.Fatalf("note length = %d want %d", len(s), want)
	}
}

func TestSamplesWithinRange(t *testing.T) {
	for _, w := range []Wave{Square, Triangle, Saw, Noise, Sine} {
		for _, v := range note(330, 0.05, 0.9, w) {
			if v > 1.0001 || v < -1.0001 {
				t.Fatalf("wave %d sample out of range: %v", w, v)
			}
		}
	}
}

func TestStereo16Format(t *testing.T) {
	mono := []float64{0, 1, -1, 2, -2} // includes values that must clamp
	b := stereo16(mono)
	if len(b) != len(mono)*4 {
		t.Fatalf("stereo16 length = %d want %d", len(b), len(mono)*4)
	}
	// Left and right channels of each frame must be identical (mono duplicated).
	for i := 0; i < len(mono); i++ {
		l := b[i*4 : i*4+2]
		r := b[i*4+2 : i*4+4]
		if l[0] != r[0] || l[1] != r[1] {
			t.Fatalf("frame %d channels differ", i)
		}
	}
}

func TestConcatAndSilence(t *testing.T) {
	a := note(440, 0.1, 0.3, Square)
	s := silence(0.1)
	c := concat(a, s)
	if len(c) != len(a)+len(s) {
		t.Fatalf("concat length wrong: %d", len(c))
	}
	for _, v := range s {
		if v != 0 {
			t.Fatal("silence must be zero")
		}
	}
}

func TestSFXAllBuildNonEmpty(t *testing.T) {
	for name, pcm := range buildSFX() {
		if len(pcm) == 0 {
			t.Errorf("SFX %q is empty", name)
		}
		if len(pcm)%4 != 0 {
			t.Errorf("SFX %q not frame-aligned (len %d)", name, len(pcm))
		}
	}
}

func TestMusicAllBuildNonEmpty(t *testing.T) {
	want := []string{MusicMenu, MusicLevel1, MusicLevel2, MusicLevel3, MusicBoss, MusicVictory}
	music := buildMusic()
	for _, name := range want {
		if len(music[name]) == 0 {
			t.Errorf("music %q is empty", name)
		}
	}
}

func TestResampleFasterIsShorter(t *testing.T) {
	pcm := note(440, 1.0, 0.5, Square)
	stereo := stereo16(pcm)
	fast := resampleStereo16(stereo, 1.25)
	if len(fast) >= len(stereo) {
		t.Fatalf("1.25x resample should be shorter: got %d vs %d", len(fast), len(stereo))
	}
	if len(fast)%4 != 0 {
		t.Fatalf("resampled output not frame-aligned: %d", len(fast))
	}
}

func TestResampleGuards(t *testing.T) {
	if got := resampleStereo16(nil, 1.25); got != nil {
		t.Error("empty input should return nil")
	}
	in := stereo16(note(440, 0.1, 0.5, Square))
	if got := resampleStereo16(in, 0); len(got) != len(in) {
		t.Error("non-positive speed should return input unchanged")
	}
}

func TestPitchReference(t *testing.T) {
	if f := pitch(noteA4); f != 440 {
		t.Fatalf("A4 should be 440Hz, got %v", f)
	}
	// One octave up doubles the frequency.
	if f := pitch(noteA5); f < 879 || f > 881 {
		t.Fatalf("A5 should be ~880Hz, got %v", f)
	}
}
