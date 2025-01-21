package function

import (
	"cmp"
	"math"
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
		if a == nil || b == nil {
			panic("points can't be nil")
		}

		return cmp.Compare(a.X, b.X)
	})
}

// returns min and max X or Y value of all points
func (p Points) MinMaxXY() (minX float64, maxX float64, minY float64, maxY float64) {
	if len(p) == 0 {
		return 0, 0, 0, 0
	}

	minY = p[0].Y
	maxY = p[0].Y

	minX = p[0].X
	maxX = p[0].X

	for _, point := range p {
		if maxX < point.X {
			maxX = point.X
		}
		if minX > point.X {
			minX = point.X
		}
		if maxY < point.Y {
			maxY = point.Y
		}
		if minY > point.Y {
			minY = point.Y
		}
	}
	return
}

// ? may be useless, but may simplify some stuff so we will keep it for now
func (p Points) Log() Points {
	np := make(Points, len(p))
	for i, point := range p {
		np[i] = &Point{
			X:     point.X,
			Y:     math.Log10(point.Y),
			Error: math.Log10(point.Error),
		}
	}

	return p
}

// applies magic to a point
func (p *Point) Magie() {
	p.Y = math.Pow(p.X, 4) * p.Y
	p.Error = math.Pow(p.X, 4) * p.Error
}

// applies magic to all points
func (p Points) Magie() {
	for _, point := range p {
		point.Magie()
	}
}

// copies the points
func (p Points) Copy() Points {
	np := make(Points, len(p))
	for i, point := range p {
		np[i] = &Point{
			X:     point.X,
			Y:     point.Y,
			Error: point.Error,
		}
	}

	return np
}

// filters the points by min and max x value
func (p Points) Filter(min, max float64) Points {
	np := make(Points, 0)
	for _, point := range p {
		if point.X >= min && point.X <= max {
			np = append(np, &Point{
				X:     point.X,
				Y:     point.Y,
				Error: point.Error,
			})
		}
	}

	return np
}

// represents a coordinate with x and y value
type Coordinate struct {
	X float64
	Y float64
}
