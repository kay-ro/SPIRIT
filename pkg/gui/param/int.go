package param

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2/widget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// standard int formater for int to string conversion
func StdIntFormater(f int) string {
	return fmt.Sprintf("%d", f)
}

// create a new int parameter
func IntParameter(defaultValue int) *Parameter[int] {
	return New(&Config[int]{
		InitialValue: defaultValue,
		Validator: func(s string) error {
			return nil
		},
		Format: StdIntFormater,
		Parser: strconv.Atoi,
	})
}

// create a new int input canvas object with a label
// returns the canvas object and the parameter
func Int(group, label string, defaultValue int) (fyne.CanvasObject, *Parameter[int]) {
	if iParams[group] == nil {
		iParams[group] = NewGroupElements[int]()
	}

	if !iParams[group].Check(label) {
		log.Fatal(errors.New("parameter key '" + label + "' already exists in group '" + group + "'"))
	}

	intParameter := IntParameter(defaultValue)
	intParameter.enableFit = widget.NewCheck("", func(b bool) {})

	// add parameter to group
	iParams[group].Add(label, intParameter)

	lbl := &canvas.Text{Text: label, Color: labelColor, TextSize: 14}

	return container.NewVBox(
		container.NewBorder(nil, nil, lbl, intParameter.enableFit),
		intParameter.Widget(),
	), intParameter
}
