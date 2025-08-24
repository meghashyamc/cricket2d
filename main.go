package main

import (
	"os"

	"github.com/meghashyamc/cricket2d/game"
)

func main() {

	g := game.NewGame()

	if err := g.Run(); err != nil {
		g.Logger.Error("error running game", "error", err)
		os.Exit(1)
	}
}
