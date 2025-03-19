package physics

import "math"

// getConvolFunc calculates the weight for convolution
func getConvolFunc(z, z0, roughness float64) float64 {
	return math.Exp(-math.Pow(z-z0, 2) / (2.0 * math.Pow(roughness, 2)))
}

// convolute performs the convolution
func convolute(znumber int, zaxis, edens []float64, roughness float64) []float64 {
	edenConv := make([]float64, znumber)

	for i := 0; i < znumber; i++ {
		thisZ := zaxis[i]

		// Find indices of relevant values
		var indices []int
		for j, z := range zaxis {
			if math.Abs(z-thisZ) <= 2.0*roughness {
				indices = append(indices, j)
			}
		}

		// Extract local axis and values
		var loczaxis, locedens []float64
		for _, idx := range indices {
			loczaxis = append(loczaxis, zaxis[idx])
			locedens = append(locedens, edens[idx])
		}

		// Calculate weights
		var weights []float64
		var weightSum float64
		for _, z := range loczaxis {
			w := getConvolFunc(z, thisZ, roughness)
			weights = append(weights, w)
			weightSum += w
		}

		// Normalize weights
		for j := range weights {
			weights[j] /= weightSum
		}

		// Compute convolved value
		var sum float64
		for j := range weights {
			sum += weights[j] * locedens[j]
		}
		edenConv[i] = sum
	}

	return edenConv
}
