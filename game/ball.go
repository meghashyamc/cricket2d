package game

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/assets"
	"github.com/meghashyamc/cricket2d/geometry"
	"github.com/meghashyamc/cricket2d/logger"
)

const (
	initialballSpeed = float64(8.3)

	ballGravity        = 0.03  // Downward distance moved in a tick
	hitSpeedMultiplier = 2     // How much the bat speed affects ball speed
	minDeflectionSpeed = 100.0 // Minimum speed after being hit
	curveStrength      = 30.0  // How much the ball curves
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
	ball := &ball{
		position: geometry.Vector{
			X: screenWidth + float64(bounds.Dx()),
			Y: startY,
		},
		velocity: geometry.Vector{
			X: -initialballSpeed,
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

func (b *ball) hit(bat *bat) bool {
	if b.isHit || !b.active {
		return false
	}

	oldVelocity := b.velocity
	b.isHit = true

	// bat angle: 0 = vertical, positive = clockwise
	normalAngle := bat.currentAngle + math.Pi/2
	normalX := math.Cos(normalAngle)
	normalY := math.Sin(normalAngle)

	// Calculate reflected velocity vector
	// Formula: reflected = incident - 2*(incident·normal)*normal
	dotProduct := b.velocity.X*normalX + b.velocity.Y*normalY
	reflectedX := b.velocity.X - 2*dotProduct*normalX
	reflectedY := b.velocity.Y - 2*dotProduct*normalY

	// Calculate deflection angle from reflected vector
	deflectionAngle := math.Atan2(reflectedY, reflectedX)

	// Add some randomness for more interesting gameplay
	deflectionAngle += (rand.Float64() - 0.5) * 0.3 // ±0.15 radians (~8.5 degrees)

	// Calculate hit speed based on swing velocity and current ball speed
	currentSpeed := math.Sqrt(b.velocity.X*b.velocity.X + b.velocity.Y*b.velocity.Y)
	hitSpeed := currentSpeed + math.Abs(bat.currentAngle-bat.previousAngle)*hitSpeedMultiplier*60.0 // Convert to pixels per frame

	// Ensure minimum speed
	if hitSpeed < minDeflectionSpeed/60.0 {
		hitSpeed = minDeflectionSpeed / 60.0
	}

	// Set new velocity based on deflection angle and hit speed
	b.velocity = geometry.Vector{
		X: -math.Cos(deflectionAngle) * hitSpeed,
		Y: -math.Sin(deflectionAngle) * hitSpeed,
	}

	// Add some upward bias to make balls fly more realistically
	if b.velocity.Y > -50.0/60.0 { // If not already going up significantly
		b.velocity.Y -= 30.0 / 60.0 // Add upward velocity
	}

	b.logger.Debug("ball hit physics calculated",
		"bat_angle", bat.currentAngle,
		"swing_angle", bat.currentAngle-bat.previousAngle,
		"deflection_angle", deflectionAngle,
		"hit_speed", hitSpeed,
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
