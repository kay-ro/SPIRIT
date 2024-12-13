package parameter

import (
	"fyne.io/fyne/v2/data/binding"
)

var uut *Parameter
var nameValue binding.String
var minValue binding.Float
var valValue binding.Float
var maxValue binding.Float
var defaultValue binding.Float
var checkValue binding.Bool

func init() {
	nameValue = binding.NewString()
	minValue = binding.NewFloat()
	maxValue = binding.NewFloat()
	valValue = binding.NewFloat()
	defaultValue = binding.NewFloat()
	checkValue = binding.NewBool()

	uut = NewParameter(nameValue, defaultValue, valValue, minValue, maxValue, checkValue)
}
