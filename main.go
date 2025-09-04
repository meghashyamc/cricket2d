package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/meghashyamc/cricket2d/config"
	"github.com/meghashyamc/cricket2d/game"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %s\n", err)
		os.Exit(1)
	}
	g, err := game.NewGame(cfg)
	if err != nil {
		os.Exit(1)
	}
	if err := g.Run(); err != nil {
		slog.Error("error running game", "err", err)
		os.Exit(1)
	}
}
