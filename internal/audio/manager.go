package audio

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// tempoBoostSpeed is how much faster music plays during the Ayran boost.
const tempoBoostSpeed = 1.25

// Manager owns the Ebitengine audio context and plays the pre-synthesised
// sound effects and looping music. All methods are nil-safe and no-op when
// muted, so callers never need to guard. A nil *Manager is a valid silent sink,
// which keeps head-less tests and code paths without audio simple.
type Manager struct {
	ctx   *audio.Context
	sfx   map[string][]byte
	music map[string][]byte

	muted bool
	boost bool

	active      []*audio.Player // live one-shot SFX players
	current     *audio.Player   // current looping music player
	currentName string
}

// NewManager creates the audio context (only valid to call once per process)
// and synthesises every effect and track up front.
func NewManager() *Manager {
	return &Manager{
		ctx:   audio.NewContext(SampleRate),
		sfx:   buildSFX(),
		music: buildMusic(),
	}
}

// Play triggers a one-shot sound effect. Overlapping plays are allowed because
// a fresh player is created each time.
func (m *Manager) Play(name string) {
	if m == nil || m.ctx == nil || m.muted {
		return
	}
	pcm, ok := m.sfx[name]
	if !ok {
		return
	}
	p := m.ctx.NewPlayerFromBytes(pcm)
	p.Play()
	m.active = append(m.active, p)
	m.pruneActive()
}

// pruneActive closes finished SFX players so they don't leak.
func (m *Manager) pruneActive() {
	kept := m.active[:0]
	for _, p := range m.active {
		if p.IsPlaying() {
			kept = append(kept, p)
		} else {
			_ = p.Close()
		}
	}
	m.active = kept
}

// PlayMusic starts (or restarts) a looping track. Requesting the track that is
// already playing is a no-op so it doesn't stutter on repeated calls.
func (m *Manager) PlayMusic(name string) {
	if m == nil || m.ctx == nil {
		return
	}
	if name == m.currentName && m.current != nil {
		return
	}
	m.currentName = name
	m.startCurrent()
}

// startCurrent (re)creates the player for currentName, honouring mute and the
// tempo boost.
func (m *Manager) startCurrent() {
	m.stopPlayer()
	pcm, ok := m.music[m.currentName]
	if !ok {
		return
	}
	if m.boost {
		pcm = resampleStereo16(pcm, tempoBoostSpeed)
	}
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
	p, err := m.ctx.NewPlayer(loop)
	if err != nil {
		return
	}
	m.current = p
	if !m.muted {
		p.Play()
	}
}

func (m *Manager) stopPlayer() {
	if m.current != nil {
		_ = m.current.Close()
		m.current = nil
	}
}

// StopMusic stops and forgets the current track.
func (m *Manager) StopMusic() {
	if m == nil {
		return
	}
	m.stopPlayer()
	m.currentName = ""
}

// SetTempoBoost speeds up (or restores) the current music for the Ayran power.
func (m *Manager) SetTempoBoost(on bool) {
	if m == nil || m.boost == on {
		return
	}
	m.boost = on
	if m.currentName != "" {
		m.startCurrent()
	}
}

// SetMuted mutes or unmutes all audio, pausing/resuming the music accordingly.
func (m *Manager) SetMuted(muted bool) {
	if m == nil {
		return
	}
	m.muted = muted
	if m.current == nil {
		return
	}
	if muted {
		m.current.Pause()
	} else {
		m.current.Play()
	}
}

// ToggleMute flips the mute state and returns the new value.
func (m *Manager) ToggleMute() bool {
	if m == nil {
		return true
	}
	m.SetMuted(!m.muted)
	return m.muted
}

// Muted reports whether audio is currently muted.
func (m *Manager) Muted() bool {
	return m != nil && m.muted
}
