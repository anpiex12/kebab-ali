package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/anpiex12/kebab-ali/internal/audio"
	"github.com/anpiex12/kebab-ali/internal/entities"
	"github.com/anpiex12/kebab-ali/internal/level"
	"github.com/anpiex12/kebab-ali/internal/physics"
)

type playState int

const (
	psBanner playState = iota
	psPlaying
	psPaused
	psCleared
	psDying
)

// playScene runs one level: simulation, collisions, camera, HUD and the boss
// fight. It also implements entities.World for the simulation.
type playScene struct {
	levelIdx int
	lvl      *level.Level
	static   *ebiten.Image

	player      *entities.Player
	enemies     []entities.Enemy
	items       []*entities.Item
	projectiles []*entities.Projectile
	bosses      []entities.Boss

	bossActive bool
	bossTaunt  int

	camX, camY float64
	checkpoint level.Point
	cpTaken    int

	timeLeft float64
	elapsed  float64

	state     playState
	banner    int
	deadTimer int
	clearTimer int
	pauseSel  int
	tick      int
}

// newPlayScene builds level idx, carrying an existing player across levels (or
// creating a fresh one for the first level).
func newPlayScene(g *Game, idx int, carry *entities.Player) *playScene {
	l := level.All()[idx]
	s := &playScene{
		levelIdx:   idx,
		lvl:        l,
		static:     bakeStatic(l),
		checkpoint: l.PlayerStart,
		timeLeft:   float64(l.TimeLimit),
		state:      psBanner,
		banner:     120,
	}
	if carry != nil {
		s.player = carry
	} else {
		s.player = entities.NewPlayer(g.Character, l.PlayerStart.X, l.PlayerStart.Y)
	}
	s.player.PlaceAt(l.PlayerStart.X, l.PlayerStart.Y)

	for _, e := range l.Enemies {
		if en := spawnEnemy(e.Kind, e.X, e.Y); en != nil {
			s.enemies = append(s.enemies, en)
		}
	}
	for _, it := range l.Items {
		if it.Kind == "coin" {
			s.items = append(s.items, entities.NewTaler(it.X, it.Y))
		}
	}
	s.camY = (l.PixelHeight() - screenH) / 2
	return s
}

func spawnEnemy(kind string, x, y float64) entities.Enemy {
	switch kind {
	case "tomato":
		return entities.NewTomato(x, y)
	case "onion":
		return entities.NewOnion(x, y)
	case "peperoni":
		return entities.NewPeperoni(x, y)
	case "cucumber":
		return entities.NewCucumber(x, y)
	}
	return nil
}

// --- entities.World ---------------------------------------------------------

func (s *playScene) Solid(tx, ty int) bool    { return s.lvl.Solid(tx, ty) }
func (s *playScene) TileSize() int             { return s.lvl.TileSize() }
func (s *playScene) PlayerRect() physics.Rect { return s.player.Body }

// --- update -----------------------------------------------------------------

func (s *playScene) Enter(g *Game) { g.Audio.PlayMusic(s.lvl.Music) }

func (s *playScene) Update(g *Game) error {
	s.tick++
	if keyPressed(ebiten.KeyF) {
		g.ToggleFullscreen()
	}
	switch s.state {
	case psBanner:
		s.banner--
		if s.banner <= 0 || keyPressed(keysConfirm...) {
			s.state = psPlaying
		}
		return nil
	case psPaused:
		return s.updatePaused(g)
	case psCleared:
		s.clearTimer--
		if s.clearTimer <= 0 {
			s.advance(g)
		}
		return nil
	case psDying:
		s.deadTimer--
		if s.deadTimer <= 0 {
			s.respawnOrGameOver(g)
		}
		return nil
	}

	// psPlaying
	if keyPressed(ebiten.KeyEscape) {
		s.pauseSel = 0
		s.state = psPaused
		g.Audio.Play(audio.SFXPause)
		return nil
	}
	if keyPressed(ebiten.KeyM) {
		g.ToggleMute()
	}
	s.simulate(g)
	return nil
}

