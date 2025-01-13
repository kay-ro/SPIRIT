package helper

import (
	"image/color"

	"fyne.io/fyne/v2/canvas"
)

// separator
func CreateSeparator() *canvas.Line {
	line := canvas.NewLine(color.Gray{Y: 100})
	line.StrokeWidth = 1
	return line
}
