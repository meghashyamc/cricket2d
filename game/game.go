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

const (
	sleepTimeBeforeShowingHighScore = 1 * time.Second
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
		return g.updateGameOver()
	case GameStateNameInput:
		return g.updateNameInput()
	}
	return nil
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

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return int(g.cfg.GetWindowWidth()), int(g.cfg.GetWindowHeight())
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

func (g *Game) updateGameOver() error {

	// Reset
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.reset()
	}

	// Allow user to enter high score
	if g.highScoreManager.IsNewHighScore(g.score) {
		time.Sleep(sleepTimeBeforeShowingHighScore)
		g.logger.Info("new high score achieved", "score", g.score)
		g.state = GameStateNameInput
		return nil
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

	g.logger.Info("game over, no new high score", "score", g.score, "current_high_score", g.highScoreManager.highScore)
	g.state = GameStateGameOver

}

func (g *Game) drawPlaying(screen *ebiten.Image) {

	// Draw stumps, bat and ball
	g.stumps.draw(screen)
	g.bat.draw(screen)

	for ball := range g.balls {
		ball.draw(screen)
	}

	// Draw other text that shows up in the game
	const (
		scoreX float64 = 20
		scoreY float64 = 30
	)

	const (
		highScoreX float64 = 20
		highScoreY float64 = 60
	)

	var (
		instructionX float64 = 20
		instructionY float64 = g.cfg.GetWindowHeight() - 30
	)

	g.drawText(screen, fmt.Sprintf("%s%d", "Score: ", g.score), scoreX, scoreY, 1, 1, color.White)
	g.drawText(screen, g.highScoreManager.GetHighScoreText("High Score: "), highScoreX, highScoreY, 1, 1, color.White)
	g.drawText(screen, "Move mouse to swing bat", instructionX, instructionY, 1, 1, color.White)

}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	g.stumps.draw(screen)
	g.bat.draw(screen)

	// Draw OUT, final score, high score and restart text
	var (
		outX float64 = g.cfg.GetWindowWidth()/2 + 50
		outY float64 = g.cfg.GetWindowHeight()/2 - 100
	)
	var (
		finalScoreX float64 = g.cfg.GetWindowWidth()/2 + 50
		finalScoreY float64 = g.cfg.GetWindowHeight()/2 - 40
	)

	var (
		highScoreX float64 = g.cfg.GetWindowWidth()/2 + 50
		highScoreY float64 = g.cfg.GetWindowHeight()/2 - 10
	)

	var (
		restartX float64 = g.cfg.GetWindowWidth()/2 + 50
		restartY float64 = g.cfg.GetWindowHeight()/2 + 30
	)
	g.drawText(screen, g.userMessage, outX, outY, 2, 2, color.RGBA{255, 50, 50, 255})
	g.drawText(screen, fmt.Sprintf("Final Score: %d", g.score), finalScoreX, finalScoreY, 1, 1, color.White)
	g.drawText(screen, g.highScoreManager.GetHighScoreText("High Score: "), highScoreX, highScoreY, 1, 1, color.White)
	g.drawText(screen, "Press R to restart", restartX, restartY, 1, 1, color.White)

}

func (g *Game) drawNameInput(screen *ebiten.Image) {

	var (
		congratsX float64 = g.cfg.GetWindowWidth()/2 - 120
		congratsY float64 = g.cfg.GetWindowHeight()/2 - 80
	)

	var (
		scoreX float64 = g.cfg.GetWindowWidth()/2 - 60
		scoreY float64 = g.cfg.GetWindowHeight()/2 - 40
	)

	var (
		namePromptX float64 = g.cfg.GetWindowWidth()/2 - 100
		namePromptY float64 = g.cfg.GetWindowHeight() / 2
	)

	var (
		nameInputX float64 = g.cfg.GetWindowWidth()/2 - 100
		nameInputY float64 = g.cfg.GetWindowHeight()/2 + 30
	)

	var (
		confirmInstructionX float64 = g.cfg.GetWindowWidth()/2 - 120
		confirmInstructionY float64 = g.cfg.GetWindowHeight()/2 + 70
	)

	g.drawText(screen, "NEW HIGH SCORE!", congratsX, congratsY, 1, 1, color.White)
	g.drawText(screen, fmt.Sprintf("Score: %d", g.score), scoreX, scoreY, 1, 1, color.White)
	g.drawText(screen, "Enter your name:", namePromptX, namePromptY, 1, 1, color.White)
	g.drawText(screen, "_", nameInputX, nameInputY, 1, 1, color.White)
	g.drawText(screen, "Press enter to confirm", confirmInstructionX, confirmInstructionY, 1, 1, color.White)

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

func (g *Game) drawText(screen *ebiten.Image, textToDraw string, posX, posY, scaleX, scaleY float64, textColor color.Color) {
	options := &text.DrawOptions{}
	options.GeoM.Scale(scaleX, scaleY)
	options.GeoM.Translate(posX, posY)
	options.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, textToDraw, assets.ScoreFont, options)
}
