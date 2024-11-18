package gui

import (
	"fmt"
	"image/color"
	"math"
	"physicsGUI/pkg/data"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

var (
	legendColor = color.White
	lineColor   = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	pointColor  = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	errorColor  = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	gridColor   = color.NRGBA{R: 128, G: 128, B: 128, A: 64}
)

type GraphCanvas struct {
	widget.BaseWidget
	data       *data.Function
	lines      []*canvas.Line
	points     []fyne.CanvasObject // i%2==1 -> circle, i%2==0 -> line
	gridLines  []*canvas.Line
	axes       []*canvas.Line
	background *canvas.Rectangle
	xLabels    []*canvas.Text
	yLabels    []*canvas.Text
	title      *canvas.Text

	config *GraphConfig
}

type GraphConfig struct {
	Title      string
	IsLog      bool
	MinValue   float64
	Resolution int
	Data       *data.Function
}

func NewGraphCanvas(config *GraphConfig) *GraphCanvas {
	g := &GraphCanvas{
		data:    config.Data,
		lines:   make([]*canvas.Line, 0),
		axes:    make([]*canvas.Line, 2),
		xLabels: make([]*canvas.Text, 0),
		yLabels: make([]*canvas.Text, 0),

		config: config,

		background: canvas.NewRectangle(color.Black),
	}
	g.ExtendBaseWidget(g)

	// axes
	g.axes[0] = canvas.NewLine(color.White) // x-axis
	g.axes[0].StrokeWidth = 2
	g.axes[1] = canvas.NewLine(color.White) // y-acis
	g.axes[1].StrokeWidth = 2

	// title
	g.title = canvas.NewText(config.Title, color.White)
	g.title.TextSize = 16
	g.title.TextStyle.Bold = true

	return g
}

// helper function for the grid lines
func createGridLine(pos fyne.Position, isVertical bool, length float32) *canvas.Line {
	line := canvas.NewLine(gridColor)
	line.StrokeWidth = 1

	if isVertical {
		line.Position1 = pos
		line.Position2 = fyne.NewPos(pos.X, pos.Y+length)
	} else {
		line.Position1 = pos
		line.Position2 = fyne.NewPos(pos.X+length, pos.Y)
	}

	return line
}

// transform values to log10 if applicable
func (g *GraphCanvas) transformValue(value float64) float64 {
	if g.config.IsLog {
		if value < g.config.MinValue {
			value = g.config.MinValue
		}
		return math.Log10(value)
	}
	return value
}

func (g *GraphCanvas) CreateRenderer() fyne.WidgetRenderer {
	return &GraphRenderer{
		graph:   g,
		objects: make([]fyne.CanvasObject, 0),
	}
}

type GraphRenderer struct {
	graph   *GraphCanvas
	objects []fyne.CanvasObject
}

func (r *GraphRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 200)
}

