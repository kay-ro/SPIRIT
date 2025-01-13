package gui

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"path/filepath"
	"physicsGUI/pkg/data"
	"physicsGUI/pkg/function"
	"physicsGUI/pkg/gui/graph"
	"physicsGUI/pkg/gui/param"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	// App reference
	App            fyne.App
	MainWindow     fyne.Window
	GraphContainer *fyne.Container
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

// separator
func createSeparator() *canvas.Line {
	line := canvas.NewLine(color.Gray{Y: 100})
	line.StrokeWidth = 1
	return line
}

// AddMainWindow builds and renders the main GUI content, it will show and run the main window,
// which is a blocking command [fyne.Window.ShowAndRun]
func AddMainWindow() {
	importButton := createImportButton(MainWindow)

	toolbar := container.NewHBox(
		importButton,
	)

	separator := createSeparator()

	topContainer := container.NewVBox(
		toolbar,
		separator,
	)

	// create dataset 2^x
	dataset := make(function.Points, 21)
	dataset2 := make(function.Points, 9)

	for i := 0; i < len(dataset); i++ {
		dataset[i] = &function.Point{
			X:     float64(i),
			Y:     math.Pow(float64(i), 3),
			Error: 1,
		}
	}
	for i := 0; i < len(dataset2); i++ {
		dataset2[i] = &function.Point{
			X:     float64(i),
			Y:     math.Pow(float64(i), 2),
			Error: 1,
		}
	}

	g1 := graph.NewGraphCanvas(&graph.GraphConfig{
		Title: "Non Logarithmic x³ + x²",
		IsLog: false,
		Functions: []*function.Function{
			function.NewFunction(dataset, function.INTERPOLATION_NONE),
			function.NewFunction(dataset2, function.INTERPOLATION_NONE),
		},
	})

	g2 := graph.NewGraphCanvas(&graph.GraphConfig{
		Title: "Logarithmic x³",
		IsLog: true,
		Functions: []*function.Function{
			function.NewFunction(dataset, function.INTERPOLATION_NONE),
			function.NewFunction(dataset2, function.INTERPOLATION_NONE),
		},
	})

	GraphContainer.Add(g1)

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

	/* dummyGraph := graph.NewGraphCanvas(&graph.GraphConfig{
		Resolution: 100,
		Title:      "Dummy Graph to load data later",
		Functions: function.Functions{
			function.NewFunction(function.Points{{
				X:     0,
				Y:     0,
				Error: 0,
			}}, function.INTERPOLATION_NONE),
		},
	}) */

	//obj, p1 := param.Float("g1", "test", 1.123)
	obj2, _ := param.FloatMinMax("g1", "test2", 1.123)

	content := container.NewBorder(
		topContainer, // top
		nil,          // bottom
		nil,          // left
		nil,          // right

		container.NewVBox(
			container.NewGridWithColumns(2, sldGraph, g1 /* , dummyGraph */),
			container.NewVBox(obj2),
			g2,
		),
	)

	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)

	MainWindow.ShowAndRun()
}
