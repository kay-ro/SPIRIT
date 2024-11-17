package data

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
)

const (
	INTERPOLATION_NONE   = iota
	INTERPOLATION_LINEAR = iota
	INTERPOLATION_SPLINE = iota
	INTERPOLATION_PCHIP  = iota
)

type Point struct {
	X   float64
	Y   float64
	ERR float64
}

type Function struct {
	data  []Point
	minX  float64
	maxX  float64
	minY  float64
	maxY  float64
	inter interpolationProvider
}

func NewFunction(data []Point, interpolationMode int) *Function {
	var maxY float64 = data[0].Y
	var minY float64 = data[0].Y
	for _, point := range data {
		if maxY < point.Y {
			maxY = point.Y
		}
		if minY > point.Y {
			minY = point.Y
		}
	}

	slices.SortFunc(data, func(a, b Point) int {
		return cmp.Compare(a.X, b.X)
	})
	var minX float64 = data[0].X
	var maxX float64 = data[len(data)-1].X

	f := Function{
		data:  data,
		minX:  minX,
		maxX:  maxX,
		minY:  minY,
		maxY:  maxY,
		inter: nil,
	}
	var interpolation interpolationProvider
	switch interpolationMode {
	case INTERPOLATION_NONE:
		break
	case INTERPOLATION_LINEAR:
		interpolation = NewLinearInterpolator(&f.data)
		break
	case INTERPOLATION_SPLINE:
		panic("Function_Error: Interpolation Not Yet implemented")
		break
	case INTERPOLATION_PCHIP:
		panic("Function_Error: Interpolation Not Yet implemented")
		break
	default:
		panic("Function_Error: Unknown interpolationMode. Please use only values provided by related const's in data package")
	}
	f.inter = interpolation
	return &f
}

func (this *Function) Model(resolution int) ([]Point, []Point) {
	if this.inter == nil {
		return this.data, this.data
	}
	var interpolatedModel = make([]Point, resolution)
	deltaX := (this.maxX - this.minX) / float64(resolution)
	for i := 0; i < resolution; i++ {
		x := this.minX + float64(i)*deltaX
		y, _ := this.inter.eval(x)

		interpolatedModel[i] = Point{
			X:   x,
			Y:   y,
			ERR: 0,
		}
	}

	return this.data, interpolatedModel
}

type interpolationProvider interface {
	eval(x float64) (float64, error)
}

type LinearInterpolator struct {
	data *[]Point
}

func NewLinearInterpolator(data *[]Point) *LinearInterpolator {
	return &LinearInterpolator{data}
}
func (this *LinearInterpolator) eval(x float64) (float64, error) {
	var lower = 0
	var upper = 1

	for i, v := range *this.data {
		if v.X < x {
			lower = i
			upper = i + 1
		} else {
			break
		}
	}
	if upper == len(*this.data) {
		return -1, errors.New(fmt.Sprintf("Interpolator_Error: Out of bounds %f is not in the scoupe between %f and %f", x, (*this.data)[0], (*this.data)[lower]))
	}

	lp := (*this.data)[lower]
	up := (*this.data)[upper]
	m := (up.Y - lp.Y) / (up.X - lp.X)
	b := lp.Y - m*lp.X
	return m*x + b, nil
}
