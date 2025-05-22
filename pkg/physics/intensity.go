package physics

import (
	"fmt"
	"math"
	"math/cmplx"
	"physicsGUI/pkg/function"
	"physicsGUI/pkg/gui/helper"
	"slices"
	"sort"
)

// need more than one? => Copy and rename it. Be careful when to use which axis.
var qzAxis = GetDefaultQZAxis(qzNumber)
var qzNumber = 500

type IntensityOptions struct {
	Background float64
	Scaling    float64
	Resolution float64
}

func CalculateIntensityPoints(edenPoints function.Points, deltaq float64, opts *IntensityOptions) function.Points {
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
	modifiedQzAxis := helper.Map(qzAxis, func(xPoint float64) float64 { return xPoint + deltaq })

	intensity := CalculateIntensity(modifiedQzAxis, deltaz, sld, opts)

	// creates list with intensity points based on edenPoints x and error and calculated intensity as y
	intensityPoints := make(function.Points, qzNumber)
	for i := range intensity {
		intensityPoints[i] = &function.Point{
			X:     qzAxis[i],
			Y:     intensity[i],
			Error: 0.0,
		}
	}

	intensityPoints_convoluted := make(function.Points, qzNumber)
	intensityPoints_convoluted = convolute(qzNumber, qzAxis, intensityPoints, 1e-5*opts.Resolution)

	return intensityPoints_convoluted
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
	qzAxis_tmp := make([]float64, qzNumber)
	for i := 0; i < qzNumber; i++ {
		qzAxis_tmp[i] = -0.02 + float64(i)*0.001
	}
	return qzAxis_tmp
}

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
		qzNumber = len(qzAxis)
	}

}

// calculate a penalty between the calculated intensity and the loaded data sets
func Sim2SigRMS(dataSets []function.Points, intensity function.Points) (float64, error) {
	if len(intensity) != qzNumber {
		return math.MaxFloat64, fmt.Errorf("rms calculation: intensity slice has the wrong length: %d vs %d", len(intensity), qzNumber)
	}
	var diff float64
	for _, dataSet := range dataSets {
		for _, point := range dataSet {
			y_intensity, err := function.GetY(intensity, point.X)
			if err != nil {
				return math.MaxFloat64, fmt.Errorf("rms calculation: there is no intensity for: %f", point.X)
			}
			diff += math.Pow(point.X, 2) * math.Pow((y_intensity-point.Y)/point.Error, 2)
		}
	}
	return diff, nil
}
