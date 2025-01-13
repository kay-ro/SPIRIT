package gui

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"path/filepath"
	"physicsGUI/pkg/data"
	"physicsGUI/pkg/function"
	"physicsGUI/pkg/gui/graph"
	"physicsGUI/pkg/gui/helper"
	"physicsGUI/pkg/gui/param"
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

	AddMainWindow()
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

// AddMainWindow builds and renders the main GUI content, it will show and run the main window,
// which is a blocking command [fyne.Window.ShowAndRun]
func AddMainWindow() {
	// create dataset 2^x
	dataset := make(function.Points, 21)

	for i := 0; i < len(dataset); i++ {
		dataset[i] = &function.Point{
			X:     float64(i),
			Y:     math.Pow(float64(i), 3),
			Error: 1,
		}
	}

	functionMap["test"] = function.NewEmptyFunction(function.INTERPOLATION_NONE)

	g1 := graph.NewGraphCanvas(&graph.GraphConfig{
		Title:     "Non Logarithmic x³ + x²",
		IsLog:     false,
		Functions: function.Functions{functionMap["test"]},
	})

	g2 := graph.NewGraphCanvas(&graph.GraphConfig{
		Title: "Logarithmic x³",
		IsLog: true,
		Functions: function.Functions{
			function.NewFunction(dataset, function.INTERPOLATION_NONE),
		},
	})

	// from refl_monolayer.pro:780
	dummyFunction := data.NewOldSLDFunction(
		[]float64{0.0, 0.346197, 0.458849, 0.334000},
		[]float64{14.2657, 10.6906},
		[]float64{3.39544, 2.15980, 3.90204},
		150)

	dummyFunction.SetInterpolation(function.INTERPOLATION_LINEAR)

	sldGraph := graph.NewGraphCanvas(&graph.GraphConfig{
		Resolution: 5,
		Title:      "Electron Density",
		Functions:  function.Functions{dummyFunction},
	})

	obj2, _ := param.Int("g1", "TEST_VAR", 1)

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
			container.NewGridWithColumns(2, sldGraph, g1, g2 /* , dummyGraph */),
			container.NewVBox(obj2),
		),
	)

	// set onchange function for recalculating data
	trigger.SetOnChange(RecalculateData)

	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)

	MainWindow.ShowAndRun()
}

func RecalculateData() {
	counter, err := param.GetInt("g1", "TEST_VAR")
	if err != nil {
		log.Println(err)
		return
	}

	d := make(function.Points, counter)

	for i := 0; i < counter; i++ {
		d[i] = &function.Point{
			X:     float64(i),
			Y:     math.Pow(float64(i), 2),
			Error: 1,
		}
	}

	functionMap["test"].SetData(d)
}
