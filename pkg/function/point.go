package function

import (
	"cmp"
	"slices"
)

// represents a point with x and y value and and error value
type Point struct {
	X     float64
	Y     float64
	Error float64
}

type Points []*Point

// sorts all points by X value
func (p Points) Sort() {
	slices.SortFunc(p, func(a, b *Point) int {
		return cmp.Compare(a.X, b.X)
	})
}

// returns min and max X value of all points
func (p Points) MinMaxX() (float64, float64) {
	p.Sort()

	return p[0].X, p[len(p)-1].X
}

// returns min and max Y value of all points
func (p Points) MinMaxY() (float64, float64) {
	if len(p) == 0 {
		return 0, 0
	}

	min := p[0].Y
	max := p[0].Y

	for _, point := range p {
		if max < point.Y {
			max = point.Y
		}
		if min > point.Y {
			min = point.Y
		}
	}

	return min, max
}

// represents a coordinate with x and y value
type Coordinate struct {
	X float64
	Y float64
}
