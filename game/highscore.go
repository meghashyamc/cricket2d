package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type HighScore struct {
	Score int    `json:"score"`
	Name  string `json:"name"`
}

type HighScoreManager struct {
	filePath  string
	highScore HighScore
}

func NewHighScoreManager() *HighScoreManager {
	// Create scores directory in user's home directory or current directory
	homeDir, err := os.UserHomeDir()
	var scorePath string
	if err != nil {
		// Fallback to current directory if can't get home dir
		scorePath = "cricket2d_highscore.json"
	} else {
		scorePath = filepath.Join(homeDir, ".cricket2d_highscore.json")
	}

	hsm := &HighScoreManager{
		filePath: scorePath,
		highScore: HighScore{
			Score: 0,
			Name:  "",
		},
	}

	hsm.Load()
	return hsm
}

func (hsm *HighScoreManager) Load() {
	data, err := os.ReadFile(hsm.filePath)
	if err != nil {
		// File doesn't exist or can't be read, use default values
		return
	}

	var loadedScore HighScore
	if err := json.Unmarshal(data, &loadedScore); err != nil {
		// Invalid JSON, use default values
		return
	}

	hsm.highScore = loadedScore
}

func (hsm *HighScoreManager) Save() error {
	data, err := json.Marshal(hsm.highScore)
	if err != nil {
		return err
	}

	return os.WriteFile(hsm.filePath, data, 0644)
}

func (hsm *HighScoreManager) IsNewHighScore(score int) bool {
	return score > hsm.highScore.Score
}

func (hsm *HighScoreManager) SetHighScore(score int, name string) error {
	hsm.highScore.Score = score
	hsm.highScore.Name = name
	return hsm.Save()
}

func (hsm *HighScoreManager) GetHighScore() HighScore {
	return hsm.highScore
}

func (hsm *HighScoreManager) GetHighScoreText() string {
	if hsm.highScore.Score == 0 {
		return "High Score: 0"
	}
	if hsm.highScore.Name == "" {
		return fmt.Sprintf("High Score: %d", hsm.highScore.Score)
	}
	return fmt.Sprintf("High Score: %d (%s)", hsm.highScore.Score, hsm.highScore.Name)
}