package function

import (
	"fmt"
)

type InterpolationFunction func(points Points, x float64) (float64, error)

func linearInterpolation(points Points, x float64) (float64, error) {
	var lower = 0
	var upper = 1

	for i, v := range points {
		if v.X < x {
			lower = i
			upper = i + 1
		} else {
			break
		}
	}

	if upper == len(points) {
		return -1, fmt.Errorf("linearInterpolator eval error: out of bounds. %f is not in the scope (%f, %f)", x, points[0], points[lower])
	}

	lp := points[lower]
	up := points[upper]
	m := (up.Y - lp.Y) / (up.X - lp.X)
	b := lp.Y - m*lp.X

	return m*x + b, nil
}
