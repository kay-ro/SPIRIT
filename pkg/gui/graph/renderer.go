package graph

import (
	"image/color"
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
type GraphRange struct {
	Min float64
	Max float64
}

// returns the minimum size needed for the graph
func (r *GraphRenderer) MinSize() fyne.Size {
	return fyne.NewSize(500, 200)
}

// initializes the base strcuture for every graph
func (r *GraphRenderer) base() {
	// background
	r.graph.background.Resize(*r.size)
	r.graph.background.Move(fyne.NewPos(0, 0))
	r.AddObject(r.graph.background)

	// title
	title := &canvas.Text{
		Text:      r.graph.config.Title,
		Color:     titleColor,
		TextSize:  16,
		TextStyle: fyne.TextStyle{Bold: true},
	}
	title.Move(fyne.NewPos(r.size.Width/2-float32(len(title.Text)*4), 0))
	r.AddObject(title)

	// x-axis
	r.AddObject(&canvas.Line{
		StrokeColor: axesColor,
		StrokeWidth: 1,
		Position1:   fyne.NewPos(r.margin-2, r.size.Height-r.margin+2),
		Position2:   fyne.NewPos(r.size.Width-r.margin/2, r.size.Height-r.margin+2),
	})

	// y-axis
	r.AddObject(&canvas.Line{
		StrokeColor: axesColor,
		StrokeWidth: 1,
		Position1:   fyne.NewPos(r.margin-2, r.size.Height-r.margin+2),
		Position2:   fyne.NewPos(r.margin-2, 0.5*r.margin),
	})
}

// draws the whole graph
func (r *GraphRenderer) Layout(size fyne.Size) {
	// clear objects
	r.objects = make([]fyne.CanvasObject, 0)

	// size of the graph
	r.size = &size

	// About layout when size is zero and therefor component is not visible
	if r.size.Width < r.MinSize().Width || r.size.Height < r.MinSize().Height {
		return
	}

	// set the base for the canvas
	r.base()

	if r.graph.config.DisplayRange != nil {
		for _, f := range r.graph.config.Functions {
			f.Range(r.graph.config.DisplayRange.Min, r.graph.config.DisplayRange.Max)
		}
		for _, d := range r.graph.loadedData {
			d.Range(r.graph.config.DisplayRange.Min, r.graph.config.DisplayRange.Max)
		}
	}
	// calculate the maximum scope
	scope := function.GetMaximumScope(append(r.graph.functions, r.graph.loadedData...)...)

	if scope == nil {
		r.DrawErrorMessage("Scope error")
		return
	}

	if (len(r.graph.functions) == 0 || r.graph.functions[0].GetDataCount() < 1) && len(r.graph.loadedData) == 0 {
		r.DrawErrorMessage("No data available")
		return
	}

	// Add Remove Buttons
	r.DrawRemoveButtons()

	// draw model lines
	if r.graph.config.IsLog {
		for _, f := range r.graph.functions {
			points, iPoints := f.Model(r.graph.config.Resolution, true)
			r.DrawGraphLog(scope, points, iPoints, pointColor, false)
		}
		for i, d := range r.graph.loadedData {
			d.Range(r.graph.config.DisplayRange.Min, r.graph.config.DisplayRange.Max)
			points, iPoints := d.Model(r.graph.config.Resolution, true)
			dataColor := DataTrackColors[i%len(DataTrackColors)]
			r.DrawGraphLog(scope, points, iPoints, dataColor, true)
		}
		r.DrawGridLog(scope)
		return
	}

	for _, f := range r.graph.functions {
		points, iPoints := f.Model(r.graph.config.Resolution, false)
		r.DrawGraphLinear(scope,
			points.Filter(r.graph.config.DisplayRange.Min, r.graph.config.DisplayRange.Max),
			iPoints.Filter(r.graph.config.DisplayRange.Min, r.graph.config.DisplayRange.Max),
			pointColor, false)
	}
	for i, d := range r.graph.loadedData {
		points, iPoints := d.Model(r.graph.config.Resolution, false)
		dataColor := DataTrackColors[i%len(DataTrackColors)]
		r.DrawGraphLinear(scope,
			points.Filter(r.graph.config.DisplayRange.Min, r.graph.config.DisplayRange.Max),
			iPoints.Filter(r.graph.config.DisplayRange.Min, r.graph.config.DisplayRange.Max),
			dataColor, true)
	}
	r.DrawGridLinear(scope)
}

// display remove buttons at the right border
func (r *GraphRenderer) DrawRemoveButtons() {
	offsetY := float32(0)
	startY := float32(0)
	startX := r.size.Width - r.margin
	for _, d := range r.graph.dataRemoveButtons {
		offsetY += d.Size().Height + RemoveButtonTopPadding
		for _, o := range d.Objects {
			o.Move(fyne.NewPos(startX, startY+offsetY))
			r.AddObject(o)
		}
	}
}

// draw an error message onto the graph
func (r *GraphRenderer) DrawErrorMessage(message string) {
	errorMsg := &canvas.Text{
		Text:     message,
		Color:    titleColor,
		TextSize: 16,
	}
	errorMsg.Move(fyne.NewPos(
		r.size.Width/2-float32(len(errorMsg.Text)*4),
		r.size.Height/2-errorMsg.Size().Height/2-r.margin/2))

	r.AddObject(errorMsg)
}

// normalizes the coodinates from the bottom left of the canvas
func (r *GraphRenderer) normalize(x float32, y float32) (float32, float32) {
	return x + r.margin, r.size.Height - r.margin - y
}

// TODO: fix the small gap between points and lines
// TODO: if points are the same size as lines
// TODO: -> points are on the bottom of the lines

// draw a grid point
func (r *GraphRenderer) DrawPoint(x float32, y float32, pointColor color.Color) {
	r.AddObject(&canvas.Circle{
		FillColor: pointColor,
		Position1: fyne.NewPos(x-pointRadius, y-pointRadius),
		Position2: fyne.NewPos(x+pointRadius, y+pointRadius),
	})
}

// draw error correction lines within bounds of graph
func (r *GraphRenderer) DrawError(x, y1, y2 float32, errorColor color.Color) {
	r.AddObject(&canvas.Line{
		StrokeColor: errorColor,
		StrokeWidth: 1,
		Position1:   fyne.NewPos(x, min(max(y1, 0), r.size.Height)),
		Position2:   fyne.NewPos(x, min(max(y2, 0), r.size.Height)),
	})
}

// helper function for the grid lines
func (r *GraphRenderer) DrawGridLine(pos fyne.Position, isVertical, isMinor bool) {
	line := &canvas.Line{
		StrokeColor: gridColor,
		StrokeWidth: 1,
		Position1:   pos,
	}

	if isVertical {
		line.Position2 = fyne.NewPos(pos.X, pos.Y+r.size.Height-1.5*r.margin)
	} else {
		line.Position2 = fyne.NewPos(pos.X+r.size.Width-1.5*r.margin, pos.Y)
	}

	if isMinor {
		line.StrokeColor = gridMinorColor
		line.StrokeWidth = 0.5
	}

	r.AddObject(line)
}

// returns the objects of the graph
func (r *GraphRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// destroy function (needs to be here to satisfy the interface)
func (r *GraphRenderer) Destroy() {}

// refresh function (needs to be here to satisfy the interface)
func (r *GraphRenderer) Refresh() {}

// add an object to the graph renderer
func (r *GraphRenderer) AddObject(object fyne.CanvasObject) {
	r.objects = append(r.objects, object)
}
