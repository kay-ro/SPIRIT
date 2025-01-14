package param

import (
	"errors"
	"image/color"
)

type ParameterGroup[T any] map[string]map[string]T

var (
	// sNextFreeID an ID for all String parameters
	// will be given to a parameter when registered, to enforce the correct sequence when reading group
	sNextFreeID = make(map[string]int)
	sParamsID   = make(map[string]map[string]int)

	// sParams is a map of string parameter groups
	// each group contains a map of parameter labels and their values
	// each group can be used for iterating through parameters of the same type
	sParams = make(ParameterGroup[*Parameter[string]])

	// fNextFreeID an ID for all Float parameters
	// will be given to a parameter when registered, to enforce the correct sequence when reading group
	fNextFreeID = make(map[string]int)
	fParamsID   = make(map[string]map[string]int)

	// fParams is a map of float parameter groups
	// each group contains a map of parameter labels and their values
	// each group can be used for iterating through parameters of the same type (edesntiy, roughness, thickness)
	fParams = make(ParameterGroup[*Parameter[float64]])

	// iNextFreeID an ID for all Int parameters
	// will be given to a parameter when registered, to enforce the correct sequence when reading group
	iNextFreeID = make(map[string]int)
	iParamsID   = make(map[string]map[string]int)

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
