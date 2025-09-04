package geometry

import (
	"math"
)

// distanceFromPointToLine calculates the shortest distance from a point to a line segment
func DistanceFromPointToLine(point, lineStart, lineEnd Vector) float64 {
	// Vector from line start to end
	lineVec := Vector{X: lineEnd.X - lineStart.X, Y: lineEnd.Y - lineStart.Y}
	// Vector from line start to point
	pointVec := Vector{X: point.X - lineStart.X, Y: point.Y - lineStart.Y}
	return pointVec.Magnitude() * math.Sin(pointVec.AngleTo(lineVec))
}
