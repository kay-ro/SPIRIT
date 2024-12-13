package parameter

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type Parameter struct {
	widget.BaseWidget
	check *widget.Check
	name  *widget.Entry
	val   *FilteredEntry
	min   *FilteredEntry
	max   *FilteredEntry
}

func (p *Parameter) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewVBox(p.name, container.NewHBox(container.NewCenter(p.check), container.NewCenter(p.val), container.NewVBox(p.max, p.min))))
}

func NewParameter(nameVal binding.String, defaultVal, value, min, max binding.Float, checkVal binding.Bool) *Parameter {

	// create name text field with linked data
	name := widget.NewEntryWithData(nameVal)
	// create filtered entry fields, which only accept runes relevant for float inputs
	val := NewFilteredEntry('0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '-', 'e', '.')
	minV := NewFilteredEntry('0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '-', 'e', '.')
	maxV := NewFilteredEntry('0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '-', 'e', '.')

	// make checkbox for locking data for minimizer
	check := widget.NewCheck("", func(b bool) {
		//TODO add icon and icon change on pressing?
	})

	// Bind gui representation to the data
	val.Bind(binding.FloatToString(value))
	minV.Bind(binding.FloatToString(min))
	maxV.Bind(binding.FloatToString(max))
	check.Bind(checkVal)

	// update placeholder text, when default Value changed
	defaultVal.AddListener(binding.NewDataListener(func() {
		// set placeholder text of val to default value, when available
		if def, err := defaultVal.Get(); err == nil {
			val.PlaceHolder = fmt.Sprint(def)
		}
	}))

	return &Parameter{
		BaseWidget: widget.BaseWidget{},
		name:       name,
		check:      check,
		val:        val,
		min:        minV,
		max:        maxV,
	}
}