func (s *playScene) simulate(g *Game) {
	// Player.
	ev := s.player.Update(gameplayInput(), s)
	if ev.Jumped {
		g.Audio.Play(audio.SFXJump)
	}
	if ev.Threw {
		x := s.player.Body.CenterX()
		s.projectiles = append(s.projectiles, entities.NewMeatSlice(x, s.player.Body.CenterY(), float64(s.player.Facing)))
		g.Audio.Play(audio.SFXThrow)
	}
	if ev.Bonked {
		s.handleBonk(g, ev.BonkTX, ev.BonkTY)
	}
	g.Audio.SetTempoBoost(s.player.AyranActive())

	// Enemies.
	for _, e := range s.enemies {
		if e.Alive() {
			s.projectiles = append(s.projectiles, e.Update(s)...)
		}
	}
	// Items and projectiles.
	for _, it := range s.items {
		it.Update(s)
	}
	for _, pr := range s.projectiles {
		pr.Update(s)
	}

	// Boss trigger and update.
	s.updateBoss(g)

	// Collisions.
	s.collidePlayerEnemies(g)
	s.collidePlayerItems(g)
	s.collideProjectiles(g)
	s.collidePlayerBoss(g)
	s.collideHazards(g)
	s.checkpoints(g)

	// Cull dead objects.
	s.enemies = cullEnemies(s.enemies)
	s.items = cullItems(s.items)
	s.projectiles = cullProjectiles(s.projectiles)

	// Camera.
	s.updateCamera()

	// Timers / scoring clock.
	s.elapsed += 1.0 / 60
	s.timeLeft -= 1.0 / 60
	if s.timeLeft <= 0 {
		s.timeLeft = 0
		s.player.Die()
		g.Audio.Play(audio.SFXHit)
	}
	if s.bossTaunt > 0 {
		s.bossTaunt--
	}

	// Deaths and clears.
	if s.player.Dead() {
		s.state = psDying
		s.deadTimer = 70
		g.Audio.SetTempoBoost(false)
	}
	if s.bossActive && len(s.bosses) > 0 && s.allBossesDefeated() {
		s.onLevelCleared(g)
	}
}

func (s *playScene) handleBonk(g *Game, tx, ty int) {
	switch s.lvl.TileAt(tx, ty) {
	case level.BoxCoin:
		s.lvl.SetTile(tx, ty, level.BoxUsed)
		s.player.AddCoins(1)
		s.player.AddScore(100)
		g.Audio.Play(audio.SFXCoin)
	case level.BoxSpit:
		s.lvl.SetTile(tx, ty, level.BoxUsed)
		s.items = append(s.items, entities.NewSpit(float64(tx)*tileF, float64(ty-1)*tileF))
		g.Audio.Play(audio.SFXPower)
	case level.BoxAyran:
		s.lvl.SetTile(tx, ty, level.BoxUsed)
		s.items = append(s.items, entities.NewAyran(float64(tx)*tileF, float64(ty-1)*tileF))
		g.Audio.Play(audio.SFXPower)
	case level.Bread:
		if s.player.Power.Big() {
			s.lvl.SetTile(tx, ty, level.Empty)
			s.player.AddScore(50)
			g.Audio.Play(audio.SFXBreak)
		}
	}
}

func (s *playScene) updateBoss(g *Game) {
	if !s.bossActive && s.lvl.BossKind != "" && s.player.Body.CenterX() > s.lvl.Boss.X-150 {
		s.bosses = entities.NewBosses(s.lvl.BossKind, s.lvl.Boss.X, s.lvl.Boss.Y)
		s.bossActive = true
		s.bossTaunt = 200
		g.Audio.Play(audio.SFXFanfare)
		g.Audio.PlayMusic(audio.MusicBoss)
	}
	if s.bossActive {
		for _, b := range s.bosses {
			if b.Alive() {
				s.projectiles = append(s.projectiles, b.Update(s)...)
			}
		}
	}
}

func (s *playScene) collidePlayerEnemies(g *Game) {
	for _, e := range s.enemies {
		if !e.Alive() || e.Dying() || !s.player.Body.Intersects(e.Rect()) {
			continue
		}
		switch {
		case e.Stompable() && physics.StompedFrom(s.player.Body, e.Rect(), s.player.VY):
			e.Stomp()
			s.player.Bounce()
			s.player.AddScore(100)
			g.Audio.Play(audio.SFXStomp)
		case s.player.AyranActive():
			e.Damage()
			s.player.AddScore(100)
			g.Audio.Play(audio.SFXStomp)
		default:
			if !s.player.Invincible() {
				s.player.TakeHit()
				g.Audio.Play(audio.SFXHit)
			}
		}
	}
}

