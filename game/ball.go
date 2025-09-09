package game

import (
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/assets"
	"github.com/meghashyamc/cricket2d/geometry"
	"github.com/meghashyamc/cricket2d/logger"
)

const (
	minInitialballSpeed = float64(8)
	maxInitialballSpeed = float64(30)

	ballGravity            = 0.03 // Downward distance moved in a tick
	hitSpeedMultiplier     = 2    // How much the bat speed affects ball speed
	minDeflectionSpeed     = 1.67 // Minimum speed per tick after being hit (for bat body hits)
	minUpwardSpeedAfterHit = 0.083
)

type ball struct {
	position geometry.Vector
	velocity geometry.Vector
	sprite   *ebiten.Image
	active   bool
	isHit    bool
	logger   logger.Logger
}

func newBall(screenWidth float64, screenHeight float64) *ball {
	sprite := assets.BallSprite
	bounds := sprite.Bounds()

	startY := 2 * rand.Float64() * screenHeight / 3
	initialBallSpeedX := rand.Float64()*(maxInitialballSpeed-minInitialballSpeed) + minInitialballSpeed
	ball := &ball{
		position: geometry.Vector{
			X: screenWidth + float64(bounds.Dx()),
			Y: startY,
		},
		velocity: geometry.Vector{
			X: -initialBallSpeedX,
			Y: 0,
		},
		sprite: sprite,
		active: true,
		isHit:  false,
		logger: logger.New(),
	}

	ball.logger.Debug("ball created", "position", ball.position, "velocity", ball.velocity)
	return ball
}

func (b *ball) update(screenWidth float64, screenHeight float64) {
	if !b.active {
		return
	}

	b.velocity.Y += ballGravity

	b.position = b.position.Add(b.velocity)

	if b.isOffScreen(screenWidth, screenHeight) {
		b.logger.Debug("ball went off screen", "position", b.position)
		b.active = false
	}
}

func (b *ball) draw(screen *ebiten.Image) {
	if !b.active {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.position.X, b.position.Y)

	// Add slight trail effect for hit balls
	if b.isHit {
		op.ColorScale.Scale(1.1, 1.1, 0.9, 1.0) // Slightly yellowish
	}

	screen.DrawImage(b.sprite, op)
}

func (b *ball) hit(bat *bat, zone collisionZone) bool {
	if b.isHit || !b.active {
		return false
	}

	oldVelocity := b.velocity
	b.isHit = true

	normal := bat.getNormal()
	// Calculate reflected velocity vector
	reflected := b.velocity.Reflect(normal)

	// Calculate deflection angle from reflected vector
	deflectionAngle := math.Atan2(reflected.Y, reflected.X)

	// Calculate hit speed based on swing velocity and current ball speed
	currentSpeed := b.velocity.Magnitude()
	hitSpeed := currentSpeed + math.Abs(bat.currentAngle-bat.previousAngle)*hitSpeedMultiplier*60.0

	var (
		// Apply different physics based on collision zone

		// How randomly the ball gets deflected after a hit
		randomnessFactor,
		// Reduced power after hit
		speedModifier,
		// Add upward bias to make balls fly more realistically
		upwardBias float64
	)

	switch zone {
	case handleZone:

		randomnessFactor = 0.6 // ±0.3 radians (~17 degrees)
		speedModifier = 0.7
		upwardBias = 0.33
		// Ensure minimum speed is lower for handle hits
		hitSpeed = clampValue(hitSpeed, minDeflectionSpeed/2, hitSpeed)

	// default is BodyZone
	default:
		randomnessFactor = 0.3
		speedModifier = 1.0
		upwardBias = 0.5

		hitSpeed = clampValue(hitSpeed, minDeflectionSpeed, hitSpeed)
	}

	// Apply randomness and speed modifier
	deflectionAngle += (rand.Float64() - 0.5) * randomnessFactor
	hitSpeed *= speedModifier

	// Set new ball velocity based on deflection angle and hit speed
	b.velocity = geometry.Vector{
		X: -math.Cos(deflectionAngle) * hitSpeed,
		Y: -math.Sin(deflectionAngle) * hitSpeed,
	}

	// If not already going up significantly
	if b.velocity.Y > -minUpwardSpeedAfterHit {
		b.velocity.Y -= upwardBias
	}

	b.logger.Debug("ball hit physics calculated",
		"collision_zone", zone,
		"bat_angle", bat.currentAngle,
		"swing_angle", bat.currentAngle-bat.previousAngle,
		"deflection_angle", deflectionAngle,
		"hit_speed", hitSpeed,
		"speed_modifier", speedModifier,
		"randomness_factor", randomnessFactor,
		"old_velocity", oldVelocity,
		"new_velocity", b.velocity,
	)

	return true
}

func (b *ball) isOffScreen(screenWidth float64, screenHeight float64) bool {
	bounds := b.sprite.Bounds()
	return b.position.Y > screenHeight+float64(bounds.Dy()) ||
		b.position.X < -float64(bounds.Dx()) ||
		b.position.X > screenWidth+float64(bounds.Dx()) ||
		b.position.Y < -float64(bounds.Dy())
}

func (b *ball) getBounds() geometry.Rect {
	bounds := b.sprite.Bounds()
	return geometry.NewRect(
		b.position.X,
		b.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}

func (b *ball) collidesWith(s *stumps) bool {

	// Check if the ball is within the stumps bounds
	return b.getBounds().Intersects(s.getBounds())
}
