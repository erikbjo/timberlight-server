package utils

import "math"

// RoundToNearest25Deg rounds a latitude or longitude to the nearest 0.25 degrees.
func RoundToNearest25Deg(coordinate float64) float64 {
	return (math.Round(coordinate * 4)) / 4
}

// RoundToNearest50Deg rounds a latitude or longitude to the nearest 0.50 degrees.
func RoundToNearest50Deg(coordinate float64) float64 {
	return (math.Round(coordinate * 2)) / 2
}

// RoundToNearest500 rounds a number to the nearest 500.
// This is used to cluster the features into 1000x1000 meter squares, ending in 500.
func RoundToNearest500(n int) int {
	base := (n / 1000) * 1000
	return base + 500
}
