package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/meghashyamc/cricket2d/config"
	"github.com/meghashyamc/cricket2d/logger"
)

type HighScore struct {
	Score int    `json:"score"`
	Name  string `json:"name"`
}

type HighScoreManager struct {
	filePath  string
	highScore HighScore
	logger    logger.Logger
}

func NewHighScoreManager(cfg *config.Config) (*HighScoreManager, error) {
	logger := logger.New()
	if err := os.MkdirAll(cfg.GetDataDir(), 0755); err != nil {
		logger.Error("could not create data directory", "error", err)
		return nil, err
	}

	scoreFilePath := filepath.Join(cfg.GetDataDir(), cfg.GetScoreFilename())

	hsm := &HighScoreManager{
		filePath: scoreFilePath,
		highScore: HighScore{
			Score: 0,
			Name:  "",
		},
		logger: logger,
	}

	hsm.logger.Debug("high score manager created", "score_path", scoreFilePath)
	hsm.Load()
	return hsm, nil
}

func (hsm *HighScoreManager) Load() {
	hsm.logger.Debug("attempting to load high score", "file_path", hsm.filePath)
	data, err := os.ReadFile(hsm.filePath)
	if err != nil {
		// File doesn't exist or can't be read, use default values
		hsm.logger.Debug("high score file not found or unreadable, using defaults", "error", err)
		return
	}

	var loadedScore HighScore
	if err := json.Unmarshal(data, &loadedScore); err != nil {
		hsm.logger.Debug("invalid JSON in high score file, using defaults", "error", err)
		return
	}

	hsm.highScore = loadedScore
	hsm.logger.Debug("high score loaded successfully", "score", loadedScore.Score, "name", loadedScore.Name)
}

func (hsm *HighScoreManager) Save() error {
	hsm.logger.Debug("attempting to save high score", "score", hsm.highScore.Score, "name", hsm.highScore.Name)
	data, err := json.Marshal(hsm.highScore)
	if err != nil {
		hsm.logger.Debug("failed to marshal high score", "error", err)
		return err
	}

	err = os.WriteFile(hsm.filePath, data, 0644)
	if err != nil {
		hsm.logger.Debug("failed to write high score file", "error", err)
		return err
	}

	hsm.logger.Debug("high score saved successfully", "filePath", hsm.filePath)

	return nil
}

func (hsm *HighScoreManager) IsNewHighScore(score int) bool {
	isNew := score > hsm.highScore.Score
	return isNew
}

func (hsm *HighScoreManager) SetHighScore(score int, name string) error {
	hsm.logger.Debug("setting new high score", "score", score, "name", name)
	hsm.highScore.Score = score
	hsm.highScore.Name = name
	return hsm.Save()
}

func (hsm *HighScoreManager) GetHighScoreText(prefixText string) string {

	if hsm.highScore.Name == "" {
		return fmt.Sprintf("%s%d", prefixText, hsm.highScore.Score)
	}
	return fmt.Sprintf("%s%d (%s)", prefixText, hsm.highScore.Score, hsm.highScore.Name)
}
