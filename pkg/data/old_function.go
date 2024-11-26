package data

import "math"

// NewOldSLDFunction returns a DataPoint function Based on the old implementation of the getEden function
//
// - eden is an array with all the eden values {eden_a,eden_1,eden_2,...,eden_n,eden_b}
//
// - d array with the d values {d_1,d_2,...,d_n}
//
// - sigma array with sigma values {sigma_a1,sigma_12,sigma_23,...,sigma_(n-1)(n),sigma_nb}
func NewOldSLDFunction(eden []float64, d []float64, sigma []float64, zNumber int) *DataFunction {
	return NewDataFunction(getEden(eden, d, sigma, zNumber), INTERPOLATION_NONE)
}

// Eden_a varies so eden[0] is now eden_a
func getEden(eden []float64, d []float64, sigma []float64, zNumber int) []Point { // TODO understand this mess?
	zAxis := getZAxis(d, zNumber)
	var z_a = make([]float64, len(d)+2)
	z_a[0] = 0.0
	for i := 1; i < len(d)+2; i++ {
		if i == 4 {
			z_a[i] = z_a[i-1] + 2*d[i-1]
		} else {
			z_a[i] = z_a[i-1] + d[i-1]
		}
	}
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
