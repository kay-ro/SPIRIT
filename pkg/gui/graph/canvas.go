package graph

import (
	"image/color"
	"physicsGUI/pkg/function"
	"slices"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// GraphCanvas represents the graphical representation of a graph.
type GraphCanvas struct {
	widget.BaseWidget
	config     *GraphConfig
	background *canvas.Rectangle

	functions         function.Functions
	loadedData        function.Functions
	dataRemoveButtons []*fyne.Container
}

// NewGraphCanvas creates a new canvas instance with a provided config.
// Specfically, it sets up the underlying structure of a canvas including lines, axes, labels and background.
// The method also calls 'ExtendBaseWidget' to cross-reference the canvas instance with the underlying fyne.BaseWidget struct.
func NewGraphCanvas(config *GraphConfig) *GraphCanvas {
	g := &GraphCanvas{
		config:     config,
		background: canvas.NewRectangle(color.Black),

		functions:  config.Functions,
		loadedData: make(function.Functions, 0),
	}

	for _, f := range g.functions {
		if f == nil {
			panic("function cannot be nil. Make sure to provide a function (even an empty one)")
		}
	}

	// needs to be to cross reference with the underlying struct
	g.ExtendBaseWidget(g)

	return g
}

func (g *GraphCanvas) MouseInCanvas(position fyne.Position) bool {
	pos := g.BaseWidget.Position()

	return position.X >= pos.X && position.X <= pos.X+g.Size().Width && position.Y >= pos.Y && position.Y <= pos.Y+g.Size().Height
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

// UpdateFunctions updates the function and refreshes the [GraphCanvas]
func (g *GraphCanvas) UpdateFunctions(newFunctions function.Functions) {
	g.functions = newFunctions
	g.Refresh()
}

func (g *GraphCanvas) AddDataTrack(dataTrack *function.Function) {
	i := len(g.loadedData)
	g.loadedData = append(g.loadedData, dataTrack)

	// create remove button
	btnRemove := widget.NewButton("🗑", func() {
		g.RemoveDataTrack(dataTrack)
	})
	btnRemove.Resize(fyne.NewSize(20, 20))
	btnColor := DataTrackColors[i%len(DataTrackColors)]
	g.dataRemoveButtons = append(g.dataRemoveButtons, container.NewStack(canvas.NewRectangle(btnColor), btnRemove))

	g.Refresh()
}

func (g *GraphCanvas) GetDataTracks() function.Functions {
	return g.loadedData
}
func (g *GraphCanvas) GetFunctions() function.Functions {
	return g.functions
}

func (g *GraphCanvas) RemoveDataTrack(dataTrack *function.Function) {
	i := slices.Index(g.loadedData, dataTrack)
	if i != -1 {
		g.loadedData = append(g.loadedData[:i], g.loadedData[i+1:]...)
		g.dataRemoveButtons = append(g.dataRemoveButtons[:i], g.dataRemoveButtons[i+1:]...)
		g.Refresh()
	}
}

func (g *GraphCanvas) AddFunction(newFunction *function.Function) {
	g.functions = append(g.functions, newFunction)
	g.Refresh()
}
