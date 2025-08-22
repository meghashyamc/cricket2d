package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/game"
)

func main() {
	ebiten.SetWindowSize(1200, 800)
	ebiten.SetWindowTitle("Cricket 2D")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	g := game.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		g.Logger.Error("error running game", "error", err)
	}
}
