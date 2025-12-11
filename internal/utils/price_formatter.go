package utils

import "math"

// RoundPrice rounds price to 2 decimal places using standard rounding (NOT truncation)
// Example: 3.14159 -> 3.14, 3.145 -> 3.15
func RoundPrice(price float64) float64 {
	return math.Round(price*100) / 100
}
