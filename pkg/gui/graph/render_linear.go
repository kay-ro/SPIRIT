package graph

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"image/color"
	"math"
	"physicsGUI/pkg/function"
)

// draw a linear graph
func (r *GraphRenderer) DrawGraphLinear(scope *function.Scope, points, iPoints function.Points, pointColor color.Color, isDataSet bool) {
	// calc available space
	availableWidth := r.size.Width - (1.5 * r.margin)
	availableHeight := r.size.Height - (1.5 * r.margin)

	//
	order := int(math.Floor(math.Log10(math.Abs(scope.MaxY - scope.MinY))))
	minX := floorInOrder(scope.MinX, order)
	minY := floorInOrder(scope.MinY, order)
	maxX := ceilInOrder(scope.MaxX, order)
	maxY := ceilInOrder(scope.MaxY, order)

	// complete range
	xRange := math.Abs(maxX - minX)
	yRange := math.Abs(maxY - minY)

	oX, oY := float32(0), float32(0)

	// draw line based on interpolated (resolution) points
	if !isDataSet {
		for i, point := range iPoints {
			// scale x value to available width
			x := float32((point.X-minX)/xRange) * availableWidth
			y := float32((point.Y-minY)/yRange) * availableHeight

			// first point is the origin
			if i == 0 {
				oX, oY = r.normalize(x, y)
				continue
			}

			xt, yt := r.normalize(x, y)

			// draw line
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
		// scale x value to available width
		x := float32((point.X-minX)/xRange) * availableWidth
		y := float32((point.Y-minY)/yRange) * availableHeight

		xt, yt := r.normalize(x, y)

		// error correction
		yE1 := float32((point.Y+point.Error-minY)/yRange) * availableHeight
		yE2 := float32((point.Y-point.Error-minY)/yRange) * availableHeight

		_, e1 := r.normalize(x, yE1)
		_, e2 := r.normalize(x, yE2)

		if isDataSet {
			r.DrawError(xt, e1, e2, errorColor)
		}
		r.DrawPoint(xt, yt, pointColor)
	}
}
func floorInOrder(num float64, order int) float64 {
	return math.Floor(num*math.Pow10(-order)) * math.Pow10(order)
}
func ceilInOrder(num float64, order int) float64 {
	return math.Ceil(num*math.Pow10(-order)) * math.Pow10(order)
}

// draw grid lines and labels for linear scale
func (r *GraphRenderer) DrawGridLinear(scope *function.Scope) {

	//
	order := int(math.Floor(math.Log10(math.Abs(scope.MaxY - scope.MinY))))
	minX := floorInOrder(scope.MinX, order)
	minY := floorInOrder(scope.MinY, order)
	maxX := ceilInOrder(scope.MaxX, order)
	maxY := ceilInOrder(scope.MaxY, order)

	// horizontal grid-lines + y-labels
	yGridCount := int(math.Ceil((maxY - minY) * math.Pow10(-order))) //TODO make -
	yStep := (maxY - minY) / float64(yGridCount)

	for i := 0; i <= yGridCount; i++ {
		value := minY + yStep*float64(i)
		yPos := r.size.Height - r.margin - float32(i)*float32(r.size.Height-1.5*r.margin)/float32(yGridCount)

		if i > 0 {
			r.DrawGridLine(fyne.NewPos(r.margin, yPos), false, false)
		}

		text := fmt.Sprintf("%.3f", value)
		if value < 0.01 {
			text = fmt.Sprintf("%.1e", value)
		}

		// label
		label := &canvas.Text{
			Text:     text,
			Color:    legendColor,
			TextSize: 12,
		}
		label.Move(fyne.NewPos(r.margin-45, yPos-10))
		r.AddObject(label)
	}

	// vertical grid-lines + x-labels
	xGridCount := int(r.size.Width / 25)
	xStep := math.Abs(maxX-minX) / float64(xGridCount)

	for i := 0; i <= xGridCount; i++ {
		xPos := r.margin + float32(i)*float32(r.size.Width-1.5*r.margin)/float32(xGridCount)

		if i > 0 {
			r.DrawGridLine(fyne.NewPos(xPos, r.margin/2), true, false)
		}

		// only draw every second label to prevent overlapping
		if i%4 == 0 {
			v := minX + xStep*float64(i)
			text := fmt.Sprintf("%.3f", v)
			if v < 0.01 {
				text = fmt.Sprintf("%.1e", v)
			}

			// label
			label := &canvas.Text{
				Text:     text,
				Color:    legendColor,
				TextSize: 12,
			}

			label.Move(fyne.NewPos(xPos-15, r.size.Height-r.margin+10))
			r.AddObject(label)
		}
	}
}
