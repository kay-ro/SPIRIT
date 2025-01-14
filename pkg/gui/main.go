package gui

import (
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"physicsGUI/pkg/data"
	"physicsGUI/pkg/function"
	"physicsGUI/pkg/gui/graph"
	"physicsGUI/pkg/gui/helper"
	"physicsGUI/pkg/gui/param"
	"physicsGUI/pkg/physics"
	"physicsGUI/pkg/trigger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	// App reference
	App            fyne.App
	MainWindow     fyne.Window
	GraphContainer *fyne.Container

	functionMap = make(map[string]*function.Function)
)

// Start GUI (function is blocking)
func Start() {
	App = app.NewWithID("GUI-Physics")
	MainWindow = App.NewWindow("Physics GUI")
	GraphContainer = container.NewVBox()

	mainWindow()
}

func createImportButton(window fyne.Window) *widget.Button {
	return widget.NewButton("Import Data", func() {

		// open dialog
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if reader == nil {
				return // user canceled
			}
			defer func() {
				if err := reader.Close(); err != nil {
					log.Println("error while closing dialog:", err)
				}
			}()

			// read file
			bytes, err := io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			// get filename
			filename := filepath.Base(reader.URI().Path())

			// handle import
			points, err := data.Parse(bytes)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			if len(points) == 0 {
				dialog.ShowError(errors.New("no data"), window)
				return
			}

			GraphContainer.RemoveAll()

			plot := graph.NewGraphCanvas(&graph.GraphConfig{
				Title:      fmt.Sprintf("Data track %d", 1),
				IsLog:      false,
				Resolution: 200,
				Functions: function.Functions{
					function.NewFunction(points, function.INTERPOLATION_NONE),
				},
			})

			GraphContainer.Add(plot)

			// show success message
			dialog.ShowInformation("Import successful",
				fmt.Sprintf("File '%s' imported", filename),
				window)
		}, window)
	})
}

// register functions which can be used for graph plotting
func registerFunctions() {
	functionMap["sld"] = function.NewEmptyFunction(function.INTERPOLATION_NONE)
	functionMap["eden"] = function.NewEmptyFunction(function.INTERPOLATION_NONE)
}

// creates the graph containers for the different graphs
func registerGraphs() *fyne.Container {
	sld := graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "SLD Graph",
		IsLog:     false,
		Functions: function.Functions{functionMap["sld"]},
	})

	eden := graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Logarithmic x³",
		IsLog:     true,
		Functions: function.Functions{functionMap["eden"]},
	})

	return container.NewGridWithColumns(2, sld, eden)
}

// creates and registers the parameter and adds them to the parameter repository
func registerParams() *fyne.Container {
	edenA, _ := param.FloatMinMax("eden", "Eden a", 0.0)
	eden1, _ := param.FloatMinMax("eden", "Eden 1", 0.346197)
	eden2, _ := param.FloatMinMax("eden", "Eden 2", 0.458849)
	edenB, _ := param.FloatMinMax("eden", "Eden b", 0.334000)

	roughnessA1, _ := param.FloatMinMax("rough", "Roughness a/1", 3.39544)
	roughness12, _ := param.FloatMinMax("rough", "Roughness 1/2", 2.15980)
	roughness2B, _ := param.FloatMinMax("rough", "Roughness 2/b", 3.90204)

	thickness1, _ := param.FloatMinMax("thick", "Thickness 1", 14.2657)
	thickness2, _ := param.FloatMinMax("thick", "Thickness 2", 10.6906)

	deltaQ, _ := param.Float("general", "deltaq", 0.0)
	background, _ := param.Float("general", "background", 10e-9)
	scaling, _ := param.Float("general", "scaling", 1.0)

	return container.NewVBox(
		container.NewGridWithColumns(4, edenA, eden1, eden2, edenB),
		container.NewGridWithColumns(4, roughnessA1, roughness12, roughness2B),
		container.NewGridWithColumns(4, thickness1, thickness2),
		container.NewGridWithColumns(4, deltaQ, background, scaling),
	)
}

// mainWindow builds and renders the main GUI content, it will show and run the main window,
// which is a blocking command [fyne.Window.ShowAndRun]
func mainWindow() {
	registerFunctions()

	content := container.NewBorder(
		container.NewVBox(
			container.NewHBox(
				createImportButton(MainWindow),
			),
			helper.CreateSeparator(),
		), // top
		nil, // bottom
		nil, // left
		nil, // right

		container.NewVBox(
			registerGraphs(),
			registerParams(),
		),
	)

	// set onchange function for recalculating data
	trigger.SetOnChange(RecalculateData)

	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)

	MainWindow.ShowAndRun()
}

const (
	ELECTRON_RADIUS = 2.81e-5 // classical electron radius in angstrom
	zNumber         = 150
)

// RecalculateData recalculates the data for the sld and eden graphs
func RecalculateData() {
	// Get current parameter groups
	eden, _ := param.GetFloats("eden")
	d, _ := param.GetFloats("thick")
	sigma, _ := param.GetFloats("rough")

	// get general parameters
	delta, _ := param.GetFloat("general", "deltaq")
	background, _ := param.GetFloat("general", "background")
	scaling, _ := param.GetFloat("general", "scaling")

	// calculate edensity
	edenPoints, err := physics.GetEdensities(eden, d, sigma, zNumber)
	if err == nil {
		functionMap["eden"].SetData(edenPoints)
	}

	// calculate zaxis
	zaxis := physics.GetZAxis(d, zNumber)

	// transform points into sld floats
	sld := make([]float64, len(edenPoints))
	for i, e := range edenPoints {
		sld[i] = e.Y * ELECTRON_RADIUS
	}

	intensity := physics.CalculateIntensity(zaxis, delta, sld, &physics.IntensityOptions{
		Background: background,
		Scaling:    scaling,
	})

	// reusing the edenPoints struct and updating the Y values
	for i := range intensity {
		edenPoints[i].Y = intensity[i]
	}

	functionMap["sld"].SetData(edenPoints)
}
