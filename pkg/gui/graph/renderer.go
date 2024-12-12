package graph

import (
	"fmt"
	"math"
	"physicsGUI/pkg/function"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type GraphRenderer struct {
	graph   *GraphCanvas
	objects []fyne.CanvasObject

	// size of canvas
	size *fyne.Size

	// margin for labels etc.
	margin float32
}

type GraphScope struct {
	Min function.Coordinate
	Max function.Coordinate
}

// returns the minimum size needed for the graph
func (r *GraphRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 200)
}

// initializes the base strcuture for every graph
func (r *GraphRenderer) base() {
	// init drawing objects
	r.AddObject(r.graph.background)

	// axes
	r.graph.axes[0] = canvas.NewLine(axesColor) // x-axis
	r.graph.axes[0].StrokeWidth = 1
	r.graph.axes[1] = canvas.NewLine(axesColor) // y-acis
	r.graph.axes[1].StrokeWidth = 1

	// title
	r.graph.title = canvas.NewText(r.graph.config.Title, titleColor)
	r.graph.title.TextSize = 16
	r.graph.title.TextStyle.Bold = true
	r.graph.title.Move(fyne.NewPos(r.size.Width/2-float32(len(r.graph.title.Text)*4), 0))
	r.AddObject(r.graph.title)

	// background
	r.graph.background.Resize(*r.size)
	r.graph.background.Move(fyne.NewPos(0, 0))

	// x-axis
	r.graph.axes[0].Position1 = fyne.NewPos(r.margin-2, r.size.Height-r.margin+2)
	r.graph.axes[0].Position2 = fyne.NewPos(r.size.Width-r.margin/2, r.size.Height-r.margin+2)
	r.AddObject(r.graph.axes[0])

	// y-axis
	r.graph.axes[1].Position1 = fyne.NewPos(r.margin-2, r.size.Height-r.margin+2)
	r.graph.axes[1].Position2 = fyne.NewPos(r.margin-2, 0.5*r.margin)
	r.AddObject(r.graph.axes[1])
}

// draws the whole graph
func (r *GraphRenderer) Layout(size fyne.Size) {
	// clear objects
	r.objects = make([]fyne.CanvasObject, 0)

	// size of the graph
	r.size = &size

	// set the base for the canvas
	r.base()

	// get min/max
	scope := r.graph.function.Scope
	maxData := scope.MaxY
	minData := scope.MinY

	if r.graph.config.IsLog {
		if minData < scope.MinX {
			minData = scope.MinX
		}

		// Transformiere die Werte in Log-Skala
		maxData = math.Log10(maxData)
		minData = math.Log10(minData)
	}

	// vertikale grid-lines + x-labels
	numXLines := 10
	for i := 0; i <= numXLines; i++ {
		xPos := r.margin + float32(i)*float32(r.size.Width-1.5*r.margin)/float32(numXLines)

		// grid-lines
		if i > 0 && i < numXLines { // no grid line at the edge
			gridLine := createGridLine(
				fyne.NewPos(xPos, r.margin/2),
				true,
				r.size.Height-1.5*r.margin,
			)
			r.graph.gridLines = append(r.graph.gridLines, gridLine)
		}

		// label
		if i%2 == 0 {
			value := float64(i) * scope.MaxX / float64(numXLines)
			label := canvas.NewText(fmt.Sprintf("%.1f", value), legendColor)
			label.TextSize = 12
			label.Move(fyne.NewPos(xPos-15, r.size.Height-r.margin+10))
			r.AddObject(label)
		}
	}

	// horizonzal grid-lines + y-labels
	numYLines := 10
	if r.graph.config.IsLog {
		// calc start/end values for log scale
		startExp := math.Floor(math.Log10(scope.MinX))
		endExp := math.Ceil(math.Log10(math.Pow(10, maxData)))

		// calculate intermediate steps for each scale
		for exp := startExp; exp <= endExp; exp++ {
			base := math.Pow(10, exp)

			// Füge mehr Zwischenschritte innerhalb jeder Größenordnung hinzu
			for i := 1; i < 10; i++ {
				value := base * float64(i)
				if value >= scope.MinX && value <= math.Pow(10, maxData) {
					logValue := math.Log10(value)
					yPos := r.size.Height - r.margin - float32(logValue-minData)*float32(r.size.Height-1.5*r.margin)/float32(maxData-minData)

					if yPos > r.margin/2 && yPos < r.size.Height-r.margin {
						// grid-line
						gridLine := createGridLine(
							fyne.NewPos(r.margin, yPos),
							false,
							r.size.Width-1.5*r.margin,
						)
						r.graph.gridLines = append(r.graph.gridLines, gridLine)

						// label (only for "label line")
						if i == 1 {
							label := canvas.NewText(fmt.Sprintf("%.1e", value), legendColor)
							label.TextSize = 12
							label.Move(fyne.NewPos(r.margin-45, yPos-10))
							r.graph.yLabels = append(r.graph.yLabels, label)
						}
					}
				}
			}
		}
	} else {
		for i := 0; i <= numYLines; i++ {
			value := minData + (maxData-minData)*float64(i)/float64(numYLines)
			yPos := r.size.Height - r.margin - float32(i)*float32(r.size.Height-1.5*r.margin)/float32(numYLines)

			// grid-lines
			if i > 0 && i < numYLines { // no grid line at the edge
				gridLine := createGridLine(
					fyne.NewPos(r.margin, yPos),
					false,
					r.size.Width-1.5*r.margin,
				)
				r.AddObject(gridLine)
			}

			// label
			label := canvas.NewText(fmt.Sprintf("%.2f", value), legendColor)
			label.TextSize = 12
			label.Move(fyne.NewPos(r.margin-45, yPos-10))
			r.AddObject(label)
		}
	}

	// draw model lines
	//r.DrawGraphLines(maxData, minData, modelPoints)
	if r.graph.config.IsLog {
		r.DrawGraphLog(maxData, minData)
	} else {
		r.DrawGraphLinear()
	}
	// * debug
	//log.Printf("Render %d objects \n", len(r.objects))
}