func (s *playScene) collidePlayerItems(g *Game) {
	for _, it := range s.items {
		if !it.Alive() || !s.player.Body.Intersects(it.Rect()) {
			continue
		}
		switch it.Kind {
		case entities.ItemTaler:
			if s.player.AddCoins(1) {
				g.Audio.Play(audio.SFXOneUp)
			} else {
				g.Audio.Play(audio.SFXCoin)
			}
			s.player.AddScore(100)
		case entities.ItemSpit:
			s.player.GivePower()
			s.player.AddScore(1000)
			g.Audio.Play(audio.SFXPower)
		case entities.ItemAyran:
			s.player.GiveAyran()
			s.player.AddScore(500)
			g.Audio.Play(audio.SFXPower)
		}
		it.Collect()
	}
}

func (s *playScene) collideProjectiles(g *Game) {
	for _, pr := range s.projectiles {
		if !pr.Alive() {
			continue
		}
		if pr.FromPlayer() {
			for _, e := range s.enemies {
				if e.Alive() && !e.Dying() && pr.Rect().Intersects(e.Rect()) {
					e.Damage()
					pr.Kill()
					s.player.AddScore(100)
					g.Audio.Play(audio.SFXStomp)
				}
			}
			for _, b := range s.bosses {
				if b.Alive() && pr.Rect().Intersects(b.Rect()) {
					b.HitByProjectile()
					pr.Kill()
				}
			}
		} else if pr.Rect().Intersects(s.player.Body) {
			pr.Kill()
			if !s.player.Invincible() {
				s.player.TakeHit()
				g.Audio.Play(audio.SFXHit)
			}
		}
	}
}

func (s *playScene) collidePlayerBoss(g *Game) {
	for _, b := range s.bosses {
		if !b.Alive() || !s.player.Body.Intersects(b.Rect()) {
			continue
		}
		switch {
		case b.Stompable() && physics.StompedFrom(s.player.Body, b.Rect(), s.player.VY):
			b.Stomp()
			s.player.Bounce()
			g.Audio.Play(audio.SFXStomp)
			if !b.Alive() {
				s.player.AddScore(2000)
				g.Audio.Play(audio.SFXFanfare)
			}
		case s.player.AyranActive():
			b.Stomp()
			s.player.Bounce()
		default:
			if s.player.Invincible() {
				continue
			}
			switch b.Contact() {
			case entities.ContactHurt:
				s.player.TakeHit()
				g.Audio.Play(audio.SFXHit)
			case entities.ContactEnroll:
				s.player.Enroll()
				g.Audio.Play(audio.SFXHit)
			}
		}
	}
}

func (s *playScene) collideHazards(g *Game) {
	b := s.player.Body
	if s.lvl.SauceOverlap(b.X, b.Y, b.W, b.H) || b.Y > s.lvl.PixelHeight()+24 {
		if !s.player.Dead() {
			s.player.Die()
			g.Audio.Play(audio.SFXSplash)
		}
	}
}

func (s *playScene) checkpoints(g *Game) {
	for s.cpTaken < len(s.lvl.Checkpoints) {
		cp := s.lvl.Checkpoints[s.cpTaken]
		if s.player.Body.CenterX() >= cp.X {
			s.checkpoint = cp
			s.cpTaken++
			g.Audio.Play(audio.SFXCoin)
		} else {
			break
		}
	}
}

func (s *playScene) allBossesDefeated() bool {
	for _, b := range s.bosses {
		if b.Alive() {
			return false
		}
	}
	return true
}

func (s *playScene) onLevelCleared(g *Game) {
	bonus := int(s.timeLeft) * 10
	s.player.AddScore(bonus + 1000)
	s.state = psCleared
	s.clearTimer = 160
	g.Audio.Play(audio.SFXFanfare)
	g.Audio.SetTempoBoost(false)
}

