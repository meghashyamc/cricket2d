package game

import (
	"fmt"
	"image/color"
	"strings"
	"time"
	"unicode"

	"github.com/meghashyamc/cricket2d/config"
	"github.com/meghashyamc/cricket2d/logger"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/meghashyamc/cricket2d/assets"
)

type GameState int

const (
	GameStatePlaying GameState = iota
	GameStateGameOver
	GameStateNameInput
)

type Game struct {
	cfg              *config.Config
	bat              *Bat
	balls            map[*Ball]struct{}
	stumps           *Stumps
	ballSpawnTimer   *time.Ticker
	score            int
	state            GameState
	highScoreManager *HighScoreManager
	logger           logger.Logger
	userMessage      string
}

func NewGame(cfg *config.Config) (*Game, error) {
	highScoreManager, err := NewHighScoreManager(cfg)
	if err != nil {
		return nil, err
	}

	g := &Game{
		cfg:              cfg,
		bat:              NewBat(),
		balls:            make(map[*Ball]struct{}),
		stumps:           NewStumps(),
		ballSpawnTimer:   time.NewTicker(time.Duration(cfg.GetBallSpawnTime()) * time.Second),
		score:            0,
		state:            GameStatePlaying,
		highScoreManager: highScoreManager,
		logger:           logger.New(),
		userMessage:      "",
	}

	g.logger.Info("game initialized", "ball_spawn_time_seconds", cfg.GetBallSpawnTime())
	return g, nil
}

func (g *Game) Run() error {
	g.logger.Info("starting game")
	g.setupWindow()

	// Running the game calls Update() on every 'tick'
	return ebiten.RunGame(g)
}

func (g *Game) setupWindow() {
	ebiten.SetWindowSize(g.cfg.GetWindowWidth(), g.cfg.GetWindowHeight())
	ebiten.SetWindowTitle(g.cfg.GetWindowTitle())
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
}

func (g *Game) Update() error {

	switch g.state {
	case GameStatePlaying:
		return g.updatePlaying()
	case GameStateGameOver:
		return g.updateGameReset()
	case GameStateNameInput:
		return g.updateNameInput()
	}
	return nil
}

func (g *Game) updatePlaying() error {
	g.bat.Update()

	select {
	// New balls should come in at regular intervals
	case <-g.ballSpawnTimer.C:
		newBall := NewBall()
		g.balls[newBall] = struct{}{}
		g.logger.Debug("new ball spawned", "ballCount", len(g.balls), "ballPosition", newBall.position)

	// On every tick, check if the wicket has been hit by the bat
	default:
		if g.stumps.CheckCollision(nil, g.bat) {
			g.logger.Debug("bat collided with stumps", "score", g.score)
			g.stumps.Fall()
			g.endGame(gameEndMessageHitWicket)
			return nil
		}
	}

	g.updateBalls()

	return nil
}

func (g *Game) updateBalls() {
	ballsToDeactivate := make([]*Ball, 0)

	for ball := range g.balls {
		ball.Update()

		if !ball.IsActive() {
			// Remove inactive balls
			ballsToDeactivate = append(ballsToDeactivate, ball)
			continue
		}

		// Check collision with bat
		if g.bat.CheckBallCollision(ball) {
			if ball.Hit(g.bat.GetBatAngle(), g.bat.GetSwingVelocity()) {
				g.score++
				g.logger.Debug("ball hit successfully", "newScore", g.score, "ballVelocity", ball.velocity)
			}
			continue
		}

		// Check collision with stumps
		if g.stumps.CheckCollision(ball, nil) {
			g.logger.Debug("ball collided with stumps", "ballPosition", ball.position, "score", g.score)
			g.stumps.Fall()
			g.endGame("BOWLED OUT!")
			break
		}
	}

	for _, ball := range ballsToDeactivate {
		delete(g.balls, ball)
	}
}

func (g *Game) updateGameReset() error {
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.Reset()
	}
	return nil
}

func (g *Game) updateNameInput() error {
	nameInput := string(ebiten.AppendInputChars(nil))

	// Handle backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(nameInput) > 0 {
		nameInput = nameInput[:len(nameInput)-1]
	}

	// Handle enter to submit name
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if nameInput == "" {
			nameInput = "Anonymous"
		}
		// Clean the name (remove non-printable characters)
		cleanName := strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, nameInput)

		g.highScoreManager.SetHighScore(g.score, cleanName)
		g.state = GameStateGameOver
	}

	return nil
}