func (r *GraphRenderer) DrawGraphLinear() {
	points, iPoints := r.graph.function.Model(r.graph.config.Resolution)

	// calc available space
	availableWidth := r.size.Width - (1.5 * r.margin)
	availableHeight := r.size.Height - (1.5 * r.margin)

	// complete range
	xRange := math.Abs(r.graph.function.Scope.MaxX - r.graph.function.Scope.MinX)
	yRange := math.Abs(r.graph.function.Scope.MaxY - r.graph.function.Scope.MinY)

	oX, oY := float32(0), float32(0)

	// draw line based on interpolated (resolution) points
	for i, point := range iPoints {
		// scale x value to available width
		x := float32((point.X-r.graph.function.Scope.MinX)/xRange) * availableWidth
		y := float32((point.Y-r.graph.function.Scope.MinY)/yRange) * availableHeight

		// first point is the origin
		if i == 0 {
			oX, oY = r.normalize(x, y)
			continue
		}

		xt, yt := r.normalize(x, y)

		// draw line
		r.AddObject(&canvas.Line{
			StrokeColor: lineColor,
			StrokeWidth: 1,
			Position1:   fyne.NewPos(oX, oY),
			Position2:   fyne.NewPos(xt, yt),
		})

		oX, oY = xt, yt
	}

	// draw data points
	for _, point := range points {
		// scale x value to available width
		x := float32((point.X-r.graph.function.Scope.MinX)/xRange) * availableWidth
		y := float32((point.Y-r.graph.function.Scope.MinY)/yRange) * availableHeight

		xt, yt := r.normalize(x, y)
		// error correction
		yE1 := float32((point.Y+point.Error-r.graph.function.Scope.MinY)/yRange) * availableHeight
		yE2 := float32((point.Y-point.Error-r.graph.function.Scope.MinY)/yRange) * availableHeight

		_, e1 := r.normalize(x, yE1)
		_, e2 := r.normalize(x, yE2)

		r.DrawError(xt, e1, e2)
		r.DrawPoint(xt, yt)
	}
}

