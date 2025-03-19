package function

type InterpolationMode int

const (
	INTERPOLATION_NONE   InterpolationMode = 0
	INTERPOLATION_LINEAR InterpolationMode = 1
)

type FunctionInterface interface {
	Scope() (Coordinate, Coordinate)
	Model(resolution int) ([]Coordinate, []Coordinate)
	Eval(x float64) (float64, error)
}
