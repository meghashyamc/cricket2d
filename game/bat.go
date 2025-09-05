package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/assets"
	"github.com/meghashyamc/cricket2d/geometry"
	"github.com/meghashyamc/cricket2d/logger"
)

const (
	maxSwingAngle          = math.Pi / 3 // 60 degrees maximum swing
	initialbatX            = 200
	initialbatY            = 350
	batMouseHistoryLimit   = 10  // Mouse history for calculating velocity
	batSpeedLimitingFactor = 0.3 // How fast the bat follows the mouse
)

type bat struct {
	position      geometry.Vector // Position of bat handle (pivot point)
	sprite        *ebiten.Image
	currentAngle  float64         // Current rotation angle (0 = vertical)
	previousAngle float64         // Previous frame angle for velocity calculation
	lastMousePos  geometry.Vector // Last mouse position
	mouseHistory  []geometry.Vector
	logger        logger.Logger
}

func newBat() *bat {
	sprite := assets.BatSprite

	position := geometry.Vector{
		X: initialbatX,
		Y: initialbatY,
	}

	bat := &bat{
		position:      position,
		sprite:        sprite,
		currentAngle:  0, // Start vertical
		previousAngle: 0,
		lastMousePos:  geometry.Vector{X: 0, Y: 0},
		mouseHistory:  make([]geometry.Vector, 0, batMouseHistoryLimit), // Keep last 10 positions for velocity calc
		logger:        logger.New(),
	}

	bat.logger.Debug("bat created", "position", bat.position, "max_swing_angle", maxSwingAngle)
	return bat
}

func (b *bat) update() {

	currentMousePosition := getCurrentMousePosition()
	// Update mouse history
	b.mouseHistory = append(b.mouseHistory, *currentMousePosition)
	if len(b.mouseHistory) > 10 {
		b.mouseHistory = b.mouseHistory[1:]
	}

	targetAngle := b.getNewTargetAngle(currentMousePosition)

	targetAngle = clampValue(targetAngle, -maxSwingAngle, maxSwingAngle)

	// Store previous angle for swing velocity calculation (needed when bat hits ball)
	b.previousAngle = b.currentAngle

	b.currentAngle += (targetAngle - b.currentAngle) * batSpeedLimitingFactor

	b.lastMousePos = *currentMousePosition
}

func (b *bat) draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	// Get sprite bounds for centering rotation
	bounds := b.sprite.Bounds()
	spriteWidth := float64(bounds.Dx())

	// Translate to handle position (top of bat), rotate, then translate back
	op.GeoM.Translate(-spriteWidth/2, 0) // Center horizontally, keep top at origin
	op.GeoM.Rotate(b.currentAngle)
	op.GeoM.Translate(b.position.X, b.position.Y)

	// Add slight glow effect when swinging fast
	if math.Abs(b.currentAngle-b.previousAngle) > 0.05 {
		intensity := float32(math.Min(1.2, 1.0+math.Abs(b.currentAngle-b.previousAngle)*5))
		op.ColorScale.Scale(intensity, intensity, intensity, 1.0)
	}

	screen.DrawImage(b.sprite, op)
}

func (b *bat) collidesWith(s *stumps) bool {

	return b.getBounds().Intersects(s.getBounds())
}

func (b *bat) getBounds() geometry.Rect {
	// Create a more accurate collision rectangle that represents the rotated bat
	bounds := b.sprite.Bounds()
	batWidth := float64(bounds.Dx())
	batHeight := float64(bounds.Dy())

	// Calculate the four corners of the rotated bat rectangle
	// Start with corners relative to the bat center
	halfWidth := batWidth / 2

	// Original corners (before rotation)
	corners := []geometry.Vector{
		{-halfWidth, 0},         // Top-left
		{halfWidth, 0},          // Top-right
		{halfWidth, batHeight},  // Bottom-right
		{-halfWidth, batHeight}, // Bottom-left
	}

	// Rotate each corner and translate to bat position
	rotatedCorners := make([]geometry.Vector, 4)
	for i, corner := range corners {
		// Rotate the corner
		rotatedX := corner.X*math.Cos(b.currentAngle) - corner.Y*math.Sin(b.currentAngle)
		rotatedY := corner.X*math.Sin(b.currentAngle) + corner.Y*math.Cos(b.currentAngle)

		// Translate to bat position
		rotatedCorners[i] = geometry.Vector{
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

	return geometry.NewRect(minX, minY, maxX-minX, maxY-minY)
}

// Performs precise collision detection between bat and ball
func (b *bat) checkCollision(ball *ball) bool {
	ballBounds := ball.getBounds()
	ballCenter := geometry.Vector{
		X: ballBounds.X + ballBounds.Width/2,
		Y: ballBounds.Y + ballBounds.Height/2,
	}
	ballRadius := math.Min(ballBounds.Width, ballBounds.Height) / 2

	// Get bat dimensions
	bounds := b.sprite.Bounds()
	batHeight := float64(bounds.Dy())
	batWidth := float64(bounds.Dx())

	// Calculate the main hitting area of the bat (central 95% of length)
	startOffset := batHeight * 0.05 // Start 10% from handle
	endOffset := batHeight * 0.95   // End 90% down the bat

	// Calculate start and end points of the bat hitting line
	batStart := geometry.Vector{
		X: b.position.X + math.Sin(-b.currentAngle)*startOffset,
		Y: b.position.Y + math.Cos(-b.currentAngle)*startOffset,
	}
	batEnd := geometry.Vector{
		X: b.position.X + math.Sin(-b.currentAngle)*endOffset,
		Y: b.position.Y + math.Cos(-b.currentAngle)*endOffset,
	}

	// Check distance from ball center to bat line
	distance := geometry.DistanceFromPointToLine(ballCenter, batStart, batEnd)

	if distance < 0 {
		return false
	}

	return distance <= (ballRadius + batWidth/2)

}

func (b *bat) getNewTargetAngle(currentMousePosition *geometry.Vector) float64 {
	deltaX := currentMousePosition.X - b.position.X
	deltaY := currentMousePosition.Y - b.position.Y

	// Calculate angle from vertical (0 = vertical, positive = clockwise)
	return math.Atan2(-deltaX, math.Abs(deltaY))
}
