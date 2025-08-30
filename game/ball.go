package game

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/assets"
)

const (
	ballSpeed          = 500.0
	ballGravity        = 100.0 // Gravity for hit balls
	hitSpeedMultiplier = 1.5   // How much the bat speed affects ball speed
	minDeflectionSpeed = 100.0 // Minimum speed after being hit
	curveStrength      = 30.0  // How much the ball curves
)

type Ball struct {
	position Vector
	velocity Vector
	sprite   *ebiten.Image
	active   bool
	hit      bool
	time     float64 // Time since ball was created (for curve calculation)
}

func NewBall() *Ball {
	sprite := assets.BallSprite
	bounds := sprite.Bounds()

	// Spawn ball at random height from horizontal middle line and above
	midHeight := 2 * float64(screenHeight) / 3
	startY := rand.Float64()*(midHeight-100) + 50 // From top to middle line only

	return &Ball{
		position: Vector{
			X: screenWidth + float64(bounds.Dx()), // Start off-screen right
			Y: startY,
		},
		velocity: Vector{
			X: -ballSpeed / 60.0, // Move left (60 FPS)
			Y: 0,                 // No vertical movement initially
		},
		sprite: sprite,
		active: true,
		hit:    false,
	}
}

func (b *Ball) Update() {
	if !b.active {
		return
	}

	// Update time
	b.time += 1.0 / 60.0 // Increment by frame time (assuming 60 FPS)

	// Apply gravity if ball was hit
	b.velocity.Y += ballGravity / 3600.0 // 60 FPS squared

	// Update position
	b.position = b.position.Add(b.velocity)

	// Check if ball is off screen
	bounds := b.sprite.Bounds()
	if b.position.Y > screenHeight+float64(bounds.Dy()) ||
		b.position.X < -float64(bounds.Dx()) ||
		b.position.X > screenWidth+float64(bounds.Dx()) ||
		b.position.Y < -float64(bounds.Dy()) {
		b.active = false
	}
}

func (b *Ball) Draw(screen *ebiten.Image) {
	if !b.active {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.position.X, b.position.Y)

	// Add slight trail effect for hit balls
	if b.hit {
		op.ColorScale.Scale(1.1, 1.1, 0.9, 1.0) // Slightly yellowish
	}

	screen.DrawImage(b.sprite, op)
}

func (b *Ball) Collider() Rect {
	bounds := b.sprite.Bounds()
	return NewRect(
		b.position.X,
		b.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}

func (b *Ball) IsActive() bool {
	return b.active
}

func (b *Ball) Hit(batAngle float64, swingVelocity float64) bool {
	if b.hit || !b.active {
		return false
	}

	b.hit = true

	// Calculate deflection based on bat angle and swing velocity
	// Bat angle: 0 = vertical, positive = clockwise

	// Base deflection direction - perpendicular to bat
	deflectionAngle := batAngle + math.Pi/2 // 90 degrees from bat angle

	// Add some randomness for more interesting gameplay
	deflectionAngle += (rand.Float64() - 0.5) * 0.3 // ±0.15 radians (~8.5 degrees)

	// Calculate hit speed based on swing velocity and current ball speed
	currentSpeed := math.Sqrt(b.velocity.X*b.velocity.X + b.velocity.Y*b.velocity.Y)
	hitSpeed := currentSpeed + math.Abs(swingVelocity)*hitSpeedMultiplier*60.0 // Convert to pixels per frame

	// Ensure minimum speed
	if hitSpeed < minDeflectionSpeed/60.0 {
		hitSpeed = minDeflectionSpeed / 60.0
	}

	// Set new velocity based on deflection angle and hit speed
	b.velocity = Vector{
		X: -math.Cos(deflectionAngle) * hitSpeed,
		Y: -math.Sin(deflectionAngle) * hitSpeed,
	}

	// Add some upward bias to make balls fly more realistically
	if b.velocity.Y > -50.0/60.0 { // If not already going up significantly
		b.velocity.Y -= 30.0 / 60.0 // Add upward velocity
	}
	return true
}