func (r *GraphRenderer) Layout(size fyne.Size) {
	// margin for labels etc.
	margin := float32(50)

	// data
	dataPoints, modelPoints := r.graph.data.Model(r.graph.config.Resolution)

	// title position
	r.graph.title.Move(fyne.NewPos(size.Width/2-float32(len(r.graph.title.Text)*4), 0 /* margin/2 */))

	// background
	r.graph.background.Resize(size)
	r.graph.background.Move(fyne.NewPos(0, 0))

	// x-axis
	r.graph.axes[0].Position1 = fyne.NewPos(margin, size.Height-margin)
	r.graph.axes[0].Position2 = fyne.NewPos(size.Width-margin/2, size.Height-margin)

	// y-axis
	r.graph.axes[1].Position1 = fyne.NewPos(margin, size.Height-margin)
	r.graph.axes[1].Position2 = fyne.NewPos(margin, margin)

	// get min/max
	maxData := r.graph.data.MaxY
	minData := r.graph.data.MinY

	if r.graph.config.IsLog {
		if minData < r.graph.config.MinValue {
			minData = r.graph.config.MinValue
		}
		// Transformiere die Werte in Log-Skala
		maxData = math.Log10(maxData)
		minData = math.Log10(minData)
	}

	// Grid-Linien und Labels erstellen
	r.graph.gridLines = make([]*canvas.Line, 0)
	r.graph.yLabels = make([]*canvas.Text, 0)
	r.graph.xLabels = make([]*canvas.Text, 0)

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
			value := float64(i) * r.graph.data.MaxX / float64(numXLines)
			label := canvas.NewText(fmt.Sprintf("%.1f", value), legendColor)
			label.TextSize = 12
			label.Move(fyne.NewPos(xPos-15, size.Height-margin+10))
			r.graph.xLabels = append(r.graph.xLabels, label)
		}
	}

	// horizonzal grid-lines + y-labels
	numYLines := 10
	if r.graph.config.IsLog {
		// calc start/end values for log scale
		startExp := math.Floor(math.Log10(r.graph.config.MinValue))
		endExp := math.Ceil(math.Log10(math.Pow(10, maxData)))

		// calculate intermediate steps for each scale
		for exp := startExp; exp <= endExp; exp++ {
			base := math.Pow(10, exp)

			// Füge mehr Zwischenschritte innerhalb jeder Größenordnung hinzu
			for i := 1; i < 10; i++ {
				value := base * float64(i)
				if value >= r.graph.config.MinValue && value <= math.Pow(10, maxData) {
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
				r.graph.gridLines = append(r.graph.gridLines, gridLine)
			}

			// label
			label := canvas.NewText(fmt.Sprintf("%.2f", value), legendColor)
			label.TextSize = 12
			label.Move(fyne.NewPos(margin-45, yPos-10))
			r.graph.yLabels = append(r.graph.yLabels, label)
		}
	}

	// calculate scales
	xScale := (size.Width - 1.5*margin) / float32(r.graph.data.MaxX-r.graph.data.MinX)
	yScale := (size.Height - 1.5*margin) / float32(maxData)

	// draw model lines
	r.graph.lines = make([]*canvas.Line, len(modelPoints)-1)
	for i := 0; i < len(modelPoints)-1; i++ {
		line := canvas.NewLine(lineColor)
		line.StrokeWidth = 1

		y1 := r.graph.transformValue(modelPoints[i].Y)
		y2 := r.graph.transformValue(modelPoints[i+1].Y)

		x1 := margin + float32(modelPoints[i].X)*xScale
		yPos1 := size.Height - margin - float32(y1-minData)*yScale
		x2 := margin + float32(modelPoints[i+1].X)*xScale
		yPos2 := size.Height - margin - float32(y2-minData)*yScale

		line.Position1 = fyne.NewPos(x1, yPos1)
		line.Position2 = fyne.NewPos(x2, yPos2)
		r.graph.lines[i] = line
	}

	//draw data points
	r.graph.points = make([]fyne.CanvasObject, len(dataPoints)*2)
	for i := 0; i < len(dataPoints)-1; i++ {
		y := r.graph.transformValue(dataPoints[i].Y)
		var y1 float64
		if r.graph.config.IsLog {
			y1 = y
		} else {
			y1 = r.graph.transformValue(dataPoints[i].Y - dataPoints[i].ERR)
		}
		y2 := r.graph.transformValue(dataPoints[i].Y + dataPoints[i].ERR)
		x := margin + float32(dataPoints[i].X)*xScale
		yPos := size.Height - margin - float32(y-minData)*yScale
		yPos1 := size.Height - margin - float32(y1-minData)*yScale
		yPos2 := size.Height - margin - float32(y2-minData)*yScale

		point := canvas.NewCircle(color.Transparent)
		point.FillColor = color.Transparent
		point.StrokeWidth = 2
		point.StrokeColor = pointColor
		point.Position1 = fyne.NewPos(x-2, yPos-2)
		point.Position2 = fyne.NewPos(x+2, yPos+2)

		errorLine := canvas.NewLine(errorColor)
		errorLine.Position1 = fyne.NewPos(x, yPos1)
		errorLine.Position2 = fyne.NewPos(x, yPos2)
		errorLine.StrokeWidth = 2

		r.graph.points[i*2] = errorLine
		r.graph.points[i*2+1] = point
	}

	r.objects = []fyne.CanvasObject{r.graph.background}
	for _, line := range r.graph.gridLines {
		r.objects = append(r.objects, line)
	}
	r.objects = append(r.objects, r.graph.axes[0], r.graph.axes[1])

	for _, line := range r.graph.lines {
		r.objects = append(r.objects, line)
	}
	r.objects = slices.Concat(r.objects, r.graph.points)
	for _, label := range r.graph.xLabels {
		r.objects = append(r.objects, label)
	}
	for _, label := range r.graph.yLabels {
		r.objects = append(r.objects, label)
	}
	r.objects = append(r.objects, r.graph.title)
}

func (r *GraphRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *GraphRenderer) Destroy() {}

func (r *GraphRenderer) Refresh() {
	r.Layout(r.graph.Size())
	canvas.Refresh(r.graph)
}

func (g *GraphCanvas) UpdateData(newData *data.Function) {
	g.data = newData
	g.Refresh()
}
