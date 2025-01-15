package gui

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
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
	graphMap    = make(map[string]*graph.GraphCanvas)
)

// Start GUI (function is blocking)
func Start() {
	App = app.NewWithID("GUI-Physics")
	MainWindow = App.NewWindow("Physics GUI")
	GraphContainer = container.NewVBox()

	mainWindow()
}

func addDataset(reader io.ReadCloser, uri fyne.URI, err error) function.Points {
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return nil
	}
	if reader == nil {
		return nil // user canceled
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Println("error while closing dialog:", err)
		}
	}()

	// read file
	bytes, err := io.ReadAll(reader)
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return nil
	}

	// get filename
	filename := filepath.Base(uri.Name())

	// handle import
	points, err := data.Parse(bytes)
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return nil
	}

	if len(points) == 0 {
		dialog.ShowError(errors.New("no data"), MainWindow)
		return nil
	}

	// show success message
	dialog.ShowInformation("Import successful",
		fmt.Sprintf("File '%s' imported", filename),
		MainWindow)

	return points
}

func createImportButton(window fyne.Window) *widget.Button {
	return widget.NewButton("Import Data", func() {
		// open dialog
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			addDataset(reader, reader.URI(), err)
			// TODO: HELP
		}, window)
	})
}

// register functions which can be used for graph plotting
func registerFunctions() {
	functionMap["sld"] = function.NewEmptyFunction(function.INTERPOLATION_NONE)
	functionMap["eden"] = function.NewEmptyFunction(function.INTERPOLATION_NONE)
	functionMap["test"] = function.NewFunction(function.Points{}, function.INTERPOLATION_NONE)
}

// creates the graph containers for the different graphs
func registerGraphs() *fyne.Container {
	graphMap["sld"] = graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Intensity Graph",
		IsLog:     false,
		Functions: function.Functions{functionMap["sld"]},
	})

	graphMap["eden"] = graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Edensity Graph",
		IsLog:     false,
		Functions: function.Functions{functionMap["eden"]},
	})

	graphMap["test"] = graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Test Graph",
		IsLog:     true,
		Functions: function.Functions{functionMap["test"]},
	})

	return container.NewGridWithColumns(2, graphMap["eden"], graphMap["sld"], graphMap["test"])
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

func onDrop(position fyne.Position, uri []fyne.URI) {
	for mapIdentifier, u := range graphMap {
		if u.MouseInCanvas(position) {
			fmt.Println("Dropped on graph:", mapIdentifier)

			for _, v := range uri {
				rc, err := os.OpenFile(v.Path(), os.O_RDONLY, 0666)
				if err != nil {
					dialog.ShowError(err, MainWindow)
					return
				}

				if points := addDataset(rc, v, nil); points != nil {
					points.Magie()
					newFunction := function.NewFunction(points, function.INTERPOLATION_NONE)
					graphMap[mapIdentifier].AddDataTrack(newFunction)
				}
			}
			return
		}
	}
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

		container.NewVSplit(
			registerGraphs(),
			registerParams(),
		),
	)

	// set onchange function for recalculating data
	trigger.SetOnChange(RecalculateData)

	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)
	MainWindow.SetOnDropped(onDrop)

	MainWindow.ShowAndRun()
}

const (
	ELECTRON_RADIUS = 2.81e-5 // classical electron radius in angstrom
	ZNUMBER         = 150
	QZNUMBER        = 500
)

var qzAxis = physics.GetDefaultQZAxis(QZNUMBER)

func testFunc() {
	counter := 11

	d := make(function.Points, counter)

	for i := 0; i < counter; i++ {
		d[i] = &function.Point{
			X:     float64(i),
			Y:     math.Pow(float64(i), 2),
			Error: 1,
		}
	}

	functionMap["sld"].SetData(d)
	functionMap["eden"].SetData(d)
}

// RecalculateData recalculates the data for the sld and eden graphs
func RecalculateData() {
	// Get current parameter groups
	eden, err := param.GetFloats("eden")
	if err != nil {
		return
	}
	d, err := param.GetFloats("thick")
	if err != nil {
		return
	}
	sigma, err := param.GetFloats("rough")
	if err != nil {
		return
	}

	// get general parameters
	delta, err := param.GetFloat("general", "deltaq")
	if err != nil {
		return
	}
	background, err := param.GetFloat("general", "background")
	if err != nil {
		return
	}
	scaling, err := param.GetFloat("general", "scaling")
	if err != nil {
		return
	}

	// calculate edensity
	edenPoints, err := physics.GetEdensities(eden, d, sigma, ZNUMBER)
	if err != nil {
		fmt.Println("Error while calculating edensities:", err)
		return
	} else {
		functionMap["eden"].SetData(edenPoints)
	}

	// transform points into sld floats
	sld := make([]float64, len(edenPoints))
	for i, e := range edenPoints {
		sld[i] = e.Y * ELECTRON_RADIUS
	}

	deltaz := 0.0
	if edenPoints != nil && len(edenPoints) > 1 {
		deltaz = edenPoints[1].X - edenPoints[0].X
	}

	// calculate intensity
	modifiedQzAxis := make([]float64, len(qzAxis))
	copy(modifiedQzAxis, qzAxis)
	helper.Map(modifiedQzAxis, func(xPoint float64) float64 { return xPoint + delta })
	intensity := physics.CalculateIntensity(qzAxis, deltaz, sld, &physics.IntensityOptions{
		Background: background,
		Scaling:    scaling,
	})

	// creates list with intensity points based on edenPoints x and error and calculated intensity as y
	intensityPoints := make(function.Points, QZNUMBER)
	for i := range intensity {
		intensityPoints[i] = &function.Point{
			X:     qzAxis[i],
			Y:     intensity[i],
			Error: 0.0,
		}
		function.Magie(intensityPoints[i])
		//fmt.Println(*intensityPoints[i])

	}

	functionMap["sld"].SetData(intensityPoints)
}
