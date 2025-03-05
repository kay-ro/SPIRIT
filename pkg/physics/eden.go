package physics

import (
	"fmt"
	"math"
	"physicsGUI/pkg/function"
)

//You need some different edensity calculation?
// => copy this function until you have the number of variants you need
// => give each function a unique name (GetEdensities1/GetXYEdensities)
// => insert the kind of parameters the calculation needs to the first bracket GetEdensities(param_1 type_1, ..., param_n type_n)
// => insert your calculation
// => continue by adapting RecalculateData and penaltyFunction in PortGUIPhysics\pkg\gui\main.go

// GetEdensities returns DataPoints based on the old implementation of the old getEden function
// - eden is an array with all the eden values {eden_a,eden_1,eden_2,...,eden_n,eden_b} (edensity)
// - d array with the d values {d_1,d_2,...,d_n} (Thickness)
// - sigma array with sigma values {sigma_a1,sigma_12,sigma_23,...,sigma_(n-1)(n),sigma_nb} (Roughness)
func GetEdensities(eden []float64, d []float64, sigma []float64) (function.Points, error) {
	step_n := len(d) + 1

	//If you use arrays/slices it helps to check if they have the correct length. Not necessary.
	//throw error if the param number does not match the scheme
	if len(eden) != step_n+1 {
		return nil, fmt.Errorf("Missmatch in parameter dimensionality edensities %d/thickness %d", len(eden), len(d))
	}
	if len(sigma) != step_n {
		return nil, fmt.Errorf("Missmatch in parameter dimensionality roughness %d/thickness %d", len(sigma), len(d))
	}

	//calculate distances
	var z = make([]float64, step_n)
	z[0] = 0.0
	for i := 1; i < step_n; i++ {
		z[i] = z[i-1] + d[i-1]
	}

	edensities := make(function.Points, ZNUMBER)
	zaxis := GetZAxis(d, ZNUMBER)

	for i := 0; i < ZNUMBER; i++ {
		z_i := zaxis[i]

		//calculate cumulative edensity at a specific z_i
		y := 0.0
		for step := 0; step < step_n; step++ {
			y += (eden[step+1] - eden[step]) * 0.5 * (1.0 + math.Erf((z_i-z[step])/(math.Sqrt2*math.Abs(sigma[step]))))
		}

		//create points for drawing
		edensities[i] = &function.Point{
			X:     z_i,
			Y:     y,
			Error: 0,
		}
	}

	return edensities, nil
}

func GetZAxis(d []float64, zNumber int) []float64 {
	z0 := -20.0
	var z1 = 30.0
	if len(d) > 3 {
		z1 += d[1]
	}
	for _, f := range d {
		z1 += f
	}

	zStep := (z1 - z0) / float64(zNumber)
	zAxis := make([]float64, zNumber)
	for i := 0; i < zNumber; i++ {
		zAxis[i] = z0 + float64(i)*zStep
	}

	return zAxis
}
