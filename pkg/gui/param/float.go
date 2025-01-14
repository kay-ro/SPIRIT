package param

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// standard float formater for float64 to string conversion
func StdFloatFormater(f float64) string {
	return fmt.Sprintf("%f", f)
}

// standard float parser for string to float64 conversion
func StdFloatParser(f string) (float64, error) {
	return strconv.ParseFloat(f, 64)
}

// create a new float parameter
func FloatParameter(defaultValue float64) *Parameter[float64] {
	return New(&Config[float64]{
		InitialValue: defaultValue,
		Validator: func(s string) error {
			if _, err := strconv.ParseFloat(s, 64); err != nil {
				return errors.New("keine gültige Zahl")
			}

			return nil
		},
		Format: StdFloatFormater,
		Parser: StdFloatParser,
	})
}

// create a new float input canvas object with a label
// returns the canvas object and the parameter
func Float(group, label string, defaultValue float64) (fyne.CanvasObject, *Parameter[float64]) {
	if fParams[group] == nil {
		fNextFreeID[group] = 0
		fParamsID[group] = make(map[string]int)
		fParams[group] = make(map[string]*Parameter[float64])
	}

	if fParams[group][label] != nil {
		log.Fatal(errors.New("parameter key '" + label + "' already exists in group '" + group + "'"))
	}

	floatParameter := FloatParameter(defaultValue)
	fParams[group][label] = floatParameter
	fParamsID[group][label] = fNextFreeID[group]
	fNextFreeID[group] += 1

	lbl := &canvas.Text{Text: label, Color: labelColor, TextSize: 14}

	return container.NewVBox(
		lbl,
		floatParameter.Widget(),
	), floatParameter
}

// create a new float input canvas object with a label and two min max input fields
// returns the canvas object and the parameter
func FloatMinMax(group, label string, defaultValue float64) (fyne.CanvasObject, *Parameter[float64]) {
	if fParams[group] == nil {
		fNextFreeID[group] = 0
		fParamsID[group] = make(map[string]int)
		fParams[group] = make(map[string]*Parameter[float64])
	}

	if fParams[group][label] != nil {
		log.Fatal(errors.New("parameter key '" + label + "' already exists in group '" + group + "'"))
	}

	min := New(&Config[float64]{
		InitialValue: 0,
		Validator: func(s string) error {
			if _, err := strconv.ParseFloat(s, 64); err != nil {
				return errors.New("keine gültige Zahl")
			}

			return nil
		},
		Format: StdFloatFormater,
		Parser: StdFloatParser,
	})

	max := New(&Config[float64]{
		InitialValue: 100,
		Validator: func(s string) error {
			if _, err := strconv.ParseFloat(s, 64); err != nil {
				return errors.New("keine gültige Zahl")
			}

			return nil
		},
		Format: StdFloatFormater,
		Parser: StdFloatParser,
	})

	// TODO: add validator update for min/max changes

	param := New(&Config[float64]{
		InitialValue: defaultValue,
		Validator: func(s string) error {
			value, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return errors.New("keine gültige Zahl")
			}

			// TODO: handle min/max not set (empty)

			gin, err := min.Get()
			if err != nil {
				return err
			}

			gax, err := max.Get()
			if err != nil {
				return err
			}

			if value < gin || value > gax {
				return fmt.Errorf("value (%s) out of range (%f - %f)", s, gin, gax)
			}

			return nil
		},
		Format: StdFloatFormater,
		Parser: StdFloatParser,
	})
	param.SetRelatives(min, max)

	fParams[group][label] = param
	fParamsID[group][label] = fNextFreeID[group]
	fNextFreeID[group] += 1

	lbl := &canvas.Text{Text: label, Color: labelColor, TextSize: 14}
	minL := &canvas.Text{Text: "Minimum", Color: minMaxColor, TextSize: 11}
	maxL := &canvas.Text{Text: "Maximum", Color: minMaxColor, TextSize: 11}

	return container.NewVBox(
		lbl,
		param.Widget(),
		container.NewGridWithColumns(2,
			minL,
			maxL,
		),
		container.NewGridWithColumns(2,
			min.Widget(),
			max.Widget(),
		),
	), param
}
