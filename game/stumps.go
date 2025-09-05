package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/assets"
	"github.com/meghashyamc/cricket2d/geometry"
	"github.com/meghashyamc/cricket2d/logger"
)

const (
	initialstumpsX        = 30
	initialstumpsYPercent = 0.9 // Percentage of screen height (starting from top) where stumps are placed
)

type stumps struct {
	position  geometry.Vector
	sprite    *ebiten.Image
	outSprite *ebiten.Image
	isFallen  bool
	logger    logger.Logger
}

func newStumps(screenHeight float64) *stumps {
	sprite := assets.StumpsSprite
	bounds := sprite.Bounds()

	// Position stumps on the left side of screen, closer to the bottom
	pos := geometry.Vector{
		X: initialstumpsX,
		Y: initialstumpsYPercent * (screenHeight - float64(bounds.Dy())),
	}

	stumps := &stumps{
		position:  pos,
		sprite:    sprite,
		outSprite: assets.StumpsOutSprite,
		isFallen:  false,
		logger:    logger.New(),
	}

	stumps.logger.Debug("stumps created", "position", pos, "bounds", bounds)
	return stumps
}

func (s *stumps) draw(screen *ebiten.Image) {
	var currentSprite *ebiten.Image
	if s.isFallen && s.outSprite != nil {
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

func (s *stumps) checkCollision(ball *ball, bat *bat) bool {
	if s.isFallen {
		return false
	}

	var ballCollided, batCollided bool
	if ball != nil && ball.active {
		ballCollided = ball.collidesWith(s)
	}

	if bat != nil {
		batCollided = bat.collidesWith(s)
	}

	return ballCollided || batCollided
}
func (s *stumps) fall() {
	s.logger.Debug("stumps falling")
	s.isFallen = true
}

func (s *stumps) reset() {
	s.logger.Debug("stumps reset")
	s.isFallen = false
}

func (s *stumps) getBounds() geometry.Rect {
	bounds := s.sprite.Bounds()
	return geometry.NewRect(
		s.position.X,
		s.position.Y,
		float64(bounds.Dx()),
		float64(bounds.Dy()),
	)
}
