package graph

import (
	"fmt"
	"log"
	"math"
	"physicsGUI/pkg/function"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type GraphRenderer struct {
	graph   *GraphCanvas
	objects []fyne.CanvasObject
}

// returns the minimum size needed for the graph
func (r *GraphRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 200)
}

// initializes the base strcuture for every graph
func (r *GraphRenderer) base(margin float32, size fyne.Size) {
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
	r.graph.title.Move(fyne.NewPos(size.Width/2-float32(len(r.graph.title.Text)*4), 0))
	r.AddObject(r.graph.title)

	// background
	r.graph.background.Resize(size)
	r.graph.background.Move(fyne.NewPos(0, 0))

	// x-axis
	r.graph.axes[0].Position1 = fyne.NewPos(margin, size.Height-margin)
	r.graph.axes[0].Position2 = fyne.NewPos(size.Width-margin/2, size.Height-margin)
	r.AddObject(r.graph.axes[0])

	// y-axis
	r.graph.axes[1].Position1 = fyne.NewPos(margin, size.Height-margin)
	r.graph.axes[1].Position2 = fyne.NewPos(margin, margin)
	r.AddObject(r.graph.axes[1])
}

// draws the whole graph
func (r *GraphRenderer) Layout(size fyne.Size) {
	r.objects = make([]fyne.CanvasObject, 0)

	// margin for labels etc.
	margin := float32(50)

	// set the base for the canvas
	r.base(margin, size)

	// data
	dataPoints, modelPoints := r.graph.function.Model(r.graph.config.Resolution)

	// get min/max
	minDataP, maxDataP := r.graph.function.Scope()
	maxData := maxDataP.Y
	minData := minDataP.Y

	if r.graph.config.IsLog {
		if minData < minDataP.X {
			minData = minDataP.X
		}

		// Transformiere die Werte in Log-Skala
		maxData = math.Log10(maxData)
		minData = math.Log10(minData)
	}

	// vertikale grid-lines + x-labels
	numXLines := 10
	for i := 0; i <= numXLines; i++ {
		xPos := margin + float32(i)*float32(size.Width-1.5*margin)/float32(numXLines)

		// grid-lines
		if i > 0 && i < numXLines { // no grid line at the edge
			gridLine := createGridLine(
				fyne.NewPos(xPos, margin/2),
				true,
				size.Height-1.5*margin,
			)
			r.graph.gridLines = append(r.graph.gridLines, gridLine)
		}

		// label
		if i%2 == 0 {
			value := float64(i) * maxDataP.X / float64(numXLines)
			label := canvas.NewText(fmt.Sprintf("%.1f", value), legendColor)
			label.TextSize = 12
			label.Move(fyne.NewPos(xPos-15, size.Height-margin+10))
			r.AddObject(label)
		}
	}

	// horizonzal grid-lines + y-labels
	numYLines := 10
	if r.graph.config.IsLog {
		// calc start/end values for log scale
		startExp := math.Floor(math.Log10(minDataP.X))
		endExp := math.Ceil(math.Log10(math.Pow(10, maxData)))

		// calculate intermediate steps for each scale
		for exp := startExp; exp <= endExp; exp++ {
			base := math.Pow(10, exp)

			// Füge mehr Zwischenschritte innerhalb jeder Größenordnung hinzu
			for i := 1; i < 10; i++ {
				value := base * float64(i)
				if value >= minDataP.X && value <= math.Pow(10, maxData) {
					logValue := math.Log10(value)
					yPos := size.Height - margin - float32(logValue-minData)*float32(size.Height-1.5*margin)/float32(maxData-minData)

					if yPos > margin/2 && yPos < size.Height-margin {
						// grid-line
						gridLine := createGridLine(
							fyne.NewPos(margin, yPos),
							false,
							size.Width-1.5*margin,
						)
						r.graph.gridLines = append(r.graph.gridLines, gridLine)

						// label (only for "label line")
						if i == 1 {
							label := canvas.NewText(fmt.Sprintf("%.1e", value), legendColor)
							label.TextSize = 12
							label.Move(fyne.NewPos(margin-45, yPos-10))
							r.graph.yLabels = append(r.graph.yLabels, label)
						}
					}
				}
			}
		}
	} else {
		for i := 0; i <= numYLines; i++ {
			value := minData + (maxData-minData)*float64(i)/float64(numYLines)
			yPos := size.Height - margin - float32(i)*float32(size.Height-1.5*margin)/float32(numYLines)

			// grid-lines
			if i > 0 && i < numYLines { // no grid line at the edge
				gridLine := createGridLine(
					fyne.NewPos(margin, yPos),
					false,
					size.Width-1.5*margin,
				)
				r.AddObject(gridLine)
			}

			// label
			label := canvas.NewText(fmt.Sprintf("%.2f", value), legendColor)
			label.TextSize = 12
			label.Move(fyne.NewPos(margin-45, yPos-10))
			r.AddObject(label)
		}
	}

	xScale := (size.Width - 1.5*margin) / float32(maxDataP.X-minDataP.X)
	yScale := (size.Height - 1.5*margin) / float32(maxData)

	// draw model lines
	r.DrawGraphLines(maxData, minData, size, margin, modelPoints)

	//draw data points
	for i := 0; i < len(dataPoints)-1; i++ {
		y := r.graph.transformValue(minDataP.X, dataPoints[i].Y)
		var y1 float64
		if r.graph.config.IsLog {
			y1 = y
		} else {
			y1 = r.graph.transformValue(minDataP.X, dataPoints[i].Y-dataPoints[i].Error)
		}
		y2 := r.graph.transformValue(minDataP.X, dataPoints[i].Y+dataPoints[i].Error)
		x := margin + float32(dataPoints[i].X)*xScale
		yPos := size.Height - margin - float32(y-minData)*yScale
		yPos1 := size.Height - margin - float32(y1-minData)*yScale
		yPos2 := size.Height - margin - float32(y2-minData)*yScale

		r.DrawError(x, yPos1, yPos2)
		r.DrawPoint(x, yPos)
	}

	log.Printf("Render %d objects \n", len(r.objects))
}

// TODO: needs clean
func (r *GraphRenderer) DrawGraphLines(maxData, minData float64, size fyne.Size, margin float32, points function.Points) {
	minDataP, maxDataP := r.graph.function.Scope()

	// calculate scales
	xScale := (size.Width - 1.5*margin) / float32(maxDataP.X-minDataP.X)
	yScale := (size.Height - 1.5*margin) / float32(maxData)

	for i := 0; i < len(points)-1; i++ {
		line := canvas.NewLine(lineColor)
		line.StrokeWidth = 1

		y1 := r.graph.transformValue(minData, points[i].Y)
		y2 := r.graph.transformValue(minData, points[i+1].Y)

		x1 := margin + float32(points[i].X)*xScale
		yPos1 := size.Height - margin - float32(y1-minData)*yScale
		x2 := margin + float32(points[i+1].X)*xScale
		yPos2 := size.Height - margin - float32(y2-minData)*yScale

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
		StrokeWidth: 2,
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
