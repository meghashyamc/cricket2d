package game

import (
	"fmt"
	"image/color"
	"strings"
	"time"
	"unicode"

	"github.com/meghashyamc/cricket2d/logger"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/meghashyamc/cricket2d/assets"
)

const (
	screenWidth  = 1200
	screenHeight = 800

	ballSpawnTime = 2 * time.Second
)

type GameState int

const (
	GameStatePlaying GameState = iota
	GameStateGameOver
	GameStateNameInput
)

type Game struct {
	Bat              *Bat
	balls            map[*Ball]struct{}
	stumps           *Stumps
	ballSpawnTimer   *time.Ticker
	score            int
	state            GameState
	highScoreManager *HighScoreManager
	Logger           logger.Logger
}

func NewGame() *Game {
	g := &Game{
		Bat:              NewBat(),
		balls:            make(map[*Ball]struct{}),
		stumps:           NewStumps(),
		ballSpawnTimer:   time.NewTicker(ballSpawnTime),
		score:            0,
		state:            GameStatePlaying,
		highScoreManager: NewHighScoreManager(),
		Logger:           logger.New(),
	}

	g.Logger.Debug("game initialized", "screenWidth", screenWidth, "screenHeight", screenHeight, "ballSpawnTime", ballSpawnTime)
	return g
}

func (g *Game) Run() error {
	g.Logger.Debug("starting game")
	g.setupWindow()
	// ebiten.SetTPS(1)
	return ebiten.RunGame(g)
}

func (g *Game) setupWindow() {
	ebiten.SetWindowSize(1200, 800)
	ebiten.SetWindowTitle("Cricket 2D")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
}

func (g *Game) Update() error {
	switch g.state {
	case GameStatePlaying:
		return g.updatePlaying()
	case GameStateGameOver:
		return g.updateGameOver()
	case GameStateNameInput:
		return g.updateNameInput()
	}
	return nil
}

func (g *Game) updatePlaying() error {
	g.Bat.Update()

	select {
	case <-g.ballSpawnTimer.C:
		newBall := NewBall()
		g.balls[newBall] = struct{}{}
		g.Logger.Debug("new ball spawned", "ballCount", len(g.balls), "ballPosition", newBall.position)
	default:
	}
	ballsToDeactivate := make([]*Ball, 0)
	// Update balls
	for ball := range g.balls {
		ball.Update()

		if !ball.IsActive() {
			// Remove inactive balls
			ballsToDeactivate = append(ballsToDeactivate, ball)
			continue
		}

		// Check collision with Bat using precise collision detection
		if g.Bat.CheckBallCollision(ball) {
			if ball.Hit(g.Bat.GetBatAngle(), g.Bat.GetSwingVelocity()) {
				g.score++
				g.Logger.Debug("ball hit successfully", "newScore", g.score, "ballVelocity", ball.velocity)
			}
			continue
		}

		// Check collision with stumps
		if g.stumps.CheckCollision(ball) {
			g.Logger.Debug("stumps collision detected", "ballPosition", ball.position, "score", g.score)
			g.stumps.Fall()
			g.endGame("BOWLED OUT!")
			break
		}
	}

	for _, ball := range ballsToDeactivate {
		delete(g.balls, ball)
	}

	return nil
}

func (g *Game) updateGameOver() error {
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
	g.Logger.Debug("game ended", "message", message, "finalScore", g.score)
	if g.highScoreManager.IsNewHighScore(g.score) {
		g.Logger.Debug("new high score achieved", "score", g.score)
		g.state = GameStateNameInput
	} else {
		g.Logger.Debug("game over, no new high score", "score", g.score, "currentHighScore", g.highScoreManager.GetHighScore().Score)
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
		g.drawGameOver(screen, "OUT!")
	case GameStateNameInput:
		g.drawNameInput(screen)
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Draw stumps
	g.stumps.Draw(screen)

	// Draw Bat
	g.Bat.Draw(screen)

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

func (g *Game) drawGameOver(screen *ebiten.Image, gameOverText string) {
	// Draw final score and game over message
	op := &text.DrawOptions{}
	op.GeoM.Translate(screenWidth/2-100, screenHeight/2-60)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, gameOverText, assets.ScoreFont, op)

	scoreText := fmt.Sprintf("Final Score: %d", g.score)
	op2 := &text.DrawOptions{}
	op2.GeoM.Translate(screenWidth/2-100, screenHeight/2-20)
	op2.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, scoreText, assets.ScoreFont, op2)

	// Draw high score
	highScoreText := g.highScoreManager.GetHighScoreText()
	op4 := &text.DrawOptions{}
	op4.GeoM.Translate(screenWidth/2-100, screenHeight/2+10)
	op4.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, highScoreText, assets.ScoreFont, op4)

	restartText := "Press R to restart"
	op3 := &text.DrawOptions{}
	op3.GeoM.Translate(screenWidth/2-80, screenHeight/2+50)
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
	g.Logger.Debug("resetting game")
	g.Bat = NewBat()
	g.balls = make(map[*Ball]struct{})
	g.stumps.Reset()
	g.ballSpawnTimer.Reset(ballSpawnTime)
	g.score = 0
	g.state = GameStatePlaying
	g.Logger.Debug("game reset complete", "state", g.state)
}

func (g *Game) drawCollisionRectangles(screen *ebiten.Image) {
	// Draw bat collision rectangle in red
	batRect := g.Bat.Collider()
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
