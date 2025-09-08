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

	// Draggable area constraints (relative to stumps position)
	batDragAreaRightOffset = 400 // How far right from stumps the bat can be dragged
	batDragAreaUpOffset    = 200 // How far up from stumps the bat can be dragged
	batDragAreaDownOffset  = 100 // How far down from stumps the bat can be dragged
)

type bat struct {
	position      geometry.Vector // Position of bat handle (pivot point)
	sprite        *ebiten.Image
	currentAngle  float64         // Current rotation angle (0 = vertical)
	previousAngle float64         // Previous frame angle for velocity calculation
	lastMousePos  geometry.Vector // Last mouse position
	mouseHistory  []geometry.Vector

	// Drag functionality fields
	isDragging     bool            // True when mouse button is held down for dragging
	dragOffset     geometry.Vector // Offset from bat position to mouse when drag starts
	dragStartAngle float64         // Angle when drag started (preserved during drag)

	logger logger.Logger
}

func newBat() *bat {
	sprite := assets.BatSprite

	position := geometry.Vector{
		X: initialbatX,
		Y: initialbatY,
	}

	bat := &bat{
		position:       position,
		sprite:         sprite,
		currentAngle:   -math.Pi / 3,
		previousAngle:  0,
		lastMousePos:   geometry.Vector{X: 0, Y: 0},
		mouseHistory:   make([]geometry.Vector, 0, batMouseHistoryLimit), // Keep last 10 positions for velocity calc
		isDragging:     false,
		dragOffset:     geometry.Vector{X: 0, Y: 0},
		dragStartAngle: 0,
		logger:         logger.New(),
	}

	bat.logger.Debug("bat created", "position", bat.position, "max_swing_angle", maxSwingAngle)
	return bat
}

// constrainToDraggableArea ensures the bat position stays within the allowed draggable area
func (b *bat) constrainToDraggableArea(position geometry.Vector, stumpsPos geometry.Vector) geometry.Vector {
	// Define boundaries relative to stumps position
	minX := stumpsPos.X
	maxX := stumpsPos.X + batDragAreaRightOffset
	minY := stumpsPos.Y - batDragAreaUpOffset
	maxY := stumpsPos.Y + batDragAreaDownOffset

	// Clamp the position within boundaries
	constrainedX := clampValue(position.X, minX, maxX)
	constrainedY := clampValue(position.Y, minY, maxY)

	return geometry.Vector{X: constrainedX, Y: constrainedY}
}

// startDrag initializes drag mode when mouse button is first pressed
func (b *bat) startDrag(mousePos geometry.Vector) {
	b.isDragging = true
	b.dragOffset = geometry.Vector{
		X: b.position.X - mousePos.X,
		Y: b.position.Y - mousePos.Y,
	}
	b.dragStartAngle = b.currentAngle
}

// updateDragPosition moves the bat during drag mode while preserving angle
func (b *bat) updateDragPosition(mousePos geometry.Vector, stumpsPos geometry.Vector) {
	newPosition := geometry.Vector{
		X: mousePos.X + b.dragOffset.X,
		Y: mousePos.Y + b.dragOffset.Y,
	}

	b.position = b.constrainToDraggableArea(newPosition, stumpsPos)

	// Keep the angle constant during drag
	b.currentAngle = b.dragStartAngle
}

func (b *bat) update(stumpsPos geometry.Vector) {

	currentMousePosition := getCurrentMousePosition()
	// Update mouse history
	b.mouseHistory = append(b.mouseHistory, *currentMousePosition)
	if len(b.mouseHistory) > batMouseHistoryLimit {
		b.mouseHistory = b.mouseHistory[1:]
	}

	// Check mouse button state for drag functionality
	isMousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	if isMousePressed && !b.isDragging {
		// Start dragging
		b.startDrag(*currentMousePosition)
	}

	if !isMousePressed && b.isDragging {
		// Stop dragging
		b.isDragging = false
	}

	// Store previous angle for swing velocity calculation (needed when bat hits ball)
	b.previousAngle = b.currentAngle
	b.lastMousePos = *currentMousePosition
	if b.isDragging {
		// In drag mode, move the bat while preserving angle
		b.updateDragPosition(*currentMousePosition, stumpsPos)
		return
	}

	// In normal mode: adjust bat angle based on mouse position
	targetAngle := b.getNewTargetAngle(currentMousePosition)
	targetAngle = clampValue(targetAngle, -maxSwingAngle, maxSwingAngle)
	b.currentAngle += (targetAngle - b.currentAngle) * batSpeedLimitingFactor

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
