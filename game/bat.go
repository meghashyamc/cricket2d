package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/assets"
	"github.com/meghashyamc/cricket2d/logger"
)

const (
	maxSwingAngle = math.Pi / 3 // 60 degrees maximum swing
	batLength     = 100.0       // Length of the bat for physics calculations
)

type Bat struct {
	position      Vector // Position of bat handle (pivot point)
	sprite        *ebiten.Image
	currentAngle  float64  // Current rotation angle (0 = vertical)
	previousAngle float64  // Previous frame angle for velocity calculation
	swingVelocity float64  // Angular velocity
	lastMousePos  Vector   // Last mouse position
	mouseHistory  []Vector // Mouse history for calculating velocity
	logger        logger.Logger
}

func NewBat() *Bat {
	sprite := assets.BatSprite
	bounds := sprite.Bounds()

	// Position bat in the middle-left area (between stumps and right side)
	// Align with stumps height
	pos := Vector{
		X: 200,                                               // Between stumps (at ~50) and ball spawn area
		Y: float64(screenHeight) - float64(bounds.Dy()) - 80, // Same height as stumps
	}

	bat := &Bat{
		position:      pos,
		sprite:        sprite,
		currentAngle:  0, // Start vertical
		previousAngle: 0,
		swingVelocity: 0,
		lastMousePos:  Vector{0, 0},
		mouseHistory:  make([]Vector, 0, 10), // Keep last 10 positions for velocity calc
		logger:        logger.New(),
	}
	
	bat.logger.Debug("bat created", "position", bat.position, "maxSwingAngle", maxSwingAngle)
	return bat
}

func (b *Bat) Update() {
	// Get current mouse position
	mouseX, mouseY := ebiten.CursorPosition()
	currentMousePos := Vector{X: float64(mouseX), Y: float64(mouseY)}

	// Update mouse history
	b.mouseHistory = append(b.mouseHistory, currentMousePos)
	if len(b.mouseHistory) > 10 {
		b.mouseHistory = b.mouseHistory[1:]
	}

	// Calculate desired angle based on mouse position relative to bat handle
	deltaX := currentMousePos.X - b.position.X
	deltaY := currentMousePos.Y - b.position.Y

	// Calculate angle from vertical (0 = vertical, positive = clockwise)
	targetAngle := math.Atan2(-deltaX, math.Abs(deltaY)) // Note: -deltaY because Y increases downward

	// Clamp angle to maximum swing range
	if targetAngle > maxSwingAngle {
		targetAngle = maxSwingAngle
	} else if targetAngle < -maxSwingAngle {
		targetAngle = -maxSwingAngle
	}

	// Store previous angle for velocity calculation
	b.previousAngle = b.currentAngle

	// Smoothly interpolate to target angle (makes bat movement feel more natural)
	angleSpeed := 0.3 // How fast the bat follows the mouse
	b.currentAngle += (targetAngle - b.currentAngle) * angleSpeed

	// Calculate swing velocity (angular velocity)
	b.swingVelocity = b.currentAngle - b.previousAngle

	// Log significant swing movements for debugging
	if math.Abs(b.swingVelocity) > 0.05 { // Only log significant swings
		b.logger.Debug("significant bat swing", 
			"currentAngle", b.currentAngle,
			"previousAngle", b.previousAngle,
			"swingVelocity", b.swingVelocity,
			"mousePos", currentMousePos,
		)
	}

	b.lastMousePos = currentMousePos
}

func (b *Bat) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	// Get sprite bounds for centering rotation
	bounds := b.sprite.Bounds()
	spriteWidth := float64(bounds.Dx())

	// Translate to handle position (top of bat), rotate, then translate back
	op.GeoM.Translate(-spriteWidth/2, 0) // Center horizontally, keep top at origin
	op.GeoM.Rotate(b.currentAngle)
	op.GeoM.Translate(b.position.X, b.position.Y)

	// Add slight glow effect when swinging fast
	if math.Abs(b.swingVelocity) > 0.05 {
		intensity := float32(math.Min(1.2, 1.0+math.Abs(b.swingVelocity)*5))
		op.ColorScale.Scale(intensity, intensity, intensity, 1.0)
	}

	screen.DrawImage(b.sprite, op)
}

