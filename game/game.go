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

const (
	gameEndMessageHitWicket = "HIT WICKET!"
	gameEndMessageBowled    = "BOWLED!"
)

type Game struct {
	cfg              *config.Config
	bat              *bat
	balls            map[*ball]struct{}
	stumps           *stumps
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
		bat:              newBat(),
		balls:            make(map[*ball]struct{}),
		stumps:           newStumps(float64(cfg.GetWindowHeight())),
		ballSpawnTimer:   time.NewTicker(time.Duration(cfg.GetballSpawnTime()) * time.Second),
		score:            0,
		state:            GameStatePlaying,
		highScoreManager: highScoreManager,
		logger:           logger.New(),
		userMessage:      "",
	}

	g.logger.Info("game initialized", "ball_spawn_time_seconds", cfg.GetballSpawnTime())
	return g, nil
}

func (g *Game) Run() error {
	g.logger.Info("starting game")
	g.setupWindow()

	// Running the game calls Update() on every 'tick'
	return ebiten.RunGame(g)
}

func (g *Game) setupWindow() {
	ebiten.SetWindowSize(int(g.cfg.GetWindowWidth()), int(g.cfg.GetWindowHeight()))
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
	g.bat.update()

	select {
	// New balls should come in at regular intervals
	case <-g.ballSpawnTimer.C:
		newball := newBall(float64(g.cfg.GetWindowWidth()), float64(g.cfg.GetWindowHeight()))
		g.balls[newball] = struct{}{}
		g.logger.Debug("new ball spawned", "ballCount", len(g.balls), "ballPosition", newball.position)

	// On every tick, check if the wicket has been hit by the bat
	default:
		if g.stumps.checkCollision(nil, g.bat) {
			g.logger.Debug("bat collided with stumps", "score", g.score)
			g.stumps.fall()
			g.endGame(gameEndMessageHitWicket)
			return nil
		}
	}

	g.updateballs()

	return nil
}

func (g *Game) updateballs() {
	ballsToDeactivate := make([]*ball, 0)

	for ball := range g.balls {
		ball.update(g.cfg.GetWindowWidth(), g.cfg.GetWindowHeight())

		if !ball.active {
			// Remove inactive balls
			ballsToDeactivate = append(ballsToDeactivate, ball)
			continue
		}

		if g.bat.checkCollision(ball) {
			if ball.hit(g.bat) {
				g.score++
				g.logger.Debug("ball hit successfully", "newScore", g.score, "ballVelocity", ball.velocity)
			}
			continue
		}

		// Check ball's collision with stumps
		if g.stumps.checkCollision(ball, nil) {
			g.logger.Debug("ball collided with stumps", "ballPosition", ball.position, "score", g.score)
			g.stumps.fall()
			g.endGame(gameEndMessageBowled)
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
	g.userMessage = message
	if g.highScoreManager.IsNewHighScore(g.score) {
		g.logger.Info("new high score achieved", "score", g.score)
		g.state = GameStateNameInput
		return
	}

	g.logger.Info("game over, no new high score", "score", g.score, "currentHighScore", g.highScoreManager.GetHighScore().Score)
	g.state = GameStateGameOver

}

func (g *Game) draw(screen *ebiten.Image) {
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
	g.stumps.draw(screen)
	g.bat.draw(screen)

	for ball := range g.balls {
		ball.draw(screen)
	}

	var (
		scoreX float64 = 20
		scoreY float64 = 30
	)

	var (
		highScoreX float64 = 20
		highScoreY float64 = 60
	)

	var (
		instructionX float64 = 20
		instructionY float64 = g.cfg.GetWindowHeight() - 30
	)

	g.drawScore(screen, scoreX, scoreY)
	g.drawHighScore(screen, highScoreX, highScoreY)
	g.drawInstruction(screen, instructionX, instructionY)

}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	g.stumps.draw(screen)

	g.bat.draw(screen)

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

func (g *Game) reset() {
	g.logger.Debug("resetting game")
	g.bat = newBat()
	g.balls = make(map[*ball]struct{})
	g.stumps.reset()
	g.ballSpawnTimer.Reset(time.Duration(g.cfg.GetballSpawnTime()) * time.Second)
	g.score = 0
	g.state = GameStatePlaying
	g.logger.Debug("game reset complete", "state", g.state)
}

func (g *Game) drawScore(screen *ebiten.Image, posX, posY float64) {
	scoreText := fmt.Sprintf("Score: %d", g.score)
	op := &text.DrawOptions{}
	op.GeoM.Translate(posX, posY)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, scoreText, assets.ScoreFont, op)

}

func (g *Game) drawHighScore(screen *ebiten.Image, posX, posY float64) {
	highScoreText := g.highScoreManager.GetHighScoreText()
	op2 := &text.DrawOptions{}
	op2.GeoM.Translate(posX, posY)
	op2.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, highScoreText, assets.ScoreFont, op2)
}

func (g *Game) drawInstruction(screen *ebiten.Image, posX, posY float64) {
	instructionText := "Move mouse to swing bat"
	op3 := &text.DrawOptions{}
	op3.GeoM.Translate(posX, posY)
	op3.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, instructionText, assets.ScoreFont, op3)
}
