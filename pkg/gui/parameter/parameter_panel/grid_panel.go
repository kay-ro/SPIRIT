package parameter_panel

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ParameterGrid is a simple implementation for test display, that does not support Adding new Objects
type ParameterGrid struct {
	widget.BaseWidget
	objects []fyne.CanvasObject
}

func (p *ParameterGrid) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewAdaptiveGrid(5, p.objects...))
}

func NewParameterGrid(objects []fyne.CanvasObject) *ParameterGrid {
	g := &ParameterGrid{
		BaseWidget: widget.BaseWidget{},
		objects:    objects,
	}
	g.ExtendBaseWidget(g)
	return g
}