func (b *Bat) Collider() Rect {
	// Create a more accurate collision rectangle that represents the rotated bat
	bounds := b.sprite.Bounds()
	batWidth := float64(bounds.Dx())
	batHeight := float64(bounds.Dy())

	// Calculate the four corners of the rotated bat rectangle
	// Start with corners relative to the bat center
	halfWidth := batWidth / 2

	// Original corners (before rotation)
	corners := []Vector{
		{-halfWidth, 0},         // Top-left
		{halfWidth, 0},          // Top-right
		{halfWidth, batHeight},  // Bottom-right
		{-halfWidth, batHeight}, // Bottom-left
	}

	// Rotate each corner and translate to bat position
	rotatedCorners := make([]Vector, 4)
	for i, corner := range corners {
		// Rotate the corner
		rotatedX := corner.X*math.Cos(b.currentAngle) - corner.Y*math.Sin(b.currentAngle)
		rotatedY := corner.X*math.Sin(b.currentAngle) + corner.Y*math.Cos(b.currentAngle)

		// Translate to bat position
		rotatedCorners[i] = Vector{
			X: b.position.X + rotatedX,
			Y: b.position.Y + rotatedY,
		}
	}

	// Find the bounding box of the rotated bat
	minX := rotatedCorners[0].X
	maxX := rotatedCorners[0].X
	minY := rotatedCorners[0].Y
	maxY := rotatedCorners[0].Y

	for _, corner := range rotatedCorners[1:] {
		if corner.X < minX {
			minX = corner.X
		}
		if corner.X > maxX {
			maxX = corner.X
		}
		if corner.Y < minY {
			minY = corner.Y
		}
		if corner.Y > maxY {
			maxY = corner.Y
		}
	}

	return NewRect(minX, minY, maxX-minX, maxY-minY)
}

// CheckBallCollision performs precise collision detection between bat and ball
func (b *Bat) CheckBallCollision(ball *Ball) bool {
	ballRect := ball.Collider()
	ballCenter := Vector{
		X: ballRect.X + ballRect.Width/2,
		Y: ballRect.Y + ballRect.Height/2,
	}
	ballRadius := math.Min(ballRect.Width, ballRect.Height) / 2

	// Get bat dimensions
	bounds := b.sprite.Bounds()
	batHeight := float64(bounds.Dy())
	batWidth := float64(bounds.Dx())

	// Calculate the main hitting area of the bat (central 95% of length)
	startOffset := batHeight * 0.05 // Start 10% from handle
	endOffset := batHeight * 0.95   // End 90% down the bat

	// Calculate start and end points of the bat hitting line
	batStart := Vector{
		X: b.position.X + math.Sin(-b.currentAngle)*startOffset,
		Y: b.position.Y + math.Cos(-b.currentAngle)*startOffset,
	}
	batEnd := Vector{
		X: b.position.X + math.Sin(-b.currentAngle)*endOffset,
		Y: b.position.Y + math.Cos(-b.currentAngle)*endOffset,
	}

	// Check distance from ball center to bat line
	distance := b.distancePointToLine(ballCenter, batStart, batEnd)

	if distance < 0 {
		return false
	}

	collision := distance <= (ballRadius + batWidth/2)
	
	if collision {
		b.logger.Debug("collision detection details", 
			"ballCenter", ballCenter,
			"ballRadius", ballRadius,
			"batStart", batStart,
			"batEnd", batEnd,
			"distance", distance,
			"threshold", ballRadius + batWidth/2,
		)
	}
	
	return collision
}

// distancePointToLine calculates the shortest distance from a point to a line segment
func (b *Bat) distancePointToLine(point, lineStart, lineEnd Vector) float64 {
	// Vector from line start to end
	lineVec := Vector{X: lineEnd.X - lineStart.X, Y: lineEnd.Y - lineStart.Y}
	// Vector from line start to point
	pointVec := Vector{X: point.X - lineStart.X, Y: point.Y - lineStart.Y}
	return pointVec.Magnitude() * math.Sin(pointVec.AngleTo(lineVec))
}

func (b *Bat) Position() Vector {
	return b.position
}

func (b *Bat) GetBatAngle() float64 {
	return b.currentAngle
}

func (b *Bat) GetSwingVelocity() float64 {
	return b.swingVelocity
}

// Calculate the velocity of the bat tip for more realistic ball deflection
func (b *Bat) GetBatTipVelocity() Vector {
	// Calculate where the bat tip is
	bounds := b.sprite.Bounds()
	batHeight := float64(bounds.Dy())

	// Velocity is perpendicular to the bat and proportional to angular velocity
	velocityMagnitude := math.Abs(b.swingVelocity) * batHeight * 0.8
	velocityX := -math.Cos(b.currentAngle) * velocityMagnitude
	velocityY := math.Sin(b.currentAngle) * velocityMagnitude

	// Apply sign based on swing direction
	if b.swingVelocity < 0 {
		velocityX = -velocityX
		velocityY = -velocityY
	}

	return Vector{X: velocityX, Y: velocityY}
}
