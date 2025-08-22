package game

import "math"

type Vector struct {
	X float64
	Y float64
}

func (v Vector) Normalize() Vector {
	magnitude := math.Sqrt(v.X*v.X + v.Y*v.Y)
	if magnitude == 0 {
		return Vector{0, 0}
	}
	return Vector{v.X / magnitude, v.Y / magnitude}
}

func (v Vector) Add(other Vector) Vector {
	return Vector{v.X + other.X, v.Y + other.Y}
}

func (v Vector) Scale(factor float64) Vector {
	return Vector{v.X * factor, v.Y * factor}
}