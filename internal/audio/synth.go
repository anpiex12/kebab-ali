// Package audio synthesises all of the game's sound effects and music in code —
// no sampled or copyrighted audio files are used. The synthesis half (this
// file) is pure and produces raw 16-bit stereo PCM []byte, so it can be
// unit-tested without ever opening an audio device. The Manager half binds it
// to Ebitengine's audio context.
package audio

import (
	"encoding/binary"
	"math"
)

// SampleRate is the PCM sample rate used throughout (CD quality).
const SampleRate = 44100

// Wave selects an oscillator shape.
type Wave int

const (
	// Square is a classic chiptune square wave (hollow, buzzy).
	Square Wave = iota
	// Triangle is soft and flute-like, good for bass and melodies.
	Triangle
	// Saw is bright and harsh.
	Saw
	// Noise is white noise, used for percussion and splashes.
	Noise
	// Sine is a pure tone.
	Sine
)

// rng is a tiny deterministic PRNG so noise is reproducible in tests (the
// standard library's rand is avoided to keep synthesis pure and repeatable).
type rng struct{ s uint32 }

func (r *rng) next() float64 {
	r.s = r.s*1664525 + 1013904223
	return float64(r.s)/float64(math.MaxUint32)*2 - 1
}

// oscillator returns the sample value in [-1,1] for a waveform at phase p
// (in cycles; only the fractional part matters) using noise source n.
func oscillator(w Wave, p float64, n *rng) float64 {
	frac := p - math.Floor(p)
	switch w {
	case Square:
		if frac < 0.5 {
			return 1
		}
		return -1
	case Triangle:
		if frac < 0.5 {
			return 4*frac - 1
		}
		return 3 - 4*frac
	case Saw:
		return 2*frac - 1
	case Noise:
		return n.next()
	default: // Sine
		return math.Sin(2 * math.Pi * frac)
	}
}

// tone renders a mono tone of the given duration whose frequency at normalized
// time t in [0,1) is freqAt(t). A short attack and release envelope avoids
// clicks. vol is the peak amplitude in [0,1].
func tone(dur, vol float64, w Wave, freqAt func(t float64) float64) []float64 {
	n := int(dur * SampleRate)
	if n <= 0 {
		return nil
	}
	out := make([]float64, n)
	noise := &rng{s: 0x1234abcd}
	phase := 0.0
	sr := float64(SampleRate)
	attack := int(0.005 * sr)
	release := int(0.02 * sr)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n)
		f := freqAt(t)
		phase += f / SampleRate
		env := 1.0
		if i < attack {
			env = float64(i) / float64(attack)
		}
		if i > n-release {
			env = float64(n-i) / float64(release)
		}
		out[i] = oscillator(w, phase, noise) * vol * env
	}
	return out
}

// note renders a constant-frequency tone.
func note(freq, dur, vol float64, w Wave) []float64 {
	return tone(dur, vol, w, func(float64) float64 { return freq })
}

// sweep renders a tone that glides linearly from f0 to f1.
func sweep(f0, f1, dur, vol float64, w Wave) []float64 {
	return tone(dur, vol, w, func(t float64) float64 { return f0 + (f1-f0)*t })
}

// silence returns dur seconds of quiet mono samples.
func silence(dur float64) []float64 {
	return make([]float64, int(dur*SampleRate))
}

// concat joins mono sample runs end to end.
func concat(runs ...[]float64) []float64 {
	total := 0
	for _, r := range runs {
		total += len(r)
	}
	out := make([]float64, 0, total)
	for _, r := range runs {
		out = append(out, r...)
	}
	return out
}

// mix sums several mono tracks sample-by-sample (padding shorter ones with
// silence). The result is not yet clamped.
func mix(tracks ...[]float64) []float64 {
	n := 0
	for _, t := range tracks {
		if len(t) > n {
			n = len(t)
		}
	}
	out := make([]float64, n)
	for _, t := range tracks {
		for i, v := range t {
			out[i] += v
		}
	}
	return out
}

// stereo16 converts mono float samples in roughly [-1,1] to interleaved 16-bit
// little-endian stereo PCM bytes, hard-clamping to avoid wrap-around.
func stereo16(mono []float64) []byte {
	buf := make([]byte, len(mono)*4)
	for i, v := range mono {
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		s := int16(v * 32767)
		binary.LittleEndian.PutUint16(buf[i*4:], uint16(s))
		binary.LittleEndian.PutUint16(buf[i*4+2:], uint16(s))
	}
	return buf
}

// resampleStereo16 changes playback speed (>1 = faster & higher-pitched) by
// linearly interpolating 16-bit stereo frames. Used for the Ayran tempo boost.
func resampleStereo16(pcm []byte, speed float64) []byte {
	frames := len(pcm) / 4
	if speed <= 0 || frames == 0 {
		return pcm
	}
	outFrames := int(float64(frames) / speed)
	out := make([]byte, outFrames*4)
	for i := 0; i < outFrames; i++ {
		src := float64(i) * speed
		i0 := int(src)
		frac := src - float64(i0)
		i1 := i0 + 1
		if i1 >= frames {
			i1 = frames - 1
		}
		for ch := 0; ch < 2; ch++ {
			s0 := int16(binary.LittleEndian.Uint16(pcm[i0*4+ch*2:]))
			s1 := int16(binary.LittleEndian.Uint16(pcm[i1*4+ch*2:]))
			v := int16(float64(s0)*(1-frac) + float64(s1)*frac)
			binary.LittleEndian.PutUint16(out[i*4+ch*2:], uint16(v))
		}
	}
	return out
}

// pitch converts a scientific note (semitones from A4=440Hz) to a frequency.
func pitch(semitonesFromA4 int) float64 {
	return 440 * math.Pow(2, float64(semitonesFromA4)/12)
}

// Note constants as semitone offsets from A4, spanning the range the tunes use.
const (
	noteC4  = -9
	noteD4  = -7
	noteE4  = -5
	noteF4  = -4
	noteG4  = -2
	noteA4  = 0
	noteB4  = 2
	noteC5  = 3
	noteD5  = 5
	noteE5  = 7
	noteF5  = 8
	noteG5  = 10
	noteA5  = 12
	noteC3  = -21
	noteE3  = -17
	noteG3  = -14
	noteA3  = -12
	noteF3  = -16
	noteD3  = -19
)
