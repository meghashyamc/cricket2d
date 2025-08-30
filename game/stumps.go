package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/assets"
	"github.com/meghashyamc/cricket2d/logger"
)

type Stumps struct {
	position Vector
	sprite   *ebiten.Image
	fallen   bool
	logger   logger.Logger
}

func NewStumps() *Stumps {
	sprite := assets.StumpsSprite
	bounds := sprite.Bounds()

	// Position stumps on the left side of screen, closer to the bottom
	pos := Vector{
		X: 50, // Left side with some margin
		Y: float64(screenHeight) - float64(bounds.Dy()) - 80, // 80 pixels from bottom
	}

	stumps := &Stumps{
		position: pos,
		sprite:   sprite,
		fallen:   false,
		logger:   logger.New(),
	}
	
	stumps.logger.Debug("stumps created", "position", pos, "bounds", bounds)
	return stumps
}

func (s *Stumps) Update() {
	// Stumps don't need updating unless they fall
}

func (s *Stumps) Draw(screen *ebiten.Image) {
	if s.sprite == nil {
		return
	}

	options := &ebiten.DrawImageOptions{}

	if s.fallen {
		// Rotate stumps to show they've fallen
		bounds := s.sprite.Bounds()
		options.GeoM.Translate(-float64(bounds.Dx())/2, -float64(bounds.Dy())/2)
		options.GeoM.Rotate(1.57) // 90 degrees in radians
		options.GeoM.Translate(float64(bounds.Dx())/2, float64(bounds.Dy())/2)
		options.ColorScale.Scale(0.7, 0.7, 0.7, 1.0) // Darken to show they're down
	}

	options.GeoM.Translate(s.position.X, s.position.Y)
	screen.DrawImage(s.sprite, options)
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

func (s *Stumps) CheckCollision(ball *Ball) bool {
	if s.fallen || !ball.IsActive() {
		return false
	}

	return ball.Collider().Intersects(s.Collider())
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
