package graph

import (
	"image/color"
	"math"
	"physicsGUI/pkg/function"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// GraphCanvas represents the graphical representation of a graph.
type GraphCanvas struct {
	widget.BaseWidget
	function   *function.Function
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

// NewGraphCanvas creates a new canvas instance with a provided config.
// Specfically, it sets up the underlying structure of a canvas including lines, axes, labels and background.
// The method also calls 'ExtendBaseWidget' to cross-reference the canvas instance with the underlying fyne.BaseWidget struct.
func NewGraphCanvas(config *GraphConfig) *GraphCanvas {
	g := &GraphCanvas{
		lines:   make([]*canvas.Line, 0),
		axes:    make([]*canvas.Line, 2),
		xLabels: make([]*canvas.Text, 0),
		yLabels: make([]*canvas.Text, 0),

		config:     config,
		background: canvas.NewRectangle(color.Black),

		function: config.Function,
	}

	// needs to be to cross reference with the underlying struct
	g.ExtendBaseWidget(g)

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

// transform values to log10 and set to minvalue if applicable
func (g *GraphCanvas) transformValue(minValue, value float64) float64 {
	// if value is smaller than minValue, set it to minValue
	if value < minValue {
		value = minValue
	}

	if g.config.IsLog {
		return math.Log10(value)
	}

	return value
}

// CreateRenderer returns a [GraphRenderer] from a [GraphCanvas]
func (g *GraphCanvas) CreateRenderer() fyne.WidgetRenderer {
	return &GraphRenderer{
		graph:   g,
		objects: make([]fyne.CanvasObject, 0),
		size:    &fyne.Size{},
		margin:  float32(50),
	}
}

// UpdateFunction updates the function and refreshes the [GraphCanvas]
func (g *GraphCanvas) UpdateFunction(newFunction *function.Function) {
	g.function = newFunction
	g.Refresh()
}
