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
			measurements, err := data.Parse(bytes)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			if len(measurements) == 0 {
				dialog.ShowError(errors.New("no data"), window)
				return
			}

			points := make(function.Points, len(measurements))
			for i, m := range measurements {
				points[i] = m.ToPoint()
			}

			// convert to Point format
			/* points := make([][]function.Point, measurements[0].Count)
			for j, m := range measurements {
				for i := 0; i < measurements[j].Count; i++ {
					if j == 0 {
						points[i] = make([]function.Point, len(measurements))
					}
					points[i][j] = function.Point{
						X:     m.Time,
						Y:     m.Data[i],
						Error: m.Error,
					}
				}
			} */

			GraphContainer.RemoveAll()
			//minP, _ := plotFunc.Scope()
			plot := graph.NewGraphCanvas(&graph.GraphConfig{
				Title: fmt.Sprintf("Data track %d", 1),
				IsLog: false,
				//MinValue:   minP.X,
				Resolution: 200,
				Function:   function.NewFunction(points, function.INTERPOLATION_NONE),
			})

			GraphContainer.Add(plot)
			// Clear old plots and add new
			/* GraphContainer.RemoveAll()
			for i := 0; i < len(points); i++ {
				plotFunc := function.NewDataFunction(points, function.INTERPOLATION_NONE)
				minP, _ := plotFunc.Scope()
				plot := NewGraphCanvas(&GraphConfig{
					Title:      fmt.Sprintf("Data track %d", i+1),
					IsLog:      false,
					MinValue:   minP.X,
					Resolution: 200,
					Function:   plotFunc,
				})

				GraphContainer.Add(plot)
			}
			GraphContainer.Refresh() */

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
	dataset := make(function.Points, 10)
	for i := 0; i < 10; i++ {
		dataset[i] = &function.Point{
			X:     float64(i) + 0.5,
			Y:     math.Pow(2, float64(i)),
			Error: 1,
		}
	}

	g1 := graph.NewGraphCanvas(&graph.GraphConfig{
		Title:    "Logarithmic",
		IsLog:    true,
		Function: function.NewFunction(dataset, function.INTERPOLATION_NONE),
	})

	GraphContainer.Add(g1)

	dummyFunction := data.NewOldSLDFunction(
		[]float64{0.0, 0.346197, 0.458849, 0.334000},
		[]float64{14.2657, 10.6906},
		[]float64{3.39544, 2.15980, 3.90204},
		150) // from refl_monolayer.pro:780
	/* if dummyFunction == nil {
		dummyFunction = function.NewDataFunction(function.Points{{
			X:     0,
			Y:     0,
			Error: 0,
		}}, function.INTERPOLATION_NONE)
	} */
	sldGraph := graph.NewGraphCanvas(&graph.GraphConfig{
		Resolution: 100,
		Title:      "Electron Density ",
		Function:   dummyFunction,
	})

	dummyGraph := graph.NewGraphCanvas(&graph.GraphConfig{
		Resolution: 100,
		Title:      "Dummy Graph to load data later",
		Function: function.NewFunction(function.Points{{
			X:     0,
			Y:     0,
			Error: 0,
		}}, function.INTERPOLATION_NONE),
	})
	GraphContainer.Add(dummyGraph)

	profilePanel := NewProfilePanel(NewSldDefaultSettings("Settings"))
	profilePanel.OnValueChanged = func() {
		edensity := make([]float64, len(profilePanel.Profiles)+2)
		sigma := make([]float64, len(profilePanel.Profiles)+1)
		d := make([]float64, len(profilePanel.Profiles))

		var err error = nil
		edensity[0], err = profilePanel.base.Parameter[ProfileDefaultEdensityID].GetValue()
		sigma[0], err = profilePanel.base.Parameter[ProfileDefaultRoughnessID].GetValue()
		edensity[len(profilePanel.Profiles)+1], err = profilePanel.bulk.Parameter[ProfileDefaultEdensityID].GetValue()
		for i, profile := range profilePanel.Profiles {
			edensity[i+1], err = profile.Parameter[ProfileDefaultEdensityID].GetValue()
			sigma[i+1], err = profile.Parameter[ProfileDefaultRoughnessID].GetValue()
			d[i], err = profile.Parameter[ProfileDefaultThicknessID].GetValue()
		}
		var zNumberF float64 = 100.0
		zNumberF, err = profilePanel.sldSettings.Parameter[SldDefaultZNumberID].GetValue()
		zNumber := int(zNumberF)

		if err != nil {
			println(errors.Join(errors.New("error while reading default parameters"), err).Error())
		}

		newEdensity := data.NewOldSLDFunction(edensity, d, sigma, zNumber)
		if newEdensity == nil {
			println(errors.New("no old getEden function implemented for this parameter count").Error())
			return
		}
		sldGraph.UpdateFunction(newEdensity)
	}

	content := container.NewBorder(
		topContainer, // top
		nil,          // bottom
		nil,          // left
		nil,          // right

		container.NewHSplit(
			container.NewVSplit(
				sldGraph,
				profilePanel,
			),
			container.NewVScroll(g1 /* GraphContainer */),
		),
	)

	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)

	MainWindow.ShowAndRun()
}
