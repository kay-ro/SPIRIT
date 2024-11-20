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
type Function interface {
	Scope() (Position, Position)
	Model(resolution int) ([]Point, []Point)
	Eval(x float64) (float64, error)
}

type DataFunction struct {
	data  []Point
	minX  float64
	maxX  float64
	minY  float64
	maxY  float64
	inter interpolationProvider
}

func NewDataFunction(data []Point, interpolationMode int) *DataFunction {
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

	f := DataFunction{
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

func (this *DataFunction) Model(resolution int) ([]Point, []Point) {
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
func (this *DataFunction) Eval(x float64) (float64, error) {
	return this.inter.eval(x)
}

type Position struct {
	X float64
	Y float64
}

func NewPos(x float64, y float64) Position {
	return Position{
		X: x,
		Y: y,
	}
}

func (this *DataFunction) Scope() (Position, Position) {
	return NewPos(this.minX, this.minY), NewPos(this.maxX, this.maxY)
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

type FunctionSegment struct {
	start float64
	end   float64
	minY  float64
	maxY  float64
	f     *func(x float64) float64
}

func NewFunctionSegment(start float64, end float64, f *func(x float64) float64) *FunctionSegment {
	//TODO Add min max calc
	return &FunctionSegment{
		start: start,
		end:   end,
		minY:  0,
		maxY:  0,
		f:     f,
	}
}

type SegmentedFunction struct {
	segments []FunctionSegment
	minY     float64
	maxY     float64
}

func NewSegmentedFunction(segments []FunctionSegment) *SegmentedFunction {
	var minY float64 = 0
	var maxY float64 = 0

	for _, segment := range segments {
		if segment.minY < minY {
			minY = segment.minY
		}
		if segment.maxY > maxY {
			maxY = segment.maxY
		}
	}

	return &SegmentedFunction{
		segments: segments,
		minY:     minY,
		maxY:     maxY,
	}
}

func (this *SegmentedFunction) Scope() (Position, Position) {
	if this.segments == nil || len(this.segments) == 0 {
		return NewPos(0, 0), NewPos(0, 0)
	}
	return NewPos(this.segments[0].start, this.minY), NewPos(this.segments[len(this.segments)-1].end, this.maxY)
}

func (this *SegmentedFunction) Model(resolution int) ([]Point, []Point) {
	if this.segments == nil || len(this.segments) == 0 {
		return nil, nil
	}
	dx := (this.segments[len(this.segments)-1].end - this.segments[0].start) / float64(resolution)
	res := make([]Point, resolution)
	var nx = this.segments[0].start
	for i := 0; i < resolution; i++ {
		val, err := this.Eval(nx)
		if err != nil {
			println(err.Error())
			res[i] = Point{
				X:   0,
				Y:   0,
				ERR: 0,
			}
		}
		res[i] = Point{
			X:   nx,
			Y:   val,
			ERR: 0,
		}
		nx += dx
	}
	return nil, res
}

func (this *SegmentedFunction) Eval(x float64) (float64, error) {
	if this.segments == nil || len(this.segments) == 0 {
		return 0, errors.New("Evaluation_Error: Unable to evaluate when no segments are defined")
	}
	if x < this.segments[0].start || x > this.segments[len(this.segments)-1].end {
		return 0, errors.New(fmt.Sprintf("Evaluation_Error: Out of bounds %f is not in the scoupe between %f and %f", x, this.segments[0].start, this.segments[len(this.segments)-1].end))
	}
	for i := 0; i < len(this.segments); i++ {
		if this.segments[i].end >= x {
			return (*this.segments[i].f)(x), nil
		}
	}
	panic("Evaluation_Error: This should not be able to happen")
}
