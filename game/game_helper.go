package game

import (
	"cmp"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meghashyamc/cricket2d/geometry"
)

func getCurrentMousePosition() *geometry.Vector {
	mouseX, mouseY := ebiten.CursorPosition()
	currentMousePos := &geometry.Vector{X: float64(mouseX), Y: float64(mouseY)}
	return currentMousePos
}

func clampValue[T cmp.Ordered](value T, min T, max T) T {
	if value > max {
		value = max
		return value
	}

	if value < min {
		value = min
	}

	return value
}
