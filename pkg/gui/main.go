package gui

import (
	"fmt"
	"image/color"
	"io"
	"math"
	"math/rand"
	"path/filepath"
	"physicsGUI/pkg/data"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	App        fyne.App
	MainWindow fyne.Window
)

// Start GUI (function is blocking)
func Start() {
	App = app.New()
	MainWindow = App.NewWindow("Physics GUI")

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
			defer reader.Close()

			// read file
			bytes, err := io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			// get filename
			filename := filepath.Base(reader.URI().Path())

			// handle import
			if err := data.Import(bytes, filename); err != nil {
				dialog.ShowError(err, window)
				return
			}

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
	dataset := make([]data.Point, 10)
	for i := 0; i < 10; i++ {
		dataset[i] = data.Point{
			X:   float64(i),
			Y:   math.Pow(2, float64(i)),
			ERR: 1,
		}
	}

	graph1 := NewGraphCanvas(&GraphConfig{
		Title:    "Logarithmic",
		IsLog:    true,
		MinValue: 0.01,
		Data:     data.NewDataFunction(dataset, data.INTERPOLATION_NONE),
	})
	graph2 := NewGraphCanvas(&GraphConfig{
		Title: "Linear",
		Data:  data.NewDataFunction(dataset, data.INTERPOLATION_NONE),
	})
	graph3 := NewGraphCanvas(&GraphConfig{
		Title: "50ms Updates (bench)",
		Data:  data.NewDataFunction(dataset, data.INTERPOLATION_NONE),
	})

	//TODO remove Test Display
	plat1 := func(x float64) float64 { return 2 }
	grow1 := data.NewLogisticFunction(4, 2, 10, 4).GetF()
	plat2 := func(x float64) float64 { return 4 }
	grow2 := data.NewLogisticFunction(10, 4, 5, 1).GetF()
	graph4 := NewGraphCanvas(&GraphConfig{
		Resolution: 100,
		Title:      "SLD-Test",
		Data: data.NewSegmentedFunction([]data.FunctionSegment{
			data.NewFunctionSegment(0, 3, &plat1),
			data.NewFunctionSegment(3, 5, &grow1),
			data.NewFunctionSegment(5, 8, &plat2),
			data.NewFunctionSegment(8, 12, &grow2),
		}),
	})
	dummyFunction := data.NewOldSLDFunction(
		[]float64{0.0, 0.346197, 0.458849, 0.334000},
		[]float64{14.2657, 10.6906},
		[]float64{3.39544, 2.15980, 3.90204},
		150) // from refl_monolayer.pro:780
	if dummyFunction == nil {
		dummyFunction = data.NewDataFunction([]data.Point{{
			X:   0,
			Y:   0,
			ERR: 0,
		}}, data.INTERPOLATION_NONE)
	}
	testSLDGraph := NewGraphCanvas(&GraphConfig{
		Resolution: 100,
		Title:      "Test SLD",
		Data:       dummyFunction,
	})

	graphs := container.NewHSplit(
		graph1,
		graph2,
	)

	// "benchnmark"
	go func(graph *GraphCanvas) {
		for {
			newData := make([]data.Point, 10)
			for i := 0; i < 10; i++ {
				newData[i] = data.Point{
					X:   float64(i),
					Y:   float64(rand.Intn(150)),
					ERR: rand.Float64() * 20,
				}
			}

			graph.UpdateData(data.NewDataFunction(newData, data.INTERPOLATION_NONE))
			time.Sleep(50 * time.Millisecond)
		}
	}(graph3)

	content := container.NewBorder(
		topContainer, // top
		container.NewHSplit(testSLDGraph, graph4), // bottom
		nil, // left
		nil, // right

		container.NewVSplit(
			graphs,
			graph3,
		),
	)

	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)

	MainWindow.ShowAndRun()
}
