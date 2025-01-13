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
	"physicsGUI/pkg/gui/parameter"
	"physicsGUI/pkg/gui/parameter/parameter_panel"

	"fyne.io/fyne/v2/data/binding"

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
	for i := 0; i < len(dataset); i++ {
		dataset[i] = &function.Point{
			X:     float64(i),
			Y:     math.Pow(float64(i), 3),
			Error: 1,
		}
	}

	g1 := graph.NewGraphCanvas(&graph.GraphConfig{
		Title:    "Non Logarithmic x³",
		IsLog:    false,
		Function: function.NewFunction(dataset, function.INTERPOLATION_NONE),
	})

	g2 := graph.NewGraphCanvas(&graph.GraphConfig{
		Title:    "Logarithmic x³",
		IsLog:    true,
		Function: function.NewFunction(dataset, function.INTERPOLATION_NONE),
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
		Function:   dummyFunction,
	})

	/* dummyGraph := graph.NewGraphCanvas(&graph.GraphConfig{
		Resolution: 100,
		Title:      "Dummy Graph to load data later",
		Function: function.NewFunction(function.Points{{
			X:     0,
			Y:     0,
			Error: 0,
		}}, function.INTERPOLATION_NONE),
	})
	GraphContainer.Add(dummyGraph) */

	parameterName := binding.NewString()
	err := parameterName.Set("Temporary Parameter")
	if err != nil {
		log.Println("error setting parameter name:", err)
	}

	defaultVal := binding.NewFloat()
	err = defaultVal.Set(10.04)
	if err != nil {
		log.Println("error setting default value:", err)
	}
	val := binding.NewFloat()
	minV := binding.NewFloat()
	err = minV.Set(-math.MaxFloat64)
	if err != nil {
		log.Println("error setting min value:", err)
	}
	maxV := binding.NewFloat()
	err = maxV.Set(math.MaxFloat64)
	if err != nil {
		log.Println("error setting max value:", err)
	}
	checkV := binding.NewBool()
	pparam := parameter.NewParameter(parameterName, defaultVal, val, minV, maxV, checkV)
	profilePanel := parameter_panel.NewParameterGrid(3)
	profilePanel.Add(pparam)

	/* p1 := param.NewString("placeholder")
	fmt.Println(p1) */

	/* profilePanel.OnValueChanged = func() {
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
	} */

	/* str := param.AddString("tv", "testvar", "Hi!")
	str2 := param.AddString("tv", "testvar2", "Hi!")

	str2.Set("okay")

	testvar := param.String("tv", "testvar3", "Hi!")
	testvar2 := param.String("tv", "testvar4", "Hi!")
	testvar3 := param.Float("tv", "testvar5", 1.25)
	testvar4 := param.Int("tv", "testvar6", 1) */

	/* testV1 := param.New(&param.Config[float64]{
		InitialValue: 1.110,
		Validator: func(s string) error {
			if _, err := strconv.ParseFloat(s, 64); err != nil {
				return fmt.Errorf("keine gültige Zahl")
			}

			return nil
		},
		Format: func(f float64) string {
			return fmt.Sprintf("%f", f)
		},
		Parser: func(f string) (float64, error) {
			return strconv.ParseFloat(f, 64)
		},
	})

	testV1.Set(1.123) */

	//obj, p1 := param.Float("g1", "test", 1.123)
	obj2, _ := param.FloatMinMax("g1", "test2", 1.123)

	content := container.NewBorder(
		topContainer, // top
		nil,          // bottom
		nil,          // left
		nil,          // right

		container.NewHSplit(
			container.NewVSplit(
				sldGraph,
				container.NewVBox(
					/* widget.NewLabelWithData(str),
					widget.NewEntryWithData(str),
					testvar,
					testvar2,
					testvar3,
					testvar4, */
					//obj,
					obj2,
				), /* profilePanel */
			),
			/* container.NewVScroll( */
			container.NewVSplit(
				g1, g2,
			),
			/* ), */
		),
	)

	/* go func() {
		for {
			v, err := p1.Get()
			if err != nil {
				log.Println("error getting value:", err)
			} else {
				p1.Set(v + 0.01)
			}

			time.Sleep(1 * time.Second)
		}
	}() */

	MainWindow.Resize(fyne.NewSize(1000, 500))
	MainWindow.SetContent(content)

	MainWindow.ShowAndRun()
}
