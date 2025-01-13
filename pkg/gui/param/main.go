package param

import (
	"errors"
	"image/color"
)

type ParameterGroup[T any] map[string]map[string]T

var (
	// sParams is a map of string parameter groups
	// each group contains a map of parameter labels and their values
	// each group can be used for iterating through parameters of the same type
	sParams = make(ParameterGroup[*Parameter[string]])

	// fParams is a map of float parameter groups
	// each group contains a map of parameter labels and their values
	// each group can be used for iterating through parameters of the same type (edesntiy, roughness, thickness)
	fParams = make(ParameterGroup[*Parameter[float64]])

	// fParams is a map of float parameter groups
	// each group contains a map of parameter labels and their values
	// each group can be used for iterating through parameters of the same type (limits, number of slabs)
	iParams = make(ParameterGroup[*Parameter[int]])

	// label configs
	labelColor  = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	minMaxColor = color.NRGBA{R: 120, G: 120, B: 120, A: 255}

	// ErrParameterNotFound is returned when a parameter is not found
	ErrParameterNotFound = errors.New("parameter not found")
)
