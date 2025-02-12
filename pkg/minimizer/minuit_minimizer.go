package minimizer

type DummyFCN struct {
	FCN func(parameter []float64) float64
}

func NewDummyFCN(fcn func(parameter []float64) float64) *DummyFCN {
	return &DummyFCN{FCN: fcn}
}

func (d *DummyFCN) ValueOf(par []float64) float64 {
	return d.FCN(par)
}