func (g *Game) endGame(message string) {
	g.logger.Debug("game ended", "message", message, "finalScore", g.score)
	g.userMessage = message
	if g.highScoreManager.IsNewHighScore(g.score) {
		g.logger.Debug("new high score achieved", "score", g.score)
		g.state = GameStateNameInput
	} else {
		g.logger.Debug("game over, no new high score", "score", g.score, "currentHighScore", g.highScoreManager.GetHighScore().Score)
		g.state = GameStateGameOver
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen with black background (terminal-like)
	screen.Fill(color.RGBA{0, 0, 0, 255})

	switch g.state {
	case GameStatePlaying:
		g.drawPlaying(screen)
	case GameStateGameOver:
		g.drawGameOver(screen)
	case GameStateNameInput:
		g.drawNameInput(screen)
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Draw stumps
	g.stumps.Draw(screen)

	// Draw Bat
	g.bat.Draw(screen)

	// Draw balls
	for ball := range g.balls {
		ball.Draw(screen)
	}

	// Draw collision rectangles for debugging
	g.drawCollisionRectangles(screen)

	// Draw current score
	scoreText := fmt.Sprintf("Score: %d", g.score)
	op := &text.DrawOptions{}
	op.GeoM.Translate(20, 30)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, scoreText, assets.ScoreFont, op)

	// Draw high score
	highScoreText := g.highScoreManager.GetHighScoreText()
	op2 := &text.DrawOptions{}
	op2.GeoM.Translate(20, 60)
	op2.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, highScoreText, assets.ScoreFont, op2)

	// Draw instructions
	instructionText := "Move mouse to swing bat"
	op3 := &text.DrawOptions{}
	op3.GeoM.Translate(20, screenHeight-30)
	op3.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, instructionText, assets.ScoreFont, op3)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Draw stumps (will show out sprite if fallen)
	g.stumps.Draw(screen)

	// Draw bat (keep it visible)
	g.bat.Draw(screen)

	// Draw OUT text in big letters towards center-right
	outOp := &text.DrawOptions{}
	outOp.GeoM.Scale(2.0, 2.0) // Make text bigger
	outOp.GeoM.Translate(screenWidth/2+50, screenHeight/2-100)
	outOp.ColorScale.ScaleWithColor(color.RGBA{255, 50, 50, 255}) // Red color
	text.Draw(screen, g.userMessage, assets.ScoreFont, outOp)

	// Draw final score
	scoreText := fmt.Sprintf("Final Score: %d", g.score)
	op2 := &text.DrawOptions{}
	op2.GeoM.Translate(screenWidth/2+50, screenHeight/2-40)
	op2.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, scoreText, assets.ScoreFont, op2)

	// Draw high score
	highScoreText := g.highScoreManager.GetHighScoreText()
	op4 := &text.DrawOptions{}
	op4.GeoM.Translate(screenWidth/2+50, screenHeight/2-10)
	op4.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, highScoreText, assets.ScoreFont, op4)

	restartText := "Press R to restart"
	op3 := &text.DrawOptions{}
	op3.GeoM.Translate(screenWidth/2+50, screenHeight/2+30)
	op3.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, restartText, assets.ScoreFont, op3)
}

func (g *Game) drawNameInput(screen *ebiten.Image) {
	// Draw congratulations message
	congratsText := "NEW HIGH SCORE!"
	op := &text.DrawOptions{}
	op.GeoM.Translate(screenWidth/2-120, screenHeight/2-80)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, congratsText, assets.ScoreFont, op)

	scoreText := fmt.Sprintf("Score: %d", g.score)
	op2 := &text.DrawOptions{}
	op2.GeoM.Translate(screenWidth/2-60, screenHeight/2-40)
	op2.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, scoreText, assets.ScoreFont, op2)

	// Draw name input prompt
	promptText := "Enter your name:"
	op3 := &text.DrawOptions{}
	op3.GeoM.Translate(screenWidth/2-100, screenHeight/2)
	op3.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, promptText, assets.ScoreFont, op3)

	// Draw current name input with cursor
	nameText := "_"
	op4 := &text.DrawOptions{}
	op4.GeoM.Translate(screenWidth/2-100, screenHeight/2+30)
	op4.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, nameText, assets.ScoreFont, op4)

	// Draw instruction
	instructionText := "Press Enter to confirm"
	op5 := &text.DrawOptions{}
	op5.GeoM.Translate(screenWidth/2-120, screenHeight/2+70)
	op5.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, instructionText, assets.ScoreFont, op5)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1200, 800
}

func (g *Game) Reset() {
	g.logger.Debug("resetting game")
	g.bat = NewBat()
	g.balls = make(map[*Ball]struct{})
	g.stumps.Reset()
	g.ballSpawnTimer.Reset(ballSpawnTime)
	g.score = 0
	g.state = GameStatePlaying
	g.logger.Debug("game reset complete", "state", g.state)
}

func (g *Game) drawCollisionRectangles(screen *ebiten.Image) {
	// Draw bat collision rectangle in red
	batRect := g.bat.Collider()
	g.drawRectangleOutline(screen, batRect, color.RGBA{255, 0, 0, 255}) // Red

	// Draw ball collision rectangles in green
	for ball := range g.balls {
		if ball.IsActive() {
			ballRect := ball.Collider()
			g.drawRectangleOutline(screen, ballRect, color.RGBA{0, 255, 0, 255}) // Green
		}
	}

	// Draw stumps collision rectangle in blue
	stumpsRect := g.stumps.Collider()
	g.drawRectangleOutline(screen, stumpsRect, color.RGBA{0, 0, 255, 255}) // Blue
}

func (g *Game) drawRectangleOutline(screen *ebiten.Image, rect Rect, col color.Color) {
	// Create a 1-pixel white image to draw lines with
	lineImg := ebiten.NewImage(1, 1)
	lineImg.Fill(col)

	// Draw top line
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(rect.Width, 1)
	op.GeoM.Translate(rect.X, rect.Y)
	screen.DrawImage(lineImg, op)

	// Draw bottom line
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(rect.Width, 1)
	op.GeoM.Translate(rect.X, rect.Y+rect.Height-1)
	screen.DrawImage(lineImg, op)

	// Draw left line
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, rect.Height)
	op.GeoM.Translate(rect.X, rect.Y)
	screen.DrawImage(lineImg, op)

	// Draw right line
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, rect.Height)
	op.GeoM.Translate(rect.X+rect.Width-1, rect.Y)
	screen.DrawImage(lineImg, op)
}
