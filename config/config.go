package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const keyEnv = "ENV"
const envLocal = "local"

type Config struct {
	config *viper.Viper
}

func Load(env string) (*Config, error) {

	if len(env) == 0 {
		if env = os.Getenv(keyEnv); len(env) == 0 {
			env = envLocal
		}
	}

	configPath, err := getConfigPath(env)

	viperConfig := viper.New()
	if err == nil {
		viperConfig.SetConfigFile(configPath)
		if err := viperConfig.ReadInConfig(); err != nil {
			slog.Warn(fmt.Sprintf("error reading config file, %s", err))
		}
	}
	viperConfig.AutomaticEnv()

	cfg := &Config{
		config: viperConfig,
	}

	return cfg, nil
}

func (c *Config) GetWindowWidth() int {
	windowWidth := c.config.GetInt("WINDOW_WIDTH")
	if windowWidth == 0 {
		windowWidth = c.config.GetInt("window.width")
	}

	return windowWidth
}

func (c *Config) GetWindowHeight() int {
	windowHeight := c.config.GetInt("WINDOW_HEIGHT")
	if windowHeight == 0 {
		windowHeight = c.config.GetInt("window.height")
	}

	return windowHeight
}

func (c *Config) GetWindowTitle() string {
	windowTitle := c.config.GetString("WINDOW_TITLE")
	if len(windowTitle) == 0 {
		windowTitle = c.config.GetString("window.title")
	}

	return windowTitle
}

func (c *Config) GetDataDir() string {
	dataDir := c.config.GetString("DATA_DIR")
	if len(dataDir) == 0 {
		dataDir = c.config.GetString("data.dir")
	}

	return dataDir
}

func (c *Config) GetScoreFilename() string {
	scoreFilename := c.config.GetString("SCORE_FILENAME")
	if len(scoreFilename) == 0 {
		scoreFilename = c.config.GetString("data.scorefilename")
	}

	return scoreFilename
}

func (c *Config) GetBallSpawnTime() int {
	ballSpawnTimeSeconds := c.config.GetInt("BALL_SPAWN_TIME_SECONDS")
	if ballSpawnTimeSeconds == 0 {
		ballSpawnTimeSeconds = c.config.GetInt("game.ballspawntime_seconds")
	}

	return ballSpawnTimeSeconds
}

func getProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	for {
		configDir := filepath.Join(currentDir, "config")
		if info, err := os.Stat(configDir); err == nil && info.IsDir() {
			return currentDir, nil
		}

		parent := filepath.Dir(currentDir)

		if parent == currentDir {
			break
		}

		currentDir = parent
	}

	return "", fmt.Errorf("could not find project root (directory containing 'config' folder)")
}

func getConfigPath(env string) (string, error) {
	configFile := fmt.Sprintf("config.%s.yaml", env)

	projectRoot, err := getProjectRoot()
	if err != nil {
		slog.Warn("failed to find project root with config directory, will use environment variables instead", "err", err.Error())
		return "", fmt.Errorf("failed to find project root: %w", err)
	}
	configPath := filepath.Join(projectRoot, "config", configFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		slog.Warn("failed to find config file within config directory, will use environment variables instead", "err", err.Error())
		return "", fmt.Errorf("config file does not exist: %s", configPath)
	}

	return configPath, nil
}
