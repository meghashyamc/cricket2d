package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/assets"
	"github.com/meghashyamc/cricket2d/logger"
)

type Stumps struct {
	position  Vector
	sprite    *ebiten.Image
	outSprite *ebiten.Image
	fallen    bool
	logger    logger.Logger
}

func NewStumps() *Stumps {
	sprite := assets.StumpsSprite
	bounds := sprite.Bounds()

	// Position stumps on the left side of screen, closer to the bottom
	pos := Vector{
		X: 30,                                                // Left side with some margin
		Y: float64(screenHeight) - float64(bounds.Dy()) - 80, // 80 pixels from bottom
	}

	stumps := &Stumps{
		position:  pos,
		sprite:    sprite,
		outSprite: assets.StumpsOutSprite,
		fallen:    false,
		logger:    logger.New(),
	}

	stumps.logger.Debug("stumps created", "position", pos, "bounds", bounds)
	return stumps
}

func (s *Stumps) Update() {
	// Stumps don't need updating unless they fall
}

func (s *Stumps) Draw(screen *ebiten.Image) {
	var currentSprite *ebiten.Image
	if s.fallen && s.outSprite != nil {
		currentSprite = s.outSprite
	} else {
		currentSprite = s.sprite
	}

	if currentSprite == nil {
		return
	}

	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(s.position.X, s.position.Y)
	screen.DrawImage(currentSprite, options)
}

func (s *Stumps) Collider() Rect {
	bounds := s.sprite.Bounds()
	return NewRect(
		s.position.X,
		s.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}

func (s *Stumps) CheckCollision(ball *Ball, bat *Bat) bool {
	if s.fallen {
		return false
	}

	var ballCollided, batCollided bool
	if ball != nil && ball.IsActive() {
		ballCollided = ball.Collider().Intersects(s.Collider())
	}

	if bat != nil {
		batCollided = bat.Collider().Intersects(s.Collider())
	}

	return ballCollided || batCollided
}
func (s *Stumps) Fall() {
	s.logger.Debug("stumps falling")
	s.fallen = true
}

func (s *Stumps) IsFallen() bool {
	return s.fallen
}

func (s *Stumps) Reset() {
	s.logger.Debug("stumps reset")
	s.fallen = false
}
