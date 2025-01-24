package graph

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"image/color"
	"math"
	"physicsGUI/pkg/function"
)

// draw the graph in logarithmic scale
func (r *GraphRenderer) DrawGraphLog(scope *function.Scope, points, iPoints function.Points, pointColor color.Color, isDataSet bool) {
	// calc available space
	availableWidth := r.size.Width - (1.5 * r.margin)
	availableHeight := r.size.Height - (1.5 * r.margin)

	// Calculate shifts if needed for negative values
	xShift := 0.0
	if scope.MinX <= 0 {
		xShift = math.Abs(scope.MinX) + 1
	}
	yShift := 0.0
	if scope.MinY <= 0 {
		yShift = math.Abs(scope.MinY) + 2
	}

	// Calculate log ranges
	logMinX := math.Floor(math.Log10(scope.MinX + xShift))
	logMaxX := math.Ceil(math.Log10(scope.MaxX + xShift))
	logMinY := math.Floor(math.Log10(scope.MinY + yShift))
	logMaxY := math.Ceil(math.Log10(scope.MaxY + yShift))
	xRange := math.Abs(logMaxX - logMinX)
	yRange := math.Abs(logMaxY - logMinY)

	oX, oY := float32(0), float32(0)

	// draw line based on interpolated (resolution) points
	if !isDataSet {
		for i, point := range iPoints {
			// scale x and y values logarithmically
			logX := math.Log10(point.X + xShift)
			logY := math.Log10(point.Y + yShift)

			x := float32((logX-logMinX)/xRange) * availableWidth
			y := float32((logY-logMinY)/yRange) * availableHeight

			if i == 0 {
				oX, oY = r.normalize(x, y)
				continue
			}

			xt, yt := r.normalize(x, y)
			r.AddObject(&canvas.Line{
				StrokeColor: pointColor,
				StrokeWidth: 1,
				Position1:   fyne.NewPos(oX, oY),
				Position2:   fyne.NewPos(xt, yt),
			})
			oX, oY = xt, yt
		}
	}

	// draw data points
	for _, point := range points {
		// scale x and y values logarithmically
		logX := math.Log10(point.X + xShift)
		logY := math.Log10(point.Y + yShift)

		x := float32((logX-logMinX)/xRange) * availableWidth
		y := float32((logY-logMinY)/yRange) * availableHeight

		xt, yt := r.normalize(x, y)
		if isDataSet {
			// error correction (also logarithmic)
			yE1 := float32((math.Log10(point.Y+point.Error+yShift)-logMinY)/yRange) * availableHeight
			yE2 := float32((math.Log10(point.Y-point.Error+yShift)-logMinY)/yRange) * availableHeight
			_, e1 := r.normalize(x, yE1)
			_, e2 := r.normalize(x, yE2)

			r.DrawError(xt, e1, e2, errorColor)
		}
		r.DrawPoint(xt, yt, pointColor)
	}
}

func (r *GraphRenderer) DrawGridLog(scope *function.Scope) {
	// Horizontal grid-lines + y-labels (logarithmic)
	minLogY := math.Floor(math.Log10(math.Max(scope.MinY, 1e-10)))
	maxLogY := math.Ceil(math.Log10(scope.MaxY))
	yGridCount := int(maxLogY - minLogY)

	for i := 0; i <= yGridCount; i++ {
		// Calculate logarithmic value
		value := math.Pow(10, minLogY+float64(i))

		// Convert log space to screen space
		logVal := math.Log10(value)
		yPos := r.size.Height - r.margin -
			float32((logVal-minLogY)/(maxLogY-minLogY))*
				float32(r.size.Height-1.5*r.margin)

		if i > 0 {
			r.DrawGridLine(fyne.NewPos(r.margin, yPos), false, false)

			// Add minor grid lines between major decades
			for j := 2; j < 10; j++ {
				minorValue := value * float64(j)
				if minorValue < math.Pow(10, maxLogY) {
					minorLogVal := math.Log10(minorValue)
					minorYPos := r.size.Height - r.margin -
						float32((minorLogVal-minLogY)/(maxLogY-minLogY))*
							float32(r.size.Height-1.5*r.margin)

					r.DrawGridLine(fyne.NewPos(r.margin, minorYPos), false, true)
				}
			}
		}

		text := fmt.Sprintf("%.3f", value)
		if value < 0.01 {
			text = fmt.Sprintf("%.0e", value)
		}

		// Label for major grid lines
		label := &canvas.Text{
			Text:     text,
			Color:    legendColor,
			TextSize: 12,
		}
		label.Move(fyne.NewPos(r.margin-45, yPos-10))
		r.AddObject(label)
	}

	// Vertical grid-lines + x-labels (logarithmic)
	minLogX := math.Floor(math.Log10(math.Max(scope.MinX, 1e-10)))
	maxLogX := math.Ceil(math.Log10(scope.MaxX))
	xGridCount := int(maxLogX - minLogX)

	for i := 0; i <= xGridCount; i++ {
		// Calculate logarithmic value
		value := math.Pow(10, minLogX+float64(i))

		// Convert log space to screen space
		logVal := math.Log10(value)
		xPos := r.margin +
			float32((logVal-minLogX)/(maxLogX-minLogX))*
				float32(r.size.Width-1.5*r.margin)

		if i > 0 {
			r.DrawGridLine(fyne.NewPos(xPos, r.margin/2), true, false)

			// Add minor grid lines between major decades
			for j := 2; j < 10; j++ {
				minorValue := value * float64(j)
				if minorValue < math.Pow(10, maxLogX) {
					minorLogVal := math.Log10(minorValue)
					minorXPos := r.margin +
						float32((minorLogVal-minLogX)/(maxLogX-minLogX))*
							float32(r.size.Width-1.5*r.margin)
					r.DrawGridLine(fyne.NewPos(minorXPos, r.margin/2), true, true)
				}
			}
		}

		// Label for major grid lines
		text := fmt.Sprintf("%.3f", value)
		if value < 0.01 {
			text = fmt.Sprintf("%.0e", value)
		}
		label := &canvas.Text{
			Text:     text,
			Color:    legendColor,
			TextSize: 12,
		}
		label.Move(fyne.NewPos(xPos-25, r.size.Height-r.margin+10))
		r.AddObject(label)
	}
}