func (r *GraphRenderer) DrawGraphLog(maxData, minData float64) {
	points, iPoints := r.graph.function.Model(r.graph.config.Resolution)

	// calc available space
	availableWidth := r.size.Width - (1.5 * r.margin)
	availableHeight := r.size.Height - (1.5 * r.margin)

	// Calculate shifts if needed for negative values
	xShift := 0.0
	if r.graph.function.Scope.MinX <= 0 {
		xShift = math.Abs(r.graph.function.Scope.MinX) + 1
	}
	yShift := 0.0
	if r.graph.function.Scope.MinY <= 0 {
		yShift = math.Abs(r.graph.function.Scope.MinY) + 1
	}

	// Calculate log ranges
	logMinX := math.Log10(r.graph.function.Scope.MinX + xShift)
	logMaxX := math.Log10(r.graph.function.Scope.MaxX + xShift)
	logMinY := math.Log10(r.graph.function.Scope.MinY + yShift)
	logMaxY := math.Log10(r.graph.function.Scope.MaxY + yShift)
	xRange := math.Abs(logMaxX - logMinX)
	yRange := math.Abs(logMaxY - logMinY)

	oX, oY := float32(0), float32(0)

	// draw line based on interpolated (resolution) points
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
			StrokeColor: lineColor,
			StrokeWidth: 1,
			Position1:   fyne.NewPos(oX, oY),
			Position2:   fyne.NewPos(xt, yt),
		})
		oX, oY = xt, yt
	}

	// draw data points
	for _, point := range points {
		// scale x and y values logarithmically
		logX := math.Log10(point.X + xShift)
		logY := math.Log10(point.Y + yShift)

		x := float32((logX-logMinX)/xRange) * availableWidth
		y := float32((logY-logMinY)/yRange) * availableHeight

		xt, yt := r.normalize(x, y)

		// error correction (also logarithmic)
		yE1 := float32((math.Log10(point.Y+point.Error+yShift)-logMinY)/yRange) * availableHeight
		yE2 := float32((math.Log10(point.Y-point.Error+yShift)-logMinY)/yRange) * availableHeight
		_, e1 := r.normalize(x, yE1)
		_, e2 := r.normalize(x, yE2)

		r.DrawError(xt, e1, e2)
		r.DrawPoint(xt, yt)
	}
}

// normalizes the coodinates from the bottom left of the canvas
func (r *GraphRenderer) normalize(x float32, y float32) (float32, float32) {
	return x + r.margin, r.size.Height - r.margin - y
}

// TODO: needs clean
func (r *GraphRenderer) DrawGraphLine(maxData, minData float64, points function.Points) {
	scope := r.graph.function.Scope

	// calculate scales
	xScale := (r.size.Width - 1.5*r.margin) / float32(scope.MaxX-scope.MinX)
	yScale := (r.size.Height - 1.5*r.margin) / float32(maxData)

	for i := 0; i < len(points)-1; i++ {
		line := canvas.NewLine(lineColor)
		line.StrokeWidth = 1

		y1 := r.graph.transformValue(minData, points[i].Y)
		y2 := r.graph.transformValue(minData, points[i+1].Y)

		x1 := r.margin + float32(points[i].X)*xScale
		yPos1 := r.size.Height - r.margin - float32(y1-minData)*yScale
		x2 := r.margin + float32(points[i+1].X)*xScale
		yPos2 := r.size.Height - r.margin - float32(y2-minData)*yScale

		line.Position1 = fyne.NewPos(x1, yPos1)
		line.Position2 = fyne.NewPos(x2, yPos2)

		r.AddObject(line)
	}
}

// TODO: fix the small gap between points and lines
// TODO: if points are the same size as lines
// TODO: -> points are on the bottom of the lines

// draw a grid point
func (r *GraphRenderer) DrawPoint(x float32, y float32) {
	r.AddObject(&canvas.Circle{
		FillColor: pointColor,
		Position1: fyne.NewPos(x-pointRadius, y-pointRadius),
		Position2: fyne.NewPos(x+pointRadius, y+pointRadius),
	})
}

// draw error correction lines
func (r *GraphRenderer) DrawError(x, y1, y2 float32) {
	r.AddObject(&canvas.Line{
		StrokeColor: errorColor,
		StrokeWidth: 1,
		Position1:   fyne.NewPos(x, y1),
		Position2:   fyne.NewPos(x, y2),
	})
}

// returns the objects of the graph
func (r *GraphRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// ? destroy object (is performance impacted if func is empty?)
func (r *GraphRenderer) Destroy() {}

// refresh the graph by recalculating the layout and refreshing the canvas
func (r *GraphRenderer) Refresh() {
	r.Layout(r.graph.Size())
	canvas.Refresh(r.graph)
}

// add an object to the graph renderer
func (r *GraphRenderer) AddObject(object fyne.CanvasObject) {
	r.objects = append(r.objects, object)
}
