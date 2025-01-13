package param

import (
	"errors"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// standard float formater for string to string conversion
func StdStringFormater(f string) string {
	return f
}

// standard string parser for string to string conversion
func StdStringParser(f string) (string, error) {
	return f, nil
}

// create a new string parameter
func StringParameter(defaultValue string) *Parameter[string] {
	return New(&Config[string]{
		InitialValue: defaultValue,
		Validator: func(s string) error {
			return nil
		},
		Format: StdStringFormater,
		Parser: StdStringParser,
	})
}

// created a new string input field with a label
func String(group, label, defaultValue string) (fyne.CanvasObject, *Parameter[string]) {
	if sParams[group][label] != nil {
		log.Fatal(errors.New("parameter key '" + label + "' already exists in group '" + group + "'"))
	}

	stringParameter := StringParameter(defaultValue)
	sParams[group][label] = stringParameter

	lbl := &canvas.Text{Text: label, Color: labelColor, TextSize: 14}

	return container.NewVBox(
		lbl,
		stringParameter.Widget(),
	), stringParameter
}
