package geometry

import (
	"math"
)

type Vector struct {
	X float64
	Y float64
}

// DotProduct calculates the dot product of two vectors
func (v Vector) DotProduct(other Vector) float64 {
	return v.X*other.X + v.Y*other.Y
}

// reflected = incident - 2*(incident·normal)*normal
func (v Vector) Reflect(normal Vector) Vector {
	dotProduct := v.DotProduct(normal)

	reflectedX := v.X - 2*dotProduct*normal.X
	reflectedY := v.Y - 2*dotProduct*normal.Y

	return Vector{
		X: reflectedX,
		Y: reflectedY,
	}
}

// Magnitude calculates the magnitude (length) of a vector
func (v Vector) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// AngleTo calculates the angle between this vector and another vector in radians
func (v Vector) AngleTo(other Vector) float64 {
	dot := v.DotProduct(other)
	magV := v.Magnitude()
	magOther := other.Magnitude()

	// Handle zero-length vectors
	if magV == 0 || magOther == 0 {
		return 0
	}

	// cos(θ) = (A · B) / (|A| * |B|)
	cosTheta := dot / (magV * magOther)

	// Clamp to [-1, 1] to handle floating point precision issues
	if cosTheta > 1 {
		cosTheta = 1
	} else if cosTheta < -1 {
		cosTheta = -1
	}

	return math.Acos(cosTheta)
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
