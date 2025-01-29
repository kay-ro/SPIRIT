package function

type InterpolationMode int

const (
	INTERPOLATION_NONE   InterpolationMode = 0
	INTERPOLATION_LINEAR InterpolationMode = 1
	INTERPOLATION_SPLINE InterpolationMode = 2
	INTERPOLATION_PCHIP  InterpolationMode = 3
)

type FunctionInterface interface {
	Scope() (Coordinate, Coordinate)
	Model(resolution int) ([]Coordinate, []Coordinate)
	Eval(x float64) (float64, error)
}
