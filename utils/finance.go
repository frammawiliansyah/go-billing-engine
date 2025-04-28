package utils

import "math"

func PMT(rate float64, nper int, pv float64) float64 {
	if rate == 0 {
		return pv / float64(nper)
	}
	return (pv * rate) / (1 - math.Pow(1+rate, float64(-nper)))
}

func RoundFloat(x float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(x*pow) / pow
}

func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
