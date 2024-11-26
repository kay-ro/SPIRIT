package data

import "math"

func NewOldSLDFunction(eden []float64, d []float64, sigma []float64, zNumber int) *DataFunction {
	return NewDataFunction(getEden(eden, d, sigma, zNumber), INTERPOLATION_NONE)
}

func getEden(eden []float64, d []float64, sigma []float64, zNumber int) []Point { // TODO understand this mess?
	zAxis := getZAxis(d, zNumber)
	var z_a = make([]float64, len(d)+2)
	z_a[0] = 0.0
	for i := 1; i < len(d)+2; i++ {
		if i < 4 {
			z_a[i] = 0.0
		} else {
			z_a[i] = d[1]
		}
		for j := 0; j < i; j++ {
			z_a[i] += d[j]
		}
	}
	edenA := 0.333
	eden = append([]float64{edenA}, eden...)
	resEden := make([]Point, zNumber)
	for i := 0; i < zNumber; i++ {
		z := zAxis[i]
		var step float64 = 0
		for j := 0; j < len(z_a); j++ {
			step += (eden[j+1] - eden[j]) * 0.5 * (math.Erf((z - z_a[j]) / (math.Sqrt2 * sigma[j])))
		}
		resEden[i] = Point{
			X: z,
			Y: step,
		}
	}
	return resEden
}
func getZAxis(d []float64, zNumber int) []float64 {
	z0 := -20.0
	var z1 = d[1]
	for _, f := range d {
		z1 += f
	}
	z1 -= z0
	zStep := (z1 - z0) / float64(zNumber)
	zAxis := make([]float64, zNumber)
	for i := 0; i < zNumber; i++ {
		zAxis[i] = z0 + float64(i)*zStep
	}
	return zAxis
}
