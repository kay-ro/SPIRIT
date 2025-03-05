package physics

import (
	"math"
	"math/cmplx"
	"physicsGUI/pkg/function"
	"physicsGUI/pkg/gui/helper"
	"slices"
	"sort"
)

type IntensityOptions struct {
	Background float64
	Scaling    float64
}

func CalculateIntensityPoints(edenPoints function.Points, delta float64, opts *IntensityOptions) function.Points {
	// transform points into sld floats
	sld := make([]float64, len(edenPoints))
	for i, e := range edenPoints {
		sld[i] = e.Y * ELECTRON_RADIUS
	}

	deltaz := 0.0

	if edenPoints != nil && len(edenPoints) > 1 {
		deltaz = edenPoints[1].X - edenPoints[0].X
	}

	// calculate intensity
	modifiedQzAxis := make([]float64, len(qzAxis))
	copy(modifiedQzAxis, qzAxis)

	helper.Map(modifiedQzAxis, func(xPoint float64) float64 { return xPoint + delta })

	intensity := CalculateIntensity(qzAxis, deltaz, sld, opts)

	// creates list with intensity points based on edenPoints x and error and calculated intensity as y
	intensityPoints := make(function.Points, len(qzAxis))
	for i := range intensity {
		intensityPoints[i] = &function.Point{
			X:     qzAxis[i],
			Y:     intensity[i],
			Error: 0.0,
		}
	}

	return intensityPoints
}

// CalculateIntensity calculates intensity from the slds
func CalculateIntensity(qzaxis []float64, deltaz float64, sld []float64, opts *IntensityOptions) []float64 {
	// Get reflectivity values
	refl := CalculateReflectivity(qzaxis, deltaz, sld)

	// return reflectivity if no options are given (default: scaling=1, background=0)
	if opts == nil {
		return refl
	}

	// Calculate intensity with scaling and background
	intensity := make([]float64, len(refl))
	for i := range refl {
		intensity[i] = opts.Scaling*refl[i] + opts.Background
	}

	return intensity
}

// calculates reflectivity using the Parratt formalism
//
// qzaxis: momentum transfer values
//
// deltaz: layer thicknesses
//
// sld: scattering length densities
func CalculateReflectivity(qzaxis []float64, deltaz float64, sld []float64) []float64 {
	ci := complex(0, 1.0)
	c1 := complex(1.0, 0)
	c0 := complex(0, 0)

	qznumber := len(qzaxis)
	nmedia := len(sld)
	ninterfaces := nmedia - 1
	nslabs := nmedia - 2

	// Initialize output array
	refl := make([]float64, qznumber)

	// Calculate reflectivity for each q value
	for iq := 0; iq < qznumber; iq++ {
		k0 := complex(qzaxis[iq]/2.0, 0)

		// Calculate wave vectors
		k := make([]complex128, nmedia)
		for i := 0; i < nmedia; i++ {
			// ? real(k0) is fine? orig: k[i] = SQRT(DCOMPLEX(k0^2-4.0*pi*(rho[i]-rho[0]), 0))
			k[i] = cmplx.Sqrt(complex(math.Pow(real(k0), 2)-4.0*math.Pi*(sld[i]-sld[0]), 0))
		}

		// Calculate Fresnel coefficients
		rfres := make([]complex128, ninterfaces)
		for i := 0; i < ninterfaces; i++ {
			rfres[i] = (k[i] - k[i+1]) / (k[i] + k[i+1])
		}

		// Calculate phase factors
		hphase := make([]complex128, nslabs)
		fphase := make([]complex128, nslabs)
		for i := 0; i < nslabs; i++ {
			hphase[i] = cmplx.Exp(ci * k[i+1] * complex(deltaz, 0))
			fphase[i] = cmplx.Exp(2.0 * ci * k[i+1] * complex(deltaz, 0))
		}

		// Calculate partial reflectivity amplitudes
		rparr := make([]complex128, nmedia)
		for i := 0; i < nmedia; i++ {
			i2 := nmedia - 1 - i

			if i >= 2 {
				numerator := rfres[i2] + rparr[i2+1]*fphase[i2]
				denominator := c1 + rfres[i2]*rparr[i2+1]*fphase[i2]
				rparr[i2] = numerator / denominator
			} else if i == 1 {
				rparr[i2] = rfres[i2]
			} else if i == 0 {
				rparr[i2] = c0
			}
		}

		// Calculate final reflectivity
		refl[iq] = math.Pow(cmplx.Abs(rparr[0]), 2)
	}

	return refl
}

func GetDefaultQZAxis(qzNumber int) []float64 {
	qzAxis := make([]float64, qzNumber)
	for i := 0; i < qzNumber; i++ {
		qzAxis[i] = -0.02 + float64(i)*0.001
	}
	return qzAxis
}

// could be made more efficient
// use the combined experimental axis as current qz axis
func AlterQZAxis(dataSets function.Functions, graphID string) {
	if graphID == "intensity" {
		var qzValues []float64
		for _, dataSet := range dataSets {
			for _, point := range dataSet.GetData() {
				qzValues = append(qzValues, point.X)
			}
		}
		sort.Float64s(qzValues)
		qzValues = slices.Compact(qzValues)
		qzAxis = qzValues
	}

}