func (s *playScene) advance(g *Game) {
	if s.levelIdx+1 < len(level.All()) {
		g.SwitchTo(newPlayScene(g, s.levelIdx+1, s.player))
		return
	}
	g.SwitchTo(newVictoryScene(g, GameResult{Score: s.player.Score, Seconds: s.elapsed, Won: true}))
}

func (s *playScene) respawnOrGameOver(g *Game) {
	if s.player.Lives <= 0 {
		g.SwitchTo(newGameOverScene(GameResult{Score: s.player.Score, Seconds: s.elapsed, Won: false}))
		return
	}
	s.player.PlaceAt(s.checkpoint.X, s.checkpoint.Y)
	s.state = psPlaying
	g.Audio.SetTempoBoost(false)
	if s.bossActive {
		g.Audio.PlayMusic(audio.MusicBoss)
	} else {
		g.Audio.PlayMusic(s.lvl.Music)
	}
}

func (s *playScene) updateCamera() {
	s.camX = physics.Clamp(s.player.Body.CenterX()-screenW/2, 0, maxF(0, s.lvl.PixelWidth()-screenW))
}

func (s *playScene) updatePaused(g *Game) error {
	const pauseCount = 3
	switch {
	case keyPressed(keysDown...):
		s.pauseSel = (s.pauseSel + 1) % pauseCount
		g.Audio.Play(audio.SFXSelect)
	case keyPressed(keysUp...):
		s.pauseSel = (s.pauseSel + pauseCount - 1) % pauseCount
		g.Audio.Play(audio.SFXSelect)
	}
	if keyPressed(ebiten.KeyM) {
		g.ToggleMute()
	}
	if keyPressed(ebiten.KeyEscape) {
		s.state = psPlaying
		return nil
	}
	if keyPressed(keysConfirm...) {
		switch s.pauseSel {
		case 0: // resume
			s.state = psPlaying
		case 1: // restart level
			g.SwitchTo(newPlayScene(g, s.levelIdx, nil))
		case 2: // quit to menu
			g.SwitchTo(newMenuScene())
		}
	}
	return nil
}

// --- draw -------------------------------------------------------------------

func (s *playScene) Draw(g *Game, canvas *ebiten.Image) {
	fillGradient(canvas, s.lvl.Theme.SkyTop, s.lvl.Theme.SkyBottom)
	drawParallax(canvas, s.lvl.Theme, s.camX, s.tick)

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Translate(-s.camX, -s.camY)
	canvas.DrawImage(s.static, op)

	drawDynamicTiles(canvas, s.lvl, s.camX, s.camY, s.tick)
	for _, sg := range s.lvl.Signs {
		drawSign(canvas, g, sg, s.camX, s.camY)
	}
	for _, it := range s.items {
		if it.Alive() {
			drawItem(canvas, it, g.Assets, s.camX, s.camY)
		}
	}
	for _, e := range s.enemies {
		if e.Alive() {
			drawEnemy(canvas, e, s.camX, s.camY)
		}
	}
	for _, pr := range s.projectiles {
		if pr.Alive() {
			drawProjectile(canvas, pr, g.Assets, s.camX, s.camY)
		}
	}
	for _, b := range s.bosses {
		if b.Alive() {
			drawBoss(canvas, b, s.camX, s.camY)
		}
	}
	drawPlayer(canvas, s.player, g.Assets.Frames(s.player.Kind), s.camX, s.camY)

	s.drawHUD(g, canvas)
	s.drawOverlays(g, canvas)
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// slice culling helpers keep the live-object lists small.
func cullEnemies(in []entities.Enemy) []entities.Enemy {
	out := in[:0]
	for _, e := range in {
		if e.Alive() {
			out = append(out, e)
		}
	}
	return out
}

func cullItems(in []*entities.Item) []*entities.Item {
	out := in[:0]
	for _, it := range in {
		if it.Alive() {
			out = append(out, it)
		}
	}
	return out
}

func cullProjectiles(in []*entities.Projectile) []*entities.Projectile {
	out := in[:0]
	for _, pr := range in {
		if pr.Alive() {
			out = append(out, pr)
		}
	}
	return out
}
