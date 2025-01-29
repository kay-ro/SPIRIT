package helper

import (
	"fyne.io/fyne/v2/canvas"
	"image/color"
)

// separator
func CreateSeparator() *canvas.Line {
	line := canvas.NewLine(color.Gray{Y: 100})
	line.StrokeWidth = 1
	return line
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}
